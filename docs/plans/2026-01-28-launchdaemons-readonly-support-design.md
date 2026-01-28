# LaunchDaemons Read-Only Support - Design Document

**Date**: 2026-01-28
**Status**: Approved
**GitHub Issue**: [#1](https://github.com/chenwei791129/launchpal/issues/1)

## Overview

Add read-only support for system-level LaunchDaemons to LaunchPal, allowing users to monitor and inspect system services while maintaining current user-level LaunchAgents management capabilities.

## Goals

1. **Read-only monitoring** of system LaunchDaemons
2. **Clear UI distinction** between user and system services
3. **Graceful permission handling** with user-friendly guidance
4. **Extensible architecture** for future read-write capabilities

## Architecture Design

### Core Concept: Extensible Manager Architecture

Current architecture uses `Manager` interface implemented by `UserManager` for `~/Library/LaunchAgents`. We will add two new managers sharing the same interface:

1. **`SystemManager`** - Manages `/Library/LaunchDaemons` (third-party system services)
2. **`AppleSystemManager`** - Manages `/System/Library/LaunchDaemons` (Apple system services)

### Manager Characteristics

| Manager | Path | Access Level | Operations |
|---------|------|--------------|------------|
| UserManager | `~/Library/LaunchAgents` | Read-Write | All operations |
| SystemManager | `/Library/LaunchDaemons` | Read-Only (Phase 1) | List, Get, GetPlist, GetLogs |
| AppleSystemManager | `/System/Library/LaunchDaemons` | Read-Only (Phase 1) | List, Get, GetPlist, GetLogs |

### Type System Changes

**Service struct** (`internal/launchctl/types.go`):
```go
type Service struct {
    // ... existing fields ...
    Type     string `json:"type"`     // "user", "system", "apple-system"
    ReadOnly bool   `json:"readOnly"` // true for system services
}
```

**Error definitions**:
```go
var ErrReadOnlyManager = errors.New("this manager is read-only")
```

### Implementation Pattern

Read-only managers implement only safe methods:
- `List()` - List all services in directory
- `Get(name)` - Get service details
- `GetPlist(name)` - Read raw plist content
- `GetLogs(name, type)` - Read log files

Write methods return `ErrReadOnlyManager`:
- `Start(name)` → error
- `Stop(name)` → error
- `Create(config)` → error
- `Update(name, config)` → error
- `Delete(name)` → error
- `Restart(name)` → error

## Permission Detection & Handling

### Startup Permission Check

Add method to `App` struct:
```go
// CheckPermissions returns permission status for each service domain
func (a *App) CheckPermissions() map[string]bool {
    return map[string]bool{
        "user":         true, // Always true
        "system":       canReadDirectory("/Library/LaunchDaemons"),
        "apple-system": canReadDirectory("/System/Library/LaunchDaemons"),
    }
}
```

### UI Permission Feedback

When loading system/apple-system tabs with insufficient permissions:

```
⚠️ Limited Access: Some system services may not be visible.
LaunchPal needs disk access permission to read system LaunchDaemons.

[Open System Settings] button
```

## UI Integration

### Tab Structure

```
Sidebar Tabs:
├─ Services (current)          → UserManager
├─ System Services (new)       → SystemManager
└─ Apple Services (new)        → AppleSystemManager
```

### UI Behavior for Read-Only Services

- Hide "New Service" button on system tabs
- Disable/hide Start/Stop/Delete action buttons
- Display "Read-only" badge on service details
- Allow viewing plist and logs (read operations)

### Frontend API Methods

Add to `app.go`:
```go
// ListSystemServices returns all LaunchDaemon services from /Library
func (a *App) ListSystemServices() ([]launchctl.Service, error)

// ListAppleSystemServices returns all LaunchDaemon services from /System/Library
func (a *App) ListAppleSystemServices() ([]launchctl.Service, error)

// GetSystemService returns a single system service by name and type
func (a *App) GetSystemService(name string, serviceType string) (*launchctl.Service, error)
```

## Error Handling

### Edge Cases

1. **Invalid plist files**
   - Show in list as "Invalid Configuration"
   - Allow viewing raw XML for debugging

2. **Individual file permission errors**
   - Display service with "Access Denied" status
   - Show in service list but indicate unavailability

3. **Directory permission errors**
   - Show permission prompt banner
   - Empty list with helpful guidance

4. **System service status detection**
   - Use `launchctl list` (no root required)
   - Root-owned services may show "loaded" instead of "running" with PID

5. **Performance with large service counts**
   - `/System/Library/LaunchDaemons` may contain 100+ services
   - Use pagination or virtual scrolling
   - Keep search functionality enabled

## Implementation Scope

### Phase 1 (This Implementation)

**Backend**:
- Create `SystemManager` struct
- Create `AppleSystemManager` struct
- Implement read-only methods
- Add `Type` and `ReadOnly` fields to `Service`
- Add `ErrReadOnlyManager` error

**Frontend**:
- Add two new sidebar tabs
- Implement permission detection UI
- Conditionally hide/disable write operations
- Add "Read-only" visual indicators

**Testing**:
- Unit tests for List/Get methods
- Manual testing on real macOS environment
- Permission handling verification

### Future Extensibility (Phase 2)

When upgrading to read-write support:
- Implement write methods in managers
- Add privilege escalation (using `osascript` for sudo commands)
- Add confirmation dialogs for system modifications
- Enhanced operation logging
- Backup system before modifications

## Technical Considerations

### Status Detection

Current status detection in `UserManager.getServiceStatus()`:
1. Try `launchctl list <label>` to get PID
2. Fallback to `pgrep -f <program>` for running processes
3. Skip common shells to avoid false positives

This approach works for system services too, no changes needed.

### File Permissions

- User has read access to most `/Library/LaunchDaemons` plists
- `/System/Library/LaunchDaemons` may require Full Disk Access permission
- Handle `os.ErrPermission` gracefully at both file and directory levels

### Performance

- Lazy load services when tab is activated
- Cache results with refresh button
- Consider background loading for large directories

## Security Considerations

- Read-only operations pose minimal security risk
- No privilege escalation in Phase 1
- File system access limited to standard user permissions
- Future write operations will require explicit user authorization

## Success Criteria

- Users can view all system LaunchDaemons they have permission to read
- Clear visual distinction between user and system services
- Graceful degradation when permissions are insufficient
- No regression in existing UserManager functionality
- Architecture supports future write operations without major refactoring
