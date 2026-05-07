# backup-diff-preview Specification

## Purpose

The Settings → Backup History list exposes a Diff button next to each Restore button so users can preview what a backup would change before committing to a restore. The preview renders a side-by-side diff (current on the left, backup on the right, red/green coloring) capped at 10,000 lines, with the backend converting binary plists to XML so both sides remain human-readable.

## Requirements

### Requirement: Diff button next to Restore

The Settings page Backup History list SHALL display a Diff button immediately to the left of each Restore button.

#### Scenario: Diff button visible for every backup row

- **WHEN** the user opens the Settings page and the Backup History contains at least one backup
- **THEN** each row SHALL render a Diff button positioned immediately before the Restore button
- **AND** the Diff button SHALL have an accessible tooltip describing its purpose

#### Scenario: No backups means no Diff buttons

- **WHEN** the Backup History list is empty
- **THEN** no Diff buttons SHALL be rendered


<!-- @trace
source: add-backup-diff-view
updated: 2026-04-18
code:
  - frontend/app/composables/useBackupDiff.ts
  - internal/launchctl/readonly.go
  - README.md
  - frontend/package.json
  - frontend/app/utils/formatters.ts
  - internal/launchctl/types.go
  - internal/launchctl/user.go
  - frontend/package.json.md5
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/wailsjs/go/main/App.js
  - frontend/app/components/BackupDiffDialog.vue
  - frontend/app/types/wails.d.ts
  - internal/backup/backup.go
  - app.go
  - frontend/app/pages/settings.vue
  - frontend/wailsjs/go/models.ts
  - internal/plistutil/plistutil.go
tests:
  - internal/backup/backup_test.go
  - frontend/app/components/__tests__/BackupDiffDialog.test.ts
  - internal/plistutil/testhelpers_test.go
  - frontend/app/composables/__tests__/useBackupDiff.test.ts
  - internal/launchctl/user_test.go
  - internal/plistutil/plistutil_test.go
-->

---
### Requirement: Diff button shows a custom hover tooltip explaining its function

The Diff button SHALL display a visible custom tooltip on hover that explains the button's purpose, rather than relying solely on the browser's native `title` attribute tooltip.

#### Scenario: Tooltip appears on hover

- **WHEN** the user hovers the mouse cursor over the Diff button
- **THEN** a custom tooltip SHALL become visible near the button
- **AND** the tooltip SHALL contain a short description of the button's function (e.g. preview the diff against the current plist)

#### Scenario: Tooltip disappears when hover ends

- **WHEN** the mouse cursor leaves the Diff button
- **THEN** the custom tooltip SHALL become hidden

#### Scenario: Tooltip does not interfere with clicks

- **WHEN** the user clicks the Diff button while the tooltip is visible
- **THEN** the click SHALL be handled by the button (opening the diff modal)
- **AND** the tooltip SHALL NOT intercept the click event


<!-- @trace
source: add-backup-diff-view
updated: 2026-04-18
code:
  - frontend/app/composables/useBackupDiff.ts
  - internal/launchctl/readonly.go
  - README.md
  - frontend/package.json
  - frontend/app/utils/formatters.ts
  - internal/launchctl/types.go
  - internal/launchctl/user.go
  - frontend/package.json.md5
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/wailsjs/go/main/App.js
  - frontend/app/components/BackupDiffDialog.vue
  - frontend/app/types/wails.d.ts
  - internal/backup/backup.go
  - app.go
  - frontend/app/pages/settings.vue
  - frontend/wailsjs/go/models.ts
  - internal/plistutil/plistutil.go
tests:
  - internal/backup/backup_test.go
  - frontend/app/components/__tests__/BackupDiffDialog.test.ts
  - internal/plistutil/testhelpers_test.go
  - frontend/app/composables/__tests__/useBackupDiff.test.ts
  - internal/launchctl/user_test.go
  - internal/plistutil/plistutil_test.go
-->

---
### Requirement: Diff modal shows side-by-side diff with current on left and backup on right

Clicking the Diff button SHALL open a modal that displays a side-by-side diff. The left column SHALL show the current plist content and the right column SHALL show the backup content, with rows aligned so that corresponding lines appear at the same vertical position.

#### Scenario: Backup differs from current plist

- **WHEN** the user clicks the Diff button for a backup whose content differs from the current plist
- **THEN** a modal SHALL open containing two columns: current on the left and backup on the right
- **AND** lines present only on the right column (additions to be written by Restore) SHALL be visually distinguished with a green background and green text
- **AND** lines present only on the left column (content to be removed by Restore) SHALL be visually distinguished with a red background and red text
- **AND** for each removed-only line the right column SHALL render an empty placeholder row at the same vertical position
- **AND** for each added-only line the left column SHALL render an empty placeholder row at the same vertical position
- **AND** unchanged lines SHALL appear on both columns at the same vertical position without add/remove styling

#### Scenario: Backup identical to current plist

- **WHEN** the user clicks the Diff button for a backup whose content matches the current plist byte-for-byte after normalization
- **THEN** the modal SHALL open and display an empty-diff indicator message stating that no changes would occur

#### Scenario: Modal shows backup metadata

- **WHEN** the Diff modal is open
- **THEN** the modal header SHALL display the service name and the backup timestamp


<!-- @trace
source: add-backup-diff-view
updated: 2026-04-18
code:
  - frontend/app/composables/useBackupDiff.ts
  - internal/launchctl/readonly.go
  - README.md
  - frontend/package.json
  - frontend/app/utils/formatters.ts
  - internal/launchctl/types.go
  - internal/launchctl/user.go
  - frontend/package.json.md5
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/wailsjs/go/main/App.js
  - frontend/app/components/BackupDiffDialog.vue
  - frontend/app/types/wails.d.ts
  - internal/backup/backup.go
  - app.go
  - frontend/app/pages/settings.vue
  - frontend/wailsjs/go/models.ts
  - internal/plistutil/plistutil.go
tests:
  - internal/backup/backup_test.go
  - frontend/app/components/__tests__/BackupDiffDialog.test.ts
  - internal/plistutil/testhelpers_test.go
  - frontend/app/composables/__tests__/useBackupDiff.test.ts
  - internal/launchctl/user_test.go
  - internal/plistutil/plistutil_test.go
-->

---
### Requirement: Binary plist normalized to XML before diffing

The system SHALL normalize binary-format plists to XML before computing the diff so the user sees human-readable text.

#### Scenario: Backup file is in binary plist format

- **WHEN** the Diff button is clicked for a backup whose file is in binary plist format
- **THEN** the backend SHALL convert the backup content to XML plist format before returning it
- **AND** the diff displayed SHALL be a text diff of XML representations

#### Scenario: Current plist is in binary plist format

- **WHEN** the Diff button is clicked and the current plist for the service is in binary format
- **THEN** the backend SHALL convert the current plist content to XML plist format before returning it

#### Scenario: Format conversion fails

- **WHEN** binary-to-XML conversion fails for either side
- **THEN** the system SHALL fall back to the raw file content
- **AND** the modal SHALL display a warning banner indicating that format conversion failed and the diff output is likely unreadable


<!-- @trace
source: add-backup-diff-view
updated: 2026-04-18
code:
  - frontend/app/composables/useBackupDiff.ts
  - internal/launchctl/readonly.go
  - README.md
  - frontend/package.json
  - frontend/app/utils/formatters.ts
  - internal/launchctl/types.go
  - internal/launchctl/user.go
  - frontend/package.json.md5
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/wailsjs/go/main/App.js
  - frontend/app/components/BackupDiffDialog.vue
  - frontend/app/types/wails.d.ts
  - internal/backup/backup.go
  - app.go
  - frontend/app/pages/settings.vue
  - frontend/wailsjs/go/models.ts
  - internal/plistutil/plistutil.go
tests:
  - internal/backup/backup_test.go
  - frontend/app/components/__tests__/BackupDiffDialog.test.ts
  - internal/plistutil/testhelpers_test.go
  - frontend/app/composables/__tests__/useBackupDiff.test.ts
  - internal/launchctl/user_test.go
  - internal/plistutil/plistutil_test.go
-->

---
### Requirement: Diff works when current plist is absent

The Diff button SHALL remain functional when the current plist file does not exist, showing the entire backup content as additions.

#### Scenario: Service has been deleted

- **WHEN** the user clicks the Diff button for a backup whose service no longer exists on the system
- **THEN** the Diff button SHALL NOT be disabled
- **AND** the modal SHALL open with the left (current) column rendered entirely as placeholder rows
- **AND** the right (backup) column SHALL render every line of the backup with the addition styling
- **AND** the modal SHALL display an indicator that no current version exists

#### Scenario: Current plist file missing at original path

- **WHEN** the service's current plist file is missing at its expected path
- **THEN** the system SHALL treat the current content as an empty string
- **AND** the left column SHALL be rendered entirely as placeholder rows
- **AND** the right column SHALL render the full backup content with the addition styling


<!-- @trace
source: add-backup-diff-view
updated: 2026-04-18
code:
  - frontend/app/composables/useBackupDiff.ts
  - internal/launchctl/readonly.go
  - README.md
  - frontend/package.json
  - frontend/app/utils/formatters.ts
  - internal/launchctl/types.go
  - internal/launchctl/user.go
  - frontend/package.json.md5
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/wailsjs/go/main/App.js
  - frontend/app/components/BackupDiffDialog.vue
  - frontend/app/types/wails.d.ts
  - internal/backup/backup.go
  - app.go
  - frontend/app/pages/settings.vue
  - frontend/wailsjs/go/models.ts
  - internal/plistutil/plistutil.go
tests:
  - internal/backup/backup_test.go
  - frontend/app/components/__tests__/BackupDiffDialog.test.ts
  - internal/plistutil/testhelpers_test.go
  - frontend/app/composables/__tests__/useBackupDiff.test.ts
  - internal/launchctl/user_test.go
  - internal/plistutil/plistutil_test.go
-->

---
### Requirement: Diff modal supports Cancel and Restore actions

The Diff modal SHALL provide explicit Cancel and Restore controls.

#### Scenario: User cancels from diff modal

- **WHEN** the user clicks the Cancel button or clicks outside the modal
- **THEN** the modal SHALL close without modifying any plist file

#### Scenario: User restores from diff modal

- **WHEN** the user clicks the Restore button inside the Diff modal
- **THEN** the system SHALL trigger the existing Restore confirmation flow for that backup


<!-- @trace
source: add-backup-diff-view
updated: 2026-04-18
code:
  - frontend/app/composables/useBackupDiff.ts
  - internal/launchctl/readonly.go
  - README.md
  - frontend/package.json
  - frontend/app/utils/formatters.ts
  - internal/launchctl/types.go
  - internal/launchctl/user.go
  - frontend/package.json.md5
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/wailsjs/go/main/App.js
  - frontend/app/components/BackupDiffDialog.vue
  - frontend/app/types/wails.d.ts
  - internal/backup/backup.go
  - app.go
  - frontend/app/pages/settings.vue
  - frontend/wailsjs/go/models.ts
  - internal/plistutil/plistutil.go
tests:
  - internal/backup/backup_test.go
  - frontend/app/components/__tests__/BackupDiffDialog.test.ts
  - internal/plistutil/testhelpers_test.go
  - frontend/app/composables/__tests__/useBackupDiff.test.ts
  - internal/launchctl/user_test.go
  - internal/plistutil/plistutil_test.go
-->

---
### Requirement: Large diff output is bounded

The diff view SHALL bound rendered output to protect UI responsiveness.

#### Scenario: Diff exceeds display limit

- **WHEN** the number of rows required by either column of the side-by-side diff exceeds 10,000
- **THEN** the modal SHALL render only the first 10,000 rows of each column
- **AND** the modal SHALL display a notice stating that the diff was truncated and how many rows were omitted

<!-- @trace
source: add-backup-diff-view
updated: 2026-04-18
code:
  - frontend/app/composables/useBackupDiff.ts
  - internal/launchctl/readonly.go
  - README.md
  - frontend/package.json
  - frontend/app/utils/formatters.ts
  - internal/launchctl/types.go
  - internal/launchctl/user.go
  - frontend/package.json.md5
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/wailsjs/go/main/App.js
  - frontend/app/components/BackupDiffDialog.vue
  - frontend/app/types/wails.d.ts
  - internal/backup/backup.go
  - app.go
  - frontend/app/pages/settings.vue
  - frontend/wailsjs/go/models.ts
  - internal/plistutil/plistutil.go
tests:
  - internal/backup/backup_test.go
  - frontend/app/components/__tests__/BackupDiffDialog.test.ts
  - internal/plistutil/testhelpers_test.go
  - frontend/app/composables/__tests__/useBackupDiff.test.ts
  - internal/launchctl/user_test.go
  - internal/plistutil/plistutil_test.go
-->