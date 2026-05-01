# aih — agentic-ai harness

A small Go CLI that launches any agent backend (Claude Code, Codex, Crush, Kiro, OpenCode, generic) from anywhere on the system with the right workspace, env, resources, and hooks. Manifests are TOML — central registry at `~/.aih/agents/` with per-project overrides.

> **Status**: Phase 0 (bootstrap). Not usable yet. See [implementation plan](https://github.com/vaughngit/aih/blob/main/docs/plan.md) for phase progression.

## Install

```sh
go install github.com/vaughngit/aih/cmd/aih@latest
```

Requires Go 1.25+.

Pre-built binaries and Homebrew tap will land at v0.1.0 (Phase 6).

## Quickstart

*(coming in Phase 1 — `aih launch <name>` end-to-end)*

```toml
# ~/.aih/agents/k3s-infra.toml
name = "k3s-infra"
backend = "claude-code"
workspace = "~/dev/kubernetes"

resources = [
  "~/dev/second-brain/projects/k3s-cluster",
]

[env]
KUBECONFIG = "~/.kube/config-k3s-pi"

[backend.claude-code]
agent = "kubernetes"

[hooks]
pre_launch = ["git -C ~/dev/kubernetes pull --ff-only"]
```

```sh
aih launch k3s-infra            # cd's into workspace, runs configured backend
aih list                        # all manifests in central registry
aih show k3s-infra              # resolved manifest with paths expanded
```

## Design

Locked decisions (manifest format, backend plugins, project-local trust model, hook execution semantics) live in the design doc. Implementation phases live in the plan.

- **Spec**: `docs/design.md` (mirrored from `~/dev/second-brain/research/ai_agents/agentai-harness-design.md`)
- **Plan**: `docs/plan.md` (mirrored from `~/dev/second-brain/projects/aih/plan.md`)

## License

Apache 2.0 — see [LICENSE](LICENSE).
