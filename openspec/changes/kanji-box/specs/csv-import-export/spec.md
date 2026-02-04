Summary:
Defines CSV import/export format for card content.
Uses fixed column order and pipe-delimited tags.
Specifies default behavior for newly imported cards.

## ADDED Requirements

### Requirement: CSV export columns
The system SHALL export cards in CSV with columns: kanji, hiragana, english, tags.

#### Scenario: Export header order
- **WHEN** the user exports cards
- **THEN** the CSV header is "kanji,hiragana,english,tags"

### Requirement: CSV import defaults
The system SHALL create imported cards in Box 1 with nextDue set to today.

#### Scenario: Import sets initial scheduling
- **WHEN** a row is imported on local date 2026-02-04
- **THEN** the resulting card has box 1 and nextDue 2026-02-04

### Requirement: Tag parsing
The system SHALL parse the tags column using a pipe (|) delimiter.

#### Scenario: Parse tag list
- **WHEN** tags are "jlpt5|verbs|common"
- **THEN** the card tags are ["jlpt5", "verbs", "common"]
