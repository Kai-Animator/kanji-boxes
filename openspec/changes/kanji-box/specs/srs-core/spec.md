Summary:
Defines the core scheduling and movement rules for the 5-box SRS.
Establishes due calculation using local calendar dates.
Specifies how review outcomes update box and dates.
Keeps scheduling logic deterministic and testable.

## ADDED Requirements

### Requirement: Box-based scheduling
The system SHALL set a card's nextDue to today plus the interval for the card's current box (1–5 days).

#### Scenario: Schedule after box movement
- **WHEN** a card moves to Box 3 on local date 2026-02-04
- **THEN** the system sets nextDue to 2026-02-07

### Requirement: Due determination
The system SHALL consider a card due when nextDue is less than or equal to the user's local date.

#### Scenario: Due comparison uses local date
- **WHEN** today is 2026-02-04 and nextDue is 2026-02-04
- **THEN** the card is due

### Requirement: Review outcome movement
The system SHALL move a card up one box on a correct answer (max Box 5) and down one box on an incorrect answer (min Box 1).

#### Scenario: Correct answer promotes box
- **WHEN** a Box 4 card is marked correct
- **THEN** the card moves to Box 5

#### Scenario: Incorrect answer demotes box
- **WHEN** a Box 1 card is marked incorrect
- **THEN** the card remains in Box 1

### Requirement: Review result updates
The system SHALL update lastReviewed and nextDue when a review outcome is recorded.

#### Scenario: Update review metadata
- **WHEN** a user marks a card correct on local date 2026-02-04
- **THEN** lastReviewed is set to 2026-02-04 and nextDue is set based on the new box
