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
