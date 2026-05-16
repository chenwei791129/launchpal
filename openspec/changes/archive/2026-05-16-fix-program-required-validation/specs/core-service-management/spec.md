## MODIFIED Requirements

### Requirement: CRUD operations for user services

The system SHALL create a new service by writing a plist to `~/Library/LaunchAgents/<label>.plist`.
The system SHALL reject creation when the label is empty.
The system SHALL reject creation when a service with the same label already exists.
The system SHALL reject creation when both `Program` is empty AND `Arguments` is empty, returning an error indicating that the service must specify either Program or at least one argument in Arguments.
The system SHALL reject update when both `Program` is empty AND `Arguments` is empty, returning the same error.
The system SHALL ensure the LaunchAgents directory exists before creating.
The system SHALL ensure log directories exist for configured stdout/stderr paths.
The system SHALL update an existing service by stopping it first, then writing the updated plist.
The system SHALL delete a service by stopping it first, then removing the plist file.

#### Scenario: Create a new service

- **WHEN** Create is called with a valid ServiceConfig (label="com.user.test", program="/usr/bin/test")
- **THEN** a plist file is created at `~/Library/LaunchAgents/com.user.test.plist`

#### Scenario: Create with duplicate label

- **WHEN** Create is called with a label that already has a plist file
- **THEN** the system returns an error indicating the service already exists

#### Scenario: Create with empty label

- **WHEN** Create is called with an empty label
- **THEN** the system returns an error indicating "service label is required"

#### Scenario: Create with only Arguments and no Program

- **WHEN** Create is called with a ServiceConfig where Program is empty but Arguments is a non-empty array (e.g., arguments=["/usr/bin/open", "/Applications/Synology Drive Client.app"])
- **THEN** the plist is written successfully and contains `ProgramArguments` but no `Program` key

#### Scenario: Create with neither Program nor Arguments

- **WHEN** Create is called with a ServiceConfig where Program is empty AND Arguments is empty or nil
- **THEN** the system returns an error indicating that the service must specify either Program or at least one argument in Arguments
- **AND** no plist file is written

#### Scenario: Update with neither Program nor Arguments

- **WHEN** Update is called for an existing service with a ServiceConfig where Program is empty AND Arguments is empty or nil
- **THEN** the system returns an error indicating that the service must specify either Program or at least one argument in Arguments
- **AND** the existing plist file is not modified

#### Scenario: Delete an existing service

- **WHEN** Delete is called for an existing service
- **THEN** the plist file is removed from disk

#### Scenario: Delete nonexistent service

- **WHEN** Delete is called for a service that does not exist
- **THEN** the system returns an error indicating "service not found"
