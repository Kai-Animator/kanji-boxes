# Kanji Boxes (Bubble Tea)

A reviewed and slightly expanded plan for a TUI flashcard trainer using a 5‑box Leitner system for Japanese study.

---

## Goal

Build a terminal-based flashcard trainer for kanji study using Bubble Tea. Cards follow a 5‑box Leitner SRS where correct answers move cards up and incorrect answers move them down. Reviews are scheduled based on the card’s box.

---

## Core Requirements

- Bubble Tea TUI (single-screen flows).
- Front of card: **kanji** + optional **hiragana**.
- Back of card: **English meaning**.
- 5 Leitner boxes with day-based intervals.
- Review queue shows **only due cards**.
- Deterministic card movement on correct/incorrect answers.

---

## Scheduling Rules

- Box 1: every **1 day**
- Box 2: every **2 days**
- Box 3: every **3 days**
- Box 4: every **4 days**
- Box 5: every **5 days**

### Due Logic

- A card is **due** when `nextDue <= today`.
- `today` is computed using the **user’s local calendar date**.
- `nextDue` is stored as a **date (YYYY-MM-DD)** to avoid time-zone issues.

### Review Outcomes

- **Correct** → move up one box (max Box 5).
- **Incorrect** → move down one box (min Box 1).
- After movement, `nextDue = today + interval(newBox)`.

Optional (future): allow incorrect cards to be re-queued within the same session.

---

## Data Model

### Card

```go
id            string        // uuid
kanji         string
hiragana      *string       // optional
english       string
box           int           // 1–5
createdAt     date
lastReviewed  *date         // nil if never reviewed
nextDue       date
tags          []string
reviewCount   int
correctCount  int
suspended     bool
```

### App State

```go
cards         []Card
sessionStats  SessionStats
config        Config
```

---

## Persistence

- Local JSON storage.
- Recommended split:
  - `cards.json` – all card data
  - `state.json` – config and app/session metadata

- Writes use a temp file + rename to reduce corruption risk.

### Storage Location

- Default: OS user config directory (e.g. `~/.config/kanji-boxes/`).
- Optional override via CLI flag or environment variable.

---

## CSV Import / Export

### Columns

```csv
kanji,hiragana,english,tags
```

- `tags` are pipe-separated: `jlpt5|verbs|common`.
- Import behavior:
  - New cards start in **Box 1**.
  - `nextDue = today`.

- Duplicate handling: initially allow duplicates; future option to merge by content.

---

## User Flows

### Main Menu

- Review
- Add Card
- Browse
- Import / Export
- Stats
- Quit

---

### Review Session

- Pull due cards sorted by `nextDue`.
- Show card front → flip → mark correct/incorrect.
- Update:
  - box
  - `lastReviewed`
  - `nextDue`

- Display progress (e.g. `4 / 17 due`).

Key bindings (example):

- `space` – flip card
- `y` / `n` – correct / incorrect
- `q` – exit review

---

### Card Management

- Add, edit, delete cards.
- Soft delete via `suspended` flag.
- Optional toggle to show/hide hiragana during review.

---

### Browse

- List all cards.
- Filter by single tag (initially).
- View box, due date, and stats per card.

---

### Stats Screen

- Session stats:
  - reviewed
  - correct / incorrect

- Optional cumulative stats:
  - total reviews
  - accuracy
  - cards per box

---

## CLI / TUI Design Principles

- One screen per flow.
- Consistent key bindings across screens.
- Footer/status bar with hints and shortcuts.
- Deterministic, testable update logic separate from UI.

---

## Milestones

1. Project scaffold + Bubble Tea app skeleton.
2. Data model + JSON persistence.
3. Scheduling + box movement logic **with unit tests**.
4. Review flow UI.
5. Card CRUD + browse/filter.
6. CSV import/export.
7. Stats screen.

---

## Open Decisions (Documented)

- Dates stored as **local calendar dates**, not timestamps.
- Incorrect answers move down **one box only**.
- New cards are immediately due.
- Tag filtering starts simple (single tag).

---

This structure keeps the SRS logic stable, testable, and predictable while letting the Bubble Tea UI evolve independently.
