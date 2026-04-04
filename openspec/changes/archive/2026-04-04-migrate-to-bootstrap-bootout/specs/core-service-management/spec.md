## ADDED Requirements

### Requirement: Start user service via bootstrap

The system SHALL start a user service by executing `launchctl bootstrap gui/<uid> <plistPath>` where `<uid>` is the current user's UID obtained via `os.Getuid()`.
The system SHALL return an error when the plist file does not exist.
The system SHALL return an error with the launchctl output when bootstrap fails (e.g., service already loaded).

#### Scenario: Start a stopped service

- **WHEN** Start is called for an existing service that is not currently loaded
- **THEN** the system executes `launchctl bootstrap gui/<uid> <plistPath>` and returns no error

#### Scenario: Start an already loaded service

- **WHEN** Start is called for a service that is already loaded
- **THEN** the system returns an error containing the launchctl error output

#### Scenario: Start a nonexistent service

- **WHEN** Start is called for a service whose plist file does not exist
- **THEN** the system returns an error indicating "service not found"

### Requirement: Stop user service via bootout

The system SHALL stop a user service by executing `launchctl bootout gui/<uid>/<label>` where `<uid>` is the current user's UID and `<label>` is the service's Label from the plist.
The system SHALL ignore errors from bootout (the service may not be loaded).
The system SHALL attempt to kill the process via pgrep/kill as a fallback if the service program is still running after bootout.
The system SHALL skip pgrep fallback for common shell programs (`/bin/bash`, `/bin/sh`, `/bin/zsh` and their `/usr/bin` variants).

#### Scenario: Stop a running service

- **WHEN** Stop is called for a loaded service with label "com.user.app"
- **THEN** the system executes `launchctl bootout gui/<uid>/com.user.app`

#### Scenario: Stop a service that is not loaded

- **WHEN** Stop is called for a service that is not currently loaded
- **THEN** the system ignores the bootout error and returns no error

#### Scenario: Stop with fallback kill

- **WHEN** Stop is called and the service process is still running after bootout
- **THEN** the system uses pgrep to find and kill the process

### Requirement: GUI domain helper

The system SHALL provide a helper function that returns the `gui/<uid>` domain string using the current process's UID.
The system SHALL format the domain as `gui/<uid>` where `<uid>` is a decimal integer.

#### Scenario: Get GUI domain for current user

- **WHEN** the helper function is called by a process running as UID 501
- **THEN** it returns "gui/501"
