Summary:
Defines the session stats shown in the stats screen.
Tracks reviewed, correct, and incorrect counts per session.
Ensures stats update as the user reviews cards.

## ADDED Requirements

### Requirement: Session stats display
The system SHALL display reviewed, correct, and incorrect counts for the current session.

#### Scenario: Show session totals
- **WHEN** the user opens the stats screen
- **THEN** the screen shows reviewed, correct, and incorrect counts

### Requirement: Session stats update
The system SHALL update session stats after each answered review.

#### Scenario: Update after answer
- **WHEN** a user marks a card correct
- **THEN** reviewed and correct counts increase by 1
