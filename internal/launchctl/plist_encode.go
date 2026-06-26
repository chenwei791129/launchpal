package launchctl

import (
	"fmt"
	"maps"
	"os"

	"howett.net/plist"
)

// modeledPlistKeys is the single source of truth for every plist key
// BuildPlistDict can emit. The Update merge logic (mergeUnmodeledKeys) uses it
// as the removal set so that clearing or toggling a modeled key strips its
// stale on-disk value, while keys absent from this set (everything LaunchPal
// does not model) are preserved verbatim. StartInterval and
// StartCalendarInterval are mutually exclusive within a single config but both
// belong here so switching schedule kinds never leaves a stale key behind.
// TestModeledPlistKeys enforces that this set covers every key BuildPlistDict
// produces, preventing drift when new modeled keys are added.
var modeledPlistKeys = map[string]bool{
	"Label":                 true,
	"Program":               true,
	"ProgramArguments":      true,
	"RunAtLoad":             true,
	"KeepAlive":             true,
	"ThrottleInterval":      true,
	"WorkingDirectory":      true,
	"StandardOutPath":       true,
	"StandardErrorPath":     true,
	"WakeSystem":            true,
	"EnvironmentVariables":  true,
	"StartInterval":         true,
	"StartCalendarInterval": true,
}

// BuildPlistDict converts a ServiceConfig into the plist dictionary shape
// expected by launchd. The two callers (UserManager.writePlist and
// SystemManager.encodePlist) previously duplicated this field-by-field
// mapping; centralizing it here keeps future plist keys in one place.
//
// When expandPaths is true, StandardOutPath / StandardErrorPath are passed
// through expandTilde so `~/log.txt` resolves to the user's home. The
// system-daemon encoder uses expandPaths=false: system daemons write
// absolute paths under /var/log, and running as root means `~` would
// resolve to /var/root which the caller never intends.
func BuildPlistDict(config *ServiceConfig, expandPaths bool) map[string]any {
	pd := map[string]any{"Label": config.Label}
	if config.Program != "" {
		pd["Program"] = config.Program
	}
	if len(config.Arguments) > 0 {
		pd["ProgramArguments"] = config.Arguments
	}
	if config.RunAtLoad {
		pd["RunAtLoad"] = true
	}
	if v, ok := buildKeepAlive(config.KeepAlive); ok {
		pd["KeepAlive"] = v
	}
	if config.ThrottleInterval != nil {
		pd["ThrottleInterval"] = *config.ThrottleInterval
	}
	if config.WorkingDir != "" {
		pd["WorkingDirectory"] = config.WorkingDir
	}
	if config.StdoutPath != "" {
		pd["StandardOutPath"] = maybeExpandTilde(config.StdoutPath, expandPaths)
	}
	if config.StderrPath != "" {
		pd["StandardErrorPath"] = maybeExpandTilde(config.StderrPath, expandPaths)
	}
	if config.WakeSystem {
		pd["WakeSystem"] = true
	}
	if len(config.Environment) > 0 {
		pd["EnvironmentVariables"] = config.Environment
	}
	if config.Schedule != nil {
		if config.Schedule.Interval != nil {
			pd["StartInterval"] = *config.Schedule.Interval
		} else {
			pd["StartCalendarInterval"] = buildCalendarInterval(config.Schedule.Schedules)
		}
	}
	return pd
}

// readPlistMap reads the plist at path and parses it into a map[string]any,
// preserving every key — including ones LaunchPal does not model — with the
// plist library's native value types. plist.Unmarshal auto-detects XML and
// binary formats, so both round-trip through the same code path. A read or
// parse failure is returned to the caller so Update can degrade to a fresh
// write (for example a system daemon plist unreadable without Full Disk
// Access).
func readPlistMap(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read plist: %w", err)
	}
	var m map[string]any
	if _, err := plist.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse plist: %w", err)
	}
	return m, nil
}

// mergeUnmodeledKeys returns a new map containing every unmodeled key from
// existing plus every key from modeled, leaving both inputs unchanged. The
// removal set is the full modeledPlistKeys set (not just the keys modeled
// happens to carry), so a modeled key the user cleared — or a mutually
// exclusive key the user toggled away from, e.g. StartInterval →
// StartCalendarInterval — is dropped instead of being inherited back from the
// existing plist. This keeps modeled keys form-authoritative while preserving
// everything LaunchPal does not model.
func mergeUnmodeledKeys(modeled, existing map[string]any) map[string]any {
	merged := make(map[string]any, len(existing)+len(modeled))
	for k, v := range existing {
		if modeledPlistKeys[k] {
			continue
		}
		merged[k] = v
	}
	maps.Copy(merged, modeled)
	return merged
}

// buildCalendarInterval mirrors the historical launchd behavior: a single
// entry is emitted as a dict, multiple entries as an array of dicts, and
// an empty list becomes an empty dict (which launchd interprets as "every
// minute").
func buildCalendarInterval(entries []CalendarEntry) any {
	if len(entries) == 1 {
		return calendarEntryDict(entries[0])
	}
	if len(entries) > 1 {
		arr := make([]map[string]int, len(entries))
		for i, e := range entries {
			arr[i] = calendarEntryDict(e)
		}
		return arr
	}
	return make(map[string]int)
}

// calendarEntryDict turns a CalendarEntry into the integer-valued dict
// launchd expects. Only set fields are present; nil pointers are skipped.
func calendarEntryDict(e CalendarEntry) map[string]int {
	d := make(map[string]int, 5)
	if e.Minute != nil {
		d["Minute"] = *e.Minute
	}
	if e.Hour != nil {
		d["Hour"] = *e.Hour
	}
	if e.Day != nil {
		d["Day"] = *e.Day
	}
	if e.Weekday != nil {
		d["Weekday"] = *e.Weekday
	}
	if e.Month != nil {
		d["Month"] = *e.Month
	}
	return d
}

// maybeExpandTilde applies expandTilde only when enabled; system-domain
// encoding keeps paths verbatim.
func maybeExpandTilde(path string, enabled bool) string {
	if !enabled {
		return path
	}
	return expandTilde(path)
}
