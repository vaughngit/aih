//go:build !windows

package runtime

import (
	"os"
	"os/signal"
	"syscall"
)

// forwardSignals relays terminal/tty signals to the child until cmd.Wait
// returns. Returns a stop func that the caller defers.
func forwardSignals(p *os.Process) func() {
	ch := make(chan os.Signal, 4)
	signal.Notify(ch,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case sig := <-ch:
				_ = p.Signal(sig)
			case <-done:
				return
			}
		}
	}()
	return func() {
		signal.Stop(ch)
		close(done)
	}
}

// signalName extracts the signal name from an exec.ExitError if the child
// died from a signal rather than exiting with a status.
func signalName(e interface{ Sys() any }) (string, bool) {
	ws, ok := e.Sys().(syscall.WaitStatus)
	if !ok || !ws.Signaled() {
		return "", false
	}
	return ws.Signal().String(), true
}
