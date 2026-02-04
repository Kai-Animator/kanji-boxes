Summary:
Breaks implementation into setup, domain logic, storage, and UI flows.
Prioritizes pure SRS logic with tests before TUI integration.
Includes CSV import/export and stats requirements.
Ends with verification steps for core behavior.

## 1. Project Setup

- [x] 1.1 Initialize Go module and add Bubble Tea/Bubbles/Lip Gloss dependencies
- [x] 1.2 Create initial folder structure: domain, storage, model, ui/view

## 2. Domain SRS Logic

- [x] 2.1 Define Card and SessionStats types with date helpers
- [x] 2.2 Implement box movement and nextDue calculation functions
- [x] 2.3 Implement due filtering and queue ordering helpers
- [x] 2.4 Add unit tests for movement and due logic

## 3. Storage and CSV

- [x] 3.1 Implement JSON load/save with atomic write (cards.json, state.json)
- [x] 3.2 Add config-dir path resolution and optional override hook
- [x] 3.3 Implement CSV export with required columns
- [x] 3.4 Implement CSV import with tag parsing and default scheduling

## 4. TUI Shell

- [x] 4.1 Create Bubble Tea model skeleton and main menu
- [x] 4.2 Add base view styles and footer hint pattern

## 5. Review Session Flow

- [x] 5.1 Build due-only review queue from domain helpers
- [x] 5.2 Implement flip and answer actions with SRS updates
- [x] 5.3 Show progress indicator during review

## 6. Stats Screen

- [x] 6.1 Track session reviewed/correct/incorrect counts
- [x] 6.2 Render stats screen with session totals

## 7. Verification

- [x] 7.1 Manual smoke test: review session updates boxes and nextDue
- [x] 7.2 Manual smoke test: CSV import/export round-trip
