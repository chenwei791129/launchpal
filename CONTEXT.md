# LaunchPal

LaunchPal is a macOS app for inspecting and managing services registered with launchd.

## Language

### Service categories

**User Service**:
A LaunchAgent that belongs to a single macOS user.

**System Service**:
A system-wide LaunchDaemon that is not part of macOS itself. This category excludes Apple System Services.

**Apple System Service**:
A LaunchDaemon supplied by macOS. It is always read-only in LaunchPal.

### Authorization

**Admin Mode**:
LaunchPal's session-scoped authorization to manage System Services. It does not extend to Apple System Services.

### Service status

For System Services and Apple System Services, these statuses are estimates rather than direct confirmation from launchd.

**Loaded**:
For a User Service, the job is registered with launchd but has no active process.

**Running**:
For a User Service, launchd reports an active process for the job.
