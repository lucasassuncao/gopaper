# Configuration Reference

`gopaper.yaml` has two top-level sections: `configuration` and `categories`. Run `gopaper show-docs` to browse this same reference — descriptions, defaults, and allowed values — generated live from the schema, directly in the terminal.

```yaml
configuration:
  logging:
    output: "both"          # console | log | file | both | none
    file: "C:\\logs\\gopaper.log"
    level: "info"            # trace | debug | info | warn | error | fatal
    show-caller: false
  history:
    limit: 50
    file: ""
    enabled: true

categories:
  - name: "default"
    source: "C:\\wallpapers"
    mode: "crop"             # crop | tile | stretch | span | fit | center
    enabled: true
    filter:                  # optional
      match:
        glob: "landscape_*"
      age:
        max: 720h
      size:
        min: "500KB"
```

Both `logging.file`, `history.file`, and every category's `source` accept a leading `~` (or `~\` on Windows), expanded to the user's home directory at load time.

---

## `configuration.logging`

| Field | Type | Required | Default | Notes |
|---|---|---|---|---|
| `output` | string | yes | `console` | One of `console`, `log`, `file`, `both`, `none`. |
| `file` | string | required if `output` is `log`, `file`, or `both` | — | Log destination. Parent directories are created automatically. |
| `level` | string | yes | `info` | One of `trace`, `debug`, `info`, `warn`, `error`, `fatal`. |
| `show-caller` | bool | no | `false` | Includes the source file/line in each log entry. |

`output: none` discards all log output.

## `configuration.history`

Controls the wallpaper history used by `gopaper prev`/`gopaper next`.

| Field | Type | Required | Default | Notes |
|---|---|---|---|---|
| `limit` | int | no | `50` | Maximum number of past wallpapers kept. Oldest entries are dropped once exceeded. |
| `file` | string | no | `<executable_dir>/history/gopaper.json` | Custom history file path. |
| `enabled` | bool | no | `true` | Set `false` to stop recording history entirely (`prev`/`next` then have nothing to navigate). |

## `categories[]`

Each entry is one wallpaper source.

| Field | Type | Required | Notes |
|---|---|---|---|
| `name` | string | yes, unique | Display name; must not repeat across categories. |
| `source` | string | yes | Directory scanned for images (`.jpg`, `.jpeg`, `.png`, `.webp`). Not recursive. |
| `mode` | string | yes | One of `crop`, `tile`, `stretch`, `span`, `fit`, `center`. |
| `enabled` | bool | no (default `true`) | Disabled categories are skipped unless selected explicitly with `--category --include-disabled`. |
| `filter` | object | no | Narrows which files in `source` are eligible — see [FILTERS.md](FILTERS.md). |

## Wallpaper modes

| Mode | Effect |
|---|---|
| `crop` | Fills the screen, cropping edges that don't fit the aspect ratio. |
| `tile` | Repeats the image at its native size. |
| `stretch` | Stretches to fill the screen, ignoring aspect ratio. |
| `span` | Stretches across all connected monitors. |
| `fit` | Scales to fit within the screen without cropping. |
| `center` | Centers the image at its native size, no scaling. |

## Validating a config file

```pwsh
gopaper validate                 # pretty output, standard lookup
gopaper validate -c ./gopaper.yaml -f json
gopaper validate --strict         # also check that every category's source exists on disk
```

See [COMMANDS.md](COMMANDS.md#gopaper-validate) for the full flag reference.
