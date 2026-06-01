package launchctl

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
