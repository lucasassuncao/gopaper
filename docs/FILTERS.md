# Filters

Every category picks a random file from `source` that has a supported image extension (`.jpg`, `.jpeg`, `.png`, `.webp`). An optional `filter` narrows that further, so a category can, say, only pick recent screenshots or only large photos.

```yaml
categories:
  - name: "Recent Screenshots"
    source: "C:\\Screenshots"
    mode: "fit"
    enabled: true
    filter:
      match:
        glob: "screenshot_*"
      age:
        max: 720h        # last 30 days
      size:
        min: "200KB"
```

`match`, `age`, and `size` combine with **AND** semantics: a file must satisfy all three to be eligible. Omit any of them to leave that dimension unconstrained.

---

## `filter.match`

Matches by filename. `literal`, `regex`, and `glob` are mutually exclusive — set exactly one.

| Field | Meaning |
|---|---|
| `literal` | Exact filename match (whole name, including extension). |
| `regex` | RE2 regular expression tested against the filename. |
| `glob` | Wildcard pattern (`*`, `?`) tested against the filename. |
| `case-sensitive` | `false` by default — set `true` to stop lowercasing both sides before comparing. |

```yaml
filter:
  match:
    regex: '^\d{4}-\d{2}-\d{2}_'   # names starting with a date
```

```yaml
filter:
  match:
    glob: "wallpaper_*.jpg"
```

```yaml
filter:
  match:
    literal: "favorite.png"
    case-sensitive: true
```

## `filter.age`

Matches by how long ago the file was last modified. Both bounds are optional and accept Go duration syntax (`24h`, `720h`, `30m`).

```yaml
filter:
  age:
    min: 24h     # at least 1 day old
    max: 720h    # at most 30 days old
```

## `filter.size`

Matches by file size. Both bounds are optional and accept human-readable sizes: `KB`/`MB`/`GB`/`TB` (decimal, powers of 1000) or `KiB`/`MiB`/`GiB`/`TiB` (binary, powers of 1024).

```yaml
filter:
  size:
    min: "500KB"
    max: "20MB"
```

---

## Validation

Both `gopaper edit` and `gopaper validate` check filters before they can cause a runtime surprise:

- `filter.match.literal`/`regex`/`glob` are mutually exclusive.
- `filter.age.min`/`max` and `filter.size.min`/`max` must be ordered (`min` ≤ `max`).
- `regex`/`glob` must compile, and `size` strings must parse.

If a category's filter excludes every file in `source`, `gopaper` fails at wallpaper-change time with "no supported image files found... matching the configured filter" — check the filter against what's actually in the directory.

See [CONFIGURATION.md](CONFIGURATION.md) for the full category schema.
