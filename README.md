# Kanji Boxes

Bubble Tea TUI for a 5-box kanji spaced-repetition trainer.

## Requirements

- Go 1.21+

## Setup

```bash
go mod download
```

## Run

```bash
go run .
```

## Data Storage

By default, data is stored under your OS user config directory:

- macOS: `~/Library/Application Support/kanji-boxes`
- Linux: `~/.config/kanji-boxes`
- Windows: `%AppData%\kanji-boxes`

Override the data directory with:

```bash
KANJI_BOXES_DIR=/path/to/data go run .
```

Files:

- `cards.json`
- `state.json`

## Usage

### Main Menu

- `up/down`: move selection
- `enter`: select
- `q`: quit

### Review

- `space`: flip card
- `y`: mark correct
- `n`: mark incorrect
- `esc`: back to menu

Progress shows as reviewed/total due cards.

### Stats

Shows today's totals and a multi-day history.

### Browse

Shows a list of cards and basic details.

- `up/down`: move selection
- `e`: edit selected card
- `d`: delete selected card
- `esc`: back to menu

### Import / Export

- Import: appends cards from a CSV file
- Export: writes all cards to a CSV file

Follow the prompts to enter a file path.

## Tests

```bash
go test ./...
```
