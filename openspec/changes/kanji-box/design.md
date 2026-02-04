Summary:
Defines the package structure and responsibilities for the TUI, domain, and storage layers.
Clarifies local-date handling for scheduling and JSON serialization.
Specifies persistence layout and atomic write strategy.
Documents key UI flow decisions and keyboard interaction approach.

## Context

We are starting a new Go Bubble Tea TUI for a kanji flashcard trainer. Core requirements include a 5-box Leitner SRS, due-only review sessions, local JSON persistence, and CSV import/export. The system must remain offline-first and deterministic, with scheduling logic that is easy to unit test.

## Goals / Non-Goals

**Goals:**
- Separate concerns across domain, model, ui/view, and storage packages.
- Keep scheduling and movement logic pure and testable.
- Use local calendar dates for due calculations and persistence.
- Provide a simple, consistent TUI flow with clear key hints.

**Non-Goals:**
- Cloud sync or multi-device data sharing.
- Advanced scheduling algorithms beyond fixed box intervals.
- Rich editor features beyond basic add/edit/delete.

## Decisions

### Decision 1: Package boundaries
We will use distinct packages: `domain` for Card and SRS logic, `storage` for JSON/CSV I/O, `model` for Bubble Tea state/update, and `ui/view` for rendering. This keeps scheduling logic isolated from UI and I/O, supporting unit tests for SRS rules.

### Decision 2: Date representation
We will store `nextDue` and `lastReviewed` as local dates in `YYYY-MM-DD` and compare by local date in memory. This avoids time zone drift and matches the requirements for due logic.

### Decision 3: Persistence layout
We will store `cards.json` and `state.json` in the OS user config directory by default. Writes will use a temp file and rename to ensure crash safety.

### Decision 4: Review session behavior
The review flow will build a due-only queue ordered by `nextDue`. It will provide flip and answer actions, update SRS state, and show progress for the session.

## Risks / Trade-offs

- Local-date storage may complicate future time-based features → Mitigation: centralize date parsing/formatting in a single helper.
- Atomic file writes require careful error handling → Mitigation: encapsulate write logic in `storage` with clear unit tests.

## Migration Plan

- Not applicable for initial implementation.

## Open Questions

- Should CSV import/export support escaping and quoted fields beyond standard library defaults?
- Do we want a CLI flag for custom storage path in the first iteration?
