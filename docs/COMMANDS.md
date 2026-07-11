# Commands

## `gopaper`

Selects a random wallpaper from the eligible categories and sets it.

```pwsh
gopaper
gopaper -c C:\path\to\gopaper.yaml
gopaper --category "Wallhaven,Nature"
gopaper --category "Wallhaven" --include-disabled
```

| Flag | Description |
|---|---|
| `--config`, `-c` | Path to the configuration file (default: standard lookup). |
| `--category` | Comma-separated category names to restrict selection to (default: all enabled categories). |
| `--include-disabled` | Include disabled categories when selecting (works with `--category` or alone). |
| `--version`, `-v` | Print the gopaper version. |

Eligible set: with no `--category`, every category with `enabled: true`. With `--category`, exactly the named categories — an unknown name is an error, and a disabled one named explicitly is skipped with a warning unless `--include-disabled` is also set.

---

## `gopaper init`

Generates a `gopaper.yaml` at `<executable_dir>/conf/gopaper.yaml`.

```pwsh
gopaper init -i
gopaper init -t full
gopaper init -f
```

| Flag | Description |
|---|---|
| `--interactive`, `-i` | Prompt for logging settings and categories one at a time. |
| `--template`, `-t` | `basic` (default) or `full` — a ready-made multi-category example. |
| `--force`, `-f` | Overwrite an existing configuration file. |

---

## `gopaper edit`

Opens the interactive TUI editor. Full walkthrough in [EDIT.md](EDIT.md).

| Flag | Description |
|---|---|
| `--config`, `-c` | Load this config file (default: standard lookup). |
| `--output`, `-o` | Write to this file instead of the loaded config. |
| `--theme` | Theme name (default: `dark`). |
| `--list-themes` | List available themes and exit. |
| `--no-save-confirm` | Skip the save confirmation dialog. |
| `--no-delete-confirm` | Skip the block-delete confirmation dialog. |
| `--no-validate-on-save` | Allow saving with validation errors. |
| `--dump` | Record a session trace to a temp JSONL file (for bug reports). |
| `--dump-path` | Write the session trace to this file instead (implies `--dump`). |

---

## `gopaper validate`

Runs the same validators as `gopaper edit` without opening the TUI.

```pwsh
gopaper validate
gopaper validate -c ./gopaper.yaml -f json
gopaper validate --strict
gopaper validate --summary
```

| Flag | Description |
|---|---|
| `--config`, `-c` | Path to the configuration file to validate (default: standard lookup). |
| `--format`, `-f` | `pretty` (default), `plain`, or `json`. |
| `--summary` | Show only the error count, not individual violations. |
| `--strict` | Also verify that every category's `source` directory exists on disk. |

Exits non-zero when validation fails — safe to use in scripts.

---

## `gopaper show-docs`

Renders the configuration schema reference (descriptions, defaults, allowed values) directly in the terminal.

```pwsh
gopaper show-docs
gopaper show-docs --section history
gopaper show-docs --theme dracula
```

| Flag | Description |
|---|---|
| `--theme` | Theme name (default: `dark`). |
| `--list-themes` | List available themes and exit. |
| `--section` | Show only the docs for this topic (case-insensitive, partial match — e.g. `logging`, `history`, `filter`). |

---

## `gopaper prev` / `gopaper next`

Step backward/forward through recently applied wallpapers, without picking a new random one.

```pwsh
gopaper prev
gopaper next
```

No flags. Disabled entirely when `configuration.history.enabled: false` — there's nothing to navigate.

---

## `gopaper history`

Opens an interactive list of every wallpaper recorded in history (newest first). Arrow keys
(or `j`/`k`) navigate, `/` filters by filename or category, **Enter reapplies the selected
wallpaper** (with the configured transition; per-monitor entries are reapplied per monitor),
and `q` quits without changing anything. Reapplying also moves the history cursor, so a
subsequent `gopaper prev`/`next` continues from that entry.

```pwsh
gopaper history
gopaper history --category "Saltern Study"
```

| Flag | Description |
|---|---|
| `--category <name>` | Only show entries from this category (exact match). |

---

## `gopaper monitors`

Lists every connected monitor with the 1-based index gopaper uses in `configuration.behavior.monitor` (`monitor1`, `monitor2`, ...) and `categories[].monitor` — see [`behavior.monitor`](CONFIGURATION.md#behaviormonitor) for how those fields decide which screen a category lands on.

```pwsh
gopaper monitors
```

```
Index    | Name          | Position  | Size      | Device Path
monitor1 | BOE           | -1920,621 | 1920x1080 | \\?\DISPLAY#BOE07B6#4&3855a8a0&0&UID265988#{...}
monitor2 | ASUS VG32VQ1B | 0,0       | 2560x1440 | \\?\DISPLAY#AUS32E0#5&19f84e22&1&UID4354#{...}
```

No flags. Use this before writing a `monitor1`/`monitor2` pin so you know which physical
screen each index refers to:

- **Name** is the monitor's EDID-reported name (e.g. `ASUS VG32VQ1B`), read via WMI. Falls
  back to the 3-letter manufacturer PNP code (e.g. `BOE`) when the panel's EDID doesn't set a
  friendly name — common for laptop panels, which is also a hint by itself: a `BOE`/`AUO`/
  `CMN`/`LGD`-style 3-letter code with no product name usually is the built-in laptop screen.
- **Position/Size** is each monitor's desktop rectangle, same arrangement as Windows Display
  Settings — the monitor at `0,0` is the reference point; others are offset from it (e.g. a
  monitor at `1920,0` sits to its right, one at negative X sits to its left). Size tells you
  the resolution, which combined with Name is usually enough to tell two monitors apart even
  without recognizing the EDID name.
- **monitor1**, **monitor2**, etc. in this table's Index column are exactly the values
  `behavior.monitor: monitorN` and `categories[].monitor: N` expect.

Both Name and Position/Size are best-effort: on a WMI query failure, Name is blank (`-`) but
the rest of the table still renders normally.

---

## `gopaper self-update`

Downloads a release from GitHub and replaces the running binary. The old binary is kept as `gopaper.old` until the next run, and the downloaded binary's checksum is verified against the release's published manifest when one exists.

```pwsh
gopaper self-update
gopaper self-update --list
gopaper self-update --list --prerelease
gopaper self-update --version v1.2.0
```

| Flag | Description |
|---|---|
| `--repo` | GitHub repository in `owner/repo` format (default set at build time). |
| `--version` | Install this specific release tag instead of the latest. |
| `--list` | List available releases and exit. |
| `--prerelease` | Include rc/beta/alpha releases in `--list`, or as the latest target when no `--version` is given. |
| `--limit` | Maximum number of releases to show with `--list` (default `20`, max `100`). |
