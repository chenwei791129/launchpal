## ADDED Requirements

### Requirement: Reveal plist file in Finder

The system SHALL provide a "Reveal in Finder" action that opens macOS Finder and highlights the plist file for a given service. The action SHALL be available for all service types (User, System, Apple System).

The backend SHALL expose a `RevealInFinder(path string)` method that executes `open -R <path>` via `os/exec`.

#### Scenario: User clicks Reveal in Finder button

- **WHEN** user clicks the "Reveal in Finder" button on the Service Summary page
- **THEN** macOS Finder opens the directory containing the service's plist file with the file selected/highlighted

#### Scenario: Reveal in Finder for system service

- **WHEN** user views a System or Apple System service and clicks the "Reveal in Finder" button
- **THEN** Finder opens the corresponding directory (`/Library/LaunchDaemons` or `/System/Library/LaunchDaemons`) with the plist file highlighted

#### Scenario: Button does not trigger copy

- **WHEN** user clicks the "Reveal in Finder" button
- **THEN** the plist path is NOT copied to clipboard (the copy action is not triggered)
