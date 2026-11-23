# Preserve unmodeled plist keys on edit

LaunchPal manages existing launchd jobs, including jobs with settings the editor does not expose. When updating a readable plist, keep unmodeled keys and replace the entire set of modeled keys with what the editor emits. Replacing only keys emitted by the current form would restore values the user intentionally cleared, such as `RunAtLoad` when switching to On Demand. Rebuilding every plist from scratch would silently drop settings such as `Disabled`.

Keep the read and merge in the GUI rather than adding privileged read-and-merge operations to the helper. If the original plist cannot be read or parsed, the edit can still proceed, but unmodeled keys cannot be preserved. This trade-off matters for system daemons when the GUI lacks Full Disk Access.
