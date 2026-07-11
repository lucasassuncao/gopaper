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

## Automating changes (Task Scheduler)

gopaper is one-shot by design: run it and it changes the wallpaper once. To change on a
schedule or on session events, point Windows Task Scheduler at the executable.

**On an interval** — create a task with a *Daily* trigger, repeated every 30 minutes (or any
cadence), action: start `gopaper.exe`. Or from a terminal:

```pwsh
schtasks /Create /TN "gopaper" /TR "C:\path\to\gopaper.exe" /SC MINUTE /MO 30
```

**On unlock** — a fresh wallpaper every time you come back to the PC. In Task Scheduler,
create a task whose trigger is *On workstation unlock* (Trigger tab → Begin the task → "On
workstation unlock"), action: start `gopaper.exe`. There is no `schtasks` shorthand for the
unlock trigger; use the GUI or an XML task definition.

Tip: in the task's settings, enable "Run task as soon as possible after a scheduled start is
missed" so a sleeping PC catches up on wake.

**Lock screen wallpaper** is *not* supported by gopaper. Windows only exposes the lock screen
image to normal apps through the `PersonalizationCSP`/policy registry keys, which require
administrator elevation and lock the Settings > Personalization > Lock screen page while set.
If that trade-off is acceptable to you, set
`HKLM\SOFTWARE\Policies\Microsoft\Windows\Personalization\LockScreenImage` manually — gopaper
deliberately doesn't automate it.

## Where to go next

- [CONFIGURATION.md](CONFIGURATION.md) — full schema reference
- [COMMANDS.md](COMMANDS.md) — every command and flag
- [FILTERS.md](FILTERS.md) — narrowing which files a category picks from
- [EDIT.md](EDIT.md) — the interactive TUI editor
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) — common errors and fixes
