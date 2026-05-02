package backends

import (
	"fmt"
	"sort"
	"strings"
)

// keyType describes the expected scalar type of a config key. Tables /
// arrays aren't currently used by any backend; add them when needed.
type keyType int

const (
	keyString keyType = iota
	keyBool
)

func (k keyType) String() string {
	switch k {
	case keyString:
		return "string"
	case keyBool:
		return "bool"
	}
	return "unknown"
}

// checkAllowedKeys verifies every key in raw is declared in allowed and has
// the declared scalar type. Unknown keys produce an error listing the legal
// set so users can spot typos.
func checkAllowedKeys(raw map[string]any, allowed map[string]keyType, backendName string) error {
	if raw == nil {
		return nil
	}
	for k, v := range raw {
		want, ok := allowed[k]
		if !ok {
			return fmt.Errorf("%s: unknown config key %q (allowed: %s)",
				backendName, k, strings.Join(allowedKeyNames(allowed), ", "))
		}
		switch want {
		case keyString:
			if _, ok := v.(string); !ok {
				return fmt.Errorf("%s: config key %q must be a string, got %T", backendName, k, v)
			}
		case keyBool:
			if _, ok := v.(bool); !ok {
				return fmt.Errorf("%s: config key %q must be a bool, got %T", backendName, k, v)
			}
		}
	}
	return nil
}

func allowedKeyNames(allowed map[string]keyType) []string {
	names := make([]string, 0, len(allowed))
	for k := range allowed {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// checkEnum verifies a string field is one of allowed. Empty value is OK
// (treat as "not set"); call with a non-empty value to enforce.
func checkEnum(value, fieldName, backendName string, allowed []string) error {
	if value == "" {
		return nil
	}
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return fmt.Errorf("%s: %s %q must be one of: %s",
		backendName, fieldName, value, strings.Join(allowed, ", "))
}

// stringOrDefault reads a string-typed key from a backend config table.
// Returns fallback if the key is absent or the table is nil. Type errors
// are caught earlier by checkAllowedKeys, so this just falls back.
func stringOrDefault(cfg map[string]any, key, fallback string) string {
	if cfg == nil {
		return fallback
	}
	v, ok := cfg[key]
	if !ok {
		return fallback
	}
	s, ok := v.(string)
	if !ok {
		return fallback
	}
	return s
}

// boolOrDefault reads a bool-typed key from a backend config table.
func boolOrDefault(cfg map[string]any, key string, fallback bool) bool {
	if cfg == nil {
		return fallback
	}
	v, ok := cfg[key]
	if !ok {
		return fallback
	}
	b, ok := v.(bool)
	if !ok {
		return fallback
	}
	return b
}
