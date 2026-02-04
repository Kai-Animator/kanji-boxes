Summary:
Defines local JSON persistence for cards and app state.
Requires crash-safe atomic writes via temp file and rename.
Specifies storage location and date serialization rules.

## ADDED Requirements

### Requirement: Local JSON storage
The system SHALL persist cards and app state to local JSON files on disk.

#### Scenario: Save on state change
- **WHEN** a card is added or updated
- **THEN** the system writes updated JSON to disk

### Requirement: Atomic writes
The system SHALL write JSON using a temp file and rename to prevent partial writes.

#### Scenario: Atomic save behavior
- **WHEN** saving cards to disk
- **THEN** the system writes to a temp file and renames it to the target file name

### Requirement: Storage location
The system SHALL store data in the user's configuration directory by default.

#### Scenario: Default path
- **WHEN** no override is configured
- **THEN** data is stored under the OS user config directory

### Requirement: Date serialization
The system SHALL store nextDue and lastReviewed as local calendar dates in YYYY-MM-DD format.

#### Scenario: Persist date fields
- **WHEN** a card is saved with nextDue set to 2026-02-04
- **THEN** the JSON contains "nextDue": "2026-02-04"
