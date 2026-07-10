# Interactive Config Editor

`gopaper edit` opens a two-panel TUI editor for your configuration file. It is the fastest way to create or modify a `gopaper.yaml` without leaving the terminal, with inline hints, presets, and validation on every save.

```bash
gopaper edit
```

For a read-only reference of the same field descriptions, defaults, and constraints — without opening the editor — run `gopaper show-docs`.

---

## Layout

**Left panel — block list**\
Lists the top-level blocks in your config: `configuration` (with its `logging` and `history` sub-sections), and each of your categories. Navigate with **↑ / ↓**, open a block with **Enter**.

**Right panel — field editor**\
Shows the fields of the selected block. Each field has a type-appropriate control: text inputs for strings, toggles for booleans, dropdowns for enums (like `mode` and `output`). Inline hints show the field description, allowed values, and defaults.

---

## Keybindings

| Key | Action |
|---|---|
| **↑ / ↓** | Move between items |
| **Enter** | Open a block / confirm a value |
| **Esc** | Go back / cancel |
| **Tab** | Next field |
| **Ctrl+P** | Insert a whole-document template (root list) |
| **Ctrl+S** | Save |
| **Ctrl+U** | Undo last edit |
| **Ctrl+Y** | Redo |

---

## Saving and validation

**Ctrl+S** validates the entire config before writing. The editor enforces:

- `configuration.logging.output`, `configuration.logging.level`, and each category's `mode` must be one of their allowed values.
- `configuration.logging.file` is required when `logging.output` is `log`, `file`, or `both`.
- Category `name` values must be unique.
- Within a category's `filter.match`, `literal`/`regex`/`glob` are mutually exclusive.
- `filter.age.min`/`max` and `filter.size.min`/`max` must be ordered, and must compile/parse (bad regex, glob, or size string is caught here). See [FILTERS.md](FILTERS.md) for the full filter reference.

If there are errors, the editor shows them inline and refuses to save. Use `--no-validate-on-save` to override (a warning is shown, the file is still written). Run the same checks outside the editor with `gopaper validate` — see the [README](../README.md#validating-the-configuration).

---

## Presets

Press the preset picker inside a block editor to insert a ready-made `configuration` or `categories` entry — for example, one of the six wallpaper modes (`crop`, `tile`, `stretch`, `span`, `fit`, `center`), a category with a `filter` already set (`with-filter-match-glob`, `with-filter-match-regex`, `with-filter-age`, `with-filter-size`), or a logging setup (`console-info`, `file`, `both`, ...). Press **Ctrl+P** on the root list to start from a whole-document template instead (`single-category`, `multi-category`, `file-logging`, `console-and-file`, `history-limited`).

---

## Creating a new config file

Use `--output` to write to a different file than the one loaded:

```bash
gopaper edit --output ~/gopaper/gopaper.yaml
```

Useful for bootstrapping a new config from an existing template.

---

## Themes

```bash
gopaper edit --theme dracula
gopaper edit --list-themes   # see all available themes
```

The default theme is `dark`.

---

## Flags

| Flag | Description |
|---|---|
| `--config`, `-c` | Load this config file (default: standard lookup) |
| `--output`, `-o` | Write to this file instead of the loaded config |
| `--theme` | Theme name (default: `dark`) |
| `--list-themes` | List available themes and exit |
| `--no-save-confirm` | Skip the save confirmation dialog |
| `--no-delete-confirm` | Skip the block-delete confirmation dialog |
| `--no-validate-on-save` | Allow saving with validation errors |
| `--dump` | Record a session trace to a temp JSONL file (for bug reports) |
| `--dump-path` | Write the session trace to this file instead (implies `--dump`) |

When `--config` is not set, the editor looks for `gopaper.yaml` next to the executable and in its `conf` subdirectory — the same locations `gopaper` searches when running normally. If neither exists, it falls back to `<executable_dir>/conf/gopaper.yaml`, matching where `gopaper init` writes the file.

See [COMMANDS.md](COMMANDS.md#gopaper-edit) for this same table alongside every other command's flags.
