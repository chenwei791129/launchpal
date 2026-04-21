## MODIFIED Requirements

### Requirement: Get system service details

SystemManager.Get SHALL read and parse a plist from `/Library/LaunchDaemons/<name>.plist`.
AppleSystemManager.Get SHALL read and parse a plist from `/System/Library/LaunchDaemons/<name>.plist`.
Both SHALL set the `Type` field to their respective type (`"system"` or `"apple-system"`).
Both SHALL set `ReadOnly` to `true`.
Both SHALL detect plist format (xml or binary) and populate `PlistFormat`.
Both SHALL populate `Status`, `PID`, and `StatusConfidence` according to the `system-daemon-status-detection` capability.
Both SHALL return a permission denied error when the file is not readable.

#### Scenario: Get a system service

- **WHEN** Get is called with a valid service name in `/Library/LaunchDaemons`
- **THEN** the returned Service has Type=`"system"`, ReadOnly=`true`, all parsed plist fields populated, and Status/PID/StatusConfidence populated per heuristic detection

#### Scenario: Get an Apple system service

- **WHEN** Get is called with a valid service name in `/System/Library/LaunchDaemons`
- **THEN** the returned Service has Type=`"apple-system"`, ReadOnly=`true`, all parsed plist fields populated, and Status/PID/StatusConfidence populated per heuristic detection

#### Scenario: Permission denied on Get

- **WHEN** Get is called for a plist file the user cannot read
- **THEN** the system returns an error containing "permission denied"
