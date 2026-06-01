package launchctl

// KeepAlive mode discriminators for KeepAliveConfig.Mode. They record which
// plist form the value round-trips to (a launchd KeepAlive may be a bool or a
// dictionary of conditions).
const (
	KeepAliveModeBoolean    = "boolean"
	KeepAliveModeDictionary = "dictionary"
)

// KeepAliveConfig models launchd's KeepAlive key, which may be either a plain
// boolean or a dictionary of conditions. Mode records the original plist form
// so it can be round-tripped without loss. The pointer bool sub-keys
// distinguish "unset" from an explicit true/false; the map sub-keys carry
// PathState / OtherJobEnabled entries verbatim for fidelity even though the UI
// does not expose them for editing.
type KeepAliveConfig struct {
	Enabled            bool            `json:"enabled"`
	Mode               string          `json:"mode"` // "boolean" | "dictionary"
	SuccessfulExit     *bool           `json:"successfulExit,omitempty"`
	Crashed            *bool           `json:"crashed,omitempty"`
	AfterInitialDemand *bool           `json:"afterInitialDemand,omitempty"`
	NetworkState       *bool           `json:"networkState,omitempty"`
	PathState          map[string]bool `json:"pathState,omitempty"`
	OtherJobEnabled    map[string]bool `json:"otherJobEnabled,omitempty"`
}

// parseKeepAlive converts a plist KeepAlive value (nil, bool, or dictionary)
// into a structured KeepAliveConfig:
//
//   - nil / key absent → {Enabled: false} (no mode)
//   - bool             → {Enabled: <value>, Mode: "boolean"}
//   - map[string]any   → {Enabled: true, Mode: "dictionary", ...sub-keys}
//
// Unrecognized top-level types and type-mismatched inner values are ignored
// rather than treated as errors, preserving the existing "skip what we cannot
// parse" robustness of the surrounding service-loading code.
func parseKeepAlive(v any) KeepAliveConfig {
	switch v := v.(type) {
	case bool:
		return KeepAliveConfig{Enabled: v, Mode: KeepAliveModeBoolean}
	case map[string]any:
		return KeepAliveConfig{
			Enabled:            true,
			Mode:               KeepAliveModeDictionary,
			SuccessfulExit:     keepAliveBool(v, "SuccessfulExit"),
			Crashed:            keepAliveBool(v, "Crashed"),
			AfterInitialDemand: keepAliveBool(v, "AfterInitialDemand"),
			NetworkState:       keepAliveBool(v, "NetworkState"),
			PathState:          keepAliveBoolMap(v, "PathState"),
			OtherJobEnabled:    keepAliveBoolMap(v, "OtherJobEnabled"),
		}
	}
	return KeepAliveConfig{}
}

// keepAliveBool returns a pointer to the bool at key, or nil when the key is
// absent or not a bool.
func keepAliveBool(m map[string]any, key string) *bool {
	if b, ok := m[key].(bool); ok {
		return &b
	}
	return nil
}

// keepAliveBoolMap extracts a map[string]bool from a nested plist dictionary,
// dropping any entry whose value is not a bool. Returns nil (rather than an
// empty map) when the key is absent, not a dictionary, or has no bool entries,
// so the encode side can treat "no entries" uniformly.
func keepAliveBoolMap(m map[string]any, key string) map[string]bool {
	raw, ok := m[key].(map[string]any)
	if !ok || len(raw) == 0 {
		return nil
	}
	out := make(map[string]bool, len(raw))
	for k, val := range raw {
		if b, ok := val.(bool); ok {
			out[k] = b
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// buildKeepAlive renders a KeepAliveConfig into the plist value launchd
// expects, returning ok=false when the KeepAlive key should be omitted
// entirely. The rules are the single source of truth for KeepAlive encoding:
//
//   - disabled                  → omitted (ok=false)
//   - boolean mode              → bool true
//   - dictionary with sub-keys  → map of the set sub-keys
//   - dictionary with no sub-key → bool true (never an empty dictionary)
//
// An empty dictionary is semantically equivalent to true but reads as if it
// carried conditions, so it is downgraded to the boolean form.
func buildKeepAlive(c KeepAliveConfig) (any, bool) {
	if !c.Enabled {
		return nil, false
	}
	if c.Mode != KeepAliveModeDictionary {
		return true, true
	}

	dict := map[string]any{}
	if c.SuccessfulExit != nil {
		dict["SuccessfulExit"] = *c.SuccessfulExit
	}
	if c.Crashed != nil {
		dict["Crashed"] = *c.Crashed
	}
	if c.AfterInitialDemand != nil {
		dict["AfterInitialDemand"] = *c.AfterInitialDemand
	}
	if c.NetworkState != nil {
		dict["NetworkState"] = *c.NetworkState
	}
	if m := boolMapToAny(c.PathState); m != nil {
		dict["PathState"] = m
	}
	if m := boolMapToAny(c.OtherJobEnabled); m != nil {
		dict["OtherJobEnabled"] = m
	}

	if len(dict) == 0 {
		return true, true
	}
	return dict, true
}

// boolMapToAny converts a map[string]bool into the map[string]any shape the
// plist encoder expects, returning nil for an empty or nil input.
func boolMapToAny(m map[string]bool) map[string]any {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
