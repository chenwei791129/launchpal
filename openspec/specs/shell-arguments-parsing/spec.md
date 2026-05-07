# shell-arguments-parsing Specification

## Purpose

Plist `ProgramArguments` editing accepts a single space-separated text input that the frontend parses into an argv array, treating single-quoted (`'...'`) and double-quoted (`"..."`) segments as one argument each with the quotes stripped. Round-tripping through `serializeShellArgs` reproduces the original text so editing existing services is non-destructive.

## Requirements

### Requirement: Parse quoted arguments from text input

The system SHALL parse a space-separated argument string into an array, treating single-quoted (`'...'`) and double-quoted (`"..."`) segments as single arguments with the quotes stripped.

Unquoted tokens SHALL be split on whitespace boundaries.

#### Scenario: Simple unquoted arguments

- **WHEN** user enters `--print --verbose --model opus`
- **THEN** the result SHALL be `["--print", "--verbose", "--model", "opus"]`

#### Scenario: Single-quoted argument with spaces

- **WHEN** user enters `--print -p 'run daily backup and send report'`
- **THEN** the result SHALL be `["--print", "-p", "run daily backup and send report"]`

#### Scenario: Double-quoted argument with spaces

- **WHEN** user enters `--message "hello world" --verbose`
- **THEN** the result SHALL be `["--message", "hello world", "--verbose"]`

#### Scenario: Mixed quoted and unquoted arguments

- **WHEN** user enters `--flag 'value one' --other "value two" plain`
- **THEN** the result SHALL be `["--flag", "value one", "--other", "value two", "plain"]`

#### Scenario: Empty input

- **WHEN** user enters an empty string or whitespace-only string
- **THEN** the result SHALL be an empty array `[]`


<!-- @trace
source: fix-arguments-parsing
updated: 2026-03-30
code:
  - shioaji.log
  - frontend/app/pages/services/[name].vue
  - frontend/package.json
  - frontend/app/utils/shell-args.ts
  - frontend/package.json.md5
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/components/ServiceSummary.vue
tests:
  - frontend/app/utils/__tests__/shell-args.test.ts
-->

---
### Requirement: Serialize argument array to text with quoting

The system SHALL serialize an argument array back to a space-separated string. Arguments containing whitespace SHALL be wrapped in single quotes. Arguments without whitespace SHALL remain unquoted.

#### Scenario: Round-trip consistency

- **WHEN** an argument array `["--print", "-p", "run daily backup and send report"]` is serialized to text
- **THEN** the result SHALL be `--print -p 'run daily backup and send report'`

#### Scenario: Arguments without spaces

- **WHEN** an argument array `["--verbose", "--model", "opus"]` is serialized to text
- **THEN** the result SHALL be `--verbose --model opus`


<!-- @trace
source: fix-arguments-parsing
updated: 2026-03-30
code:
  - shioaji.log
  - frontend/app/pages/services/[name].vue
  - frontend/package.json
  - frontend/app/utils/shell-args.ts
  - frontend/package.json.md5
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/components/ServiceSummary.vue
tests:
  - frontend/app/utils/__tests__/shell-args.test.ts
-->

---
### Requirement: Consistent parsing across create and edit flows

Both the Create Service modal and the Edit Service form SHALL use the same parsing function to convert the arguments text input into an array.

Both the Edit form population and the Summary display SHALL use the same serialization function to convert the argument array back to text.

#### Scenario: Create and edit produce identical results

- **WHEN** a user creates a service with arguments `-p '含空格的參數'`
- **AND** later edits the same service without modifying arguments
- **AND** saves the edit form
- **THEN** the plist ProgramArguments array SHALL remain unchanged

<!-- @trace
source: fix-arguments-parsing
updated: 2026-03-30
code:
  - shioaji.log
  - frontend/app/pages/services/[name].vue
  - frontend/package.json
  - frontend/app/utils/shell-args.ts
  - frontend/package.json.md5
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/components/ServiceSummary.vue
tests:
  - frontend/app/utils/__tests__/shell-args.test.ts
-->