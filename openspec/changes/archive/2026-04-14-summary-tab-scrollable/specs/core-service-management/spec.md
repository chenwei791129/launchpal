## ADDED Requirements

### Requirement: Summary tab content is scrollable

The Summary tab panel SHALL be vertically scrollable when content exceeds the visible area, consistent with the Edit and Inspect tabs.

#### Scenario: Summary tab with overflowing content

- **WHEN** a service has enough detail (environment variables, schedule, paths, logs) that the Summary tab content exceeds the viewport height
- **THEN** a vertical scrollbar SHALL appear allowing the user to scroll to see all content

#### Scenario: Summary tab with minimal content

- **WHEN** a service has minimal detail that fits within the viewport
- **THEN** no scrollbar SHALL appear and all content SHALL be visible without scrolling
