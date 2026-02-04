Summary:
This change introduces a Bubble Tea TUI for a 5-box kanji flashcard trainer.
It defines core SRS scheduling and card movement rules with due-based reviews.
It adds local JSON persistence, CSV import/export, and basic stats.
It establishes the initial user flows for review, browsing, and card management.

## Why

Studying kanji with spaced repetition is easier and more consistent when the workflow is streamlined and offline-first. A focused TUI helps practice daily while keeping scheduling rules predictable and testable.

## What Changes

- Add a Bubble Tea TUI with a main menu, review session, browse, stats, and import/export flows.
- Implement core SRS scheduling and box movement rules with due-only review queues.
- Store cards and app state locally in JSON with crash-safe writes.
- Support CSV import/export with pipe-delimited tags.

## Capabilities

### New Capabilities
- `srs-core`: Card model, box movement, due calculation, and nextDue scheduling rules.
- `review-session`: Review flow UI that flips cards and records correct/incorrect outcomes.
- `card-storage`: JSON persistence for cards and app state with atomic writes.
- `csv-import-export`: CSV import/export for card content with tag parsing.
- `stats-screen`: Session stats display for reviewed/correct/incorrect counts.

### Modified Capabilities

## Impact

- New Go module with Bubble Tea app entry point and UI views.
- New domain and storage packages for scheduling and persistence logic.
- New CSV I/O utilities for import/export.
