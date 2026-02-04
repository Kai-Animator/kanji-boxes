Summary:
Defines the review session flow and due-only queue behavior.
Specifies flip and answer actions that drive SRS updates.
Ensures review progress is visible during a session.

## ADDED Requirements

### Requirement: Due-only review queue
The system SHALL build the review queue from cards where nextDue is less than or equal to today, ordered by nextDue ascending.

#### Scenario: Queue includes only due cards
- **WHEN** two cards have nextDue of 2026-02-03 and 2026-02-05 and today is 2026-02-04
- **THEN** only the 2026-02-03 card is included in the review queue

### Requirement: Flip and answer actions
The system SHALL allow users to flip a card to reveal the answer and record a correct or incorrect outcome for that card.

#### Scenario: Recording an answer completes a card
- **WHEN** the user marks the current card incorrect
- **THEN** the system applies the SRS movement rules and advances to the next card

### Requirement: Progress display
The system SHALL display session progress as reviewed count out of total due cards.

#### Scenario: Progress updates after review
- **WHEN** a session starts with 5 due cards and the user reviews 2
- **THEN** progress shows 2 of 5 reviewed
