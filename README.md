# Kanji Boxes

Bubble Tea TUI for a 6-box kanji spaced-repetition trainer.

Current review intervals by box:

- Box 1: 1 day
- Box 2: 3 days
- Box 3: 7 days
- Box 4: 14 days
- Box 5: 30 days
- Box 6: 60 days

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

Or build and run the binary directly:

```bash
make build
./kanji-box
```

## Install (Run From Anywhere)

```bash
make install
export PATH="$HOME/.local/bin:$PATH"
kanji-box
```

Add the PATH export to your shell profile (e.g. `~/.zshrc`) to make it permanent.

## Data Storage

By default, data is stored under your OS user config directory:

- macOS: `~/Library/Application Support/kanji-boxes`
- Linux: `~/.config/kanji-boxes`
- Windows: `%AppData%\kanji-boxes`

Override the data directory with:

```bash
KANJI_BOXES_DIR=/path/to/data kanji-box
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
