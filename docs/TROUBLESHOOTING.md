# Troubleshooting

## "configuration file not found"

`gopaper` looks for `gopaper.yaml` next to the executable, then in its `conf` subdirectory, unless you pass `-c`/`--config`. If neither exists, run:

```pwsh
gopaper init -i
```

or point explicitly at a file:

```pwsh
gopaper -c C:\path\to\gopaper.yaml
```

Check the exact search order and locations in [GETTING-STARTED.md](GETTING-STARTED.md#where-gopaper-looks-for-the-config-file).

## "enabled categories not found"

Every category in `gopaper.yaml` has `enabled: false`, or there are no categories at all. Enable at least one, or restrict to a specific one:

```pwsh
gopaper --category "MyCategory" --include-disabled
```

Run `gopaper validate` first — it catches an empty `categories` list before this ever comes up at runtime.

## "unknown category \"X\""

`--category` was given a name that doesn't match any `categories[].name` in the loaded config exactly (case-sensitive). Run without `--category` to confirm the real names, or check for a typo.

## "no supported image files found... matching the configured filter"

The category's `source` directory has no files with a supported extension (`.jpg`, `.jpeg`, `.png`, `.webp`), or its `filter` excludes everything present. Narrow down which:

```pwsh
gopaper validate --strict     # confirms source exists and is reachable
```

then check the filter against what's actually in the directory — see [FILTERS.md](FILTERS.md). A `filter.age`/`filter.size` bound that's too strict, or a `match.glob`/`match.regex` typo, are the usual causes.

## Log file isn't being written

`configuration.logging.output` must be `log`, `file`, or `both` — `console` (the default) never touches disk. When it is one of those, `configuration.logging.file` is required; `gopaper validate` reports this before you hit it at runtime. The parent directory is created automatically, but the process still needs write permission there.

## `~` in a path isn't resolving

A leading `~` or `~/` (also `~\` on Windows) in `configuration.logging.file`, `configuration.history.file`, or a category's `source` expands to the user's home directory automatically. Anything else — a bare `~username`, or `~` in the middle of a path — is left untouched. If a path still looks unexpanded, check for a typo in the prefix (it must be exactly `~` or start with `~/`/`~\`).

## `prev`/`next` say history is empty

Either no wallpaper has been set yet with the current history file, or `configuration.history.enabled: false` is set (which disables recording entirely — there's nothing to navigate). Check `configuration.history.file` if you've customized it; `prev`/`next` and the main run must agree on the same file to see the same history.

## `self-update` reports a checksum mismatch

The downloaded binary didn't match the checksum manifest published with the release — the update is aborted rather than installed. Retry (transient download corruption is the most common cause); if it persists, download the release manually from GitHub and compare checksums yourself before reporting it.

## Permission errors

Common on Windows binaries placed in `Program Files`, or on read-only mounts. Run from a directory you have write access to (for the log file, history file, and — during `self-update` — the binary's own directory), or elevate as needed.

## Still stuck?

- [`gopaper validate`](COMMANDS.md#gopaper-validate) catches most config mistakes with a specific field path and message.
- [`gopaper show-docs`](COMMANDS.md#gopaper-show-docs) prints the full schema reference, including allowed values and defaults, without leaving the terminal.
- Check the log file (if `configuration.logging.output` isn't `console`) for the exact error gopaper hit.
