# Getting Started

`gopaper` picks a random wallpaper from directories you define ("categories") and sets it as your desktop background.

## Install

Build from source (requires Go):

```pwsh
go build -o gopaper .
```

Or grab a prebuilt binary from the project's GitHub releases and put it in a folder on your `PATH`.

## Create a configuration file

```pwsh
gopaper init -i          # interactive prompts
gopaper init -t full     # a ready-made multi-category template
gopaper init             # basic template, no prompts
```

This writes `gopaper.yaml` to `<executable_dir>/conf/gopaper.yaml`. Use `-f`/`--force` to overwrite an existing file.

Prefer a visual editor? Skip `init` and go straight to:

```pwsh
gopaper edit
```

which opens a TUI with inline field hints, presets, and validation — see [EDIT.md](EDIT.md).

## Where gopaper looks for the config file

Unless you pass `-c`/`--config`, gopaper searches, in order:

1. `<executable_dir>/gopaper.yaml`
2. `<executable_dir>/conf/gopaper.yaml`

If neither exists, most commands fail with a message pointing you to `gopaper init`.

## First run

```pwsh
gopaper
```

This selects a random enabled category, then a random image inside it (filtered by its `filter` if one is set — see [FILTERS.md](FILTERS.md)), and sets it as the wallpaper in the mode configured for that category. The change is logged per `configuration.logging` and, unless disabled, recorded to history.

Undo/redo through recent wallpapers:

```pwsh
gopaper prev
gopaper next
```

Check everything is well-formed before or after editing by hand:

```pwsh
gopaper validate
```

## Where to go next

- [CONFIGURATION.md](CONFIGURATION.md) — full schema reference
- [COMMANDS.md](COMMANDS.md) — every command and flag
- [FILTERS.md](FILTERS.md) — narrowing which files a category picks from
- [EDIT.md](EDIT.md) — the interactive TUI editor
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) — common errors and fixes
