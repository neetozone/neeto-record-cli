---
name: neetorecord
description: >
  Manage NeetoRecord from the command line.
  Use when the user asks about operations exposed by the NeetoRecord CLI.
---

## Prerequisites

Run `neetorecord doctor` to check authentication and connectivity.
If not authenticated, run `neetorecord login`.

## Authentication & multi-subdomain

Credentials for every logged-in subdomain are stored together in
`~/.config/neetorecord/auth.json`. A command that talks to the API picks which
subdomain to use by these rules:

- 0 subdomains logged in → every credential-using command errors with
  "not logged in. Run 'neetorecord login' to authenticate".
- 1 subdomain logged in → that one is the implicit default; `--subdomain`
  may be omitted.
- 2+ subdomains logged in → **`--subdomain <name>` is required** on every
  credential-using command, including `doctor`. The error lists every
  logged-in subdomain so the agent can offer a choice.

`login` / `logout` / `whoami` have dedicated behavior:

| Command | Behavior |
|---|---|
| `neetorecord login --subdomain <name>` | Adds or refreshes the entry for `<name>`. No flag → prompts for the subdomain. |
| `neetorecord logout --subdomain <name>` | Removes that one entry. |
| `neetorecord logout --all` | Removes every entry. |
| `neetorecord logout` (no flag) | Removes the only entry if exactly one is logged in; errors if multiple. |
| `neetorecord whoami` | Lists every logged-in account. Marks the entry `(default)` when exactly one. |
| `neetorecord whoami --subdomain <name>` | Shows just that one. |

## Global flags (persistent on every command)

| Flag | Purpose |
|---|---|
| `--subdomain <name>` | Select which logged-in subdomain the command targets. Required when multiple are logged in. |
| `--json` | Force JSON envelope output even on a TTY. |
| `--quiet` | Emit only the raw payload — no envelope, no breadcrumbs. For action commands (create/update), emits just the resource identifier; `delete` emits `success`. Designed for scripting. |
| `--toon` | Emit TOON (Token Optimized Output Notation). Preferred for feeding list/show output back to an LLM; ~30–60% fewer tokens than JSON. |

Precedence if multiple are set: `--toon` > `--quiet` > `--json` > pretty.

## Output modes & response envelope

**Pretty (default on a TTY)** — tables for arrays, key-value for objects,
breadcrumbs appended. Not intended for machine consumption.

**JSON envelope** (non-TTY, or `--json`):
```json
{
  "data": <resource body>,
  "breadcrumbs": [{ "label": "List", "command": "neetorecord <resource> list" }],
  "pagination": {
    "current_page_number": 1,
    "total_pages": 10,
    "total_records": 250
  }
}
```
`breadcrumbs` is omitted when empty. `pagination` is present only for list
commands.

**Quiet** (`--quiet`) — `data` contents only, no envelope. For action
commands `PrintQuiet` unwraps a single-key wrapper and prints the first of
`sid` / `id` / `name`. For `delete` it prints `success`.

**TOON** (`--toon`) — same data as JSON, re-encoded into TOON. Shape is
equivalent but whitespace/keys are compressed. Parse by re-reading keys as
you would JSON.

### Pagination

List commands accept `--page` (1-indexed) and `--page-size` (max 100).
The envelope's `pagination` field always exposes:
`current_page_number`, `total_pages`, `total_records`. Agents should loop
by incrementing `--page` until `current_page_number == total_pages`.

## Discovery

The full, always-accurate command tree (including any flags added after
this skill was built) is available as JSON:

```bash
neetorecord commands
```

Each catalog entry has `command`, `description`, optional `flags` (with
`name`, `type`, `default`, `description`, `required`), and `subcommands`.
Use this whenever a user asks about a flag or command not covered below.

## Diagnostics & IDE setup

| Command | Purpose |
|---|---|
| `doctor` | Auth check + API reachability + version. Uses `--subdomain` when multiple are logged in. |
| `version` | Print CLI version / commit / build date. |
| `commands` | Emit the full command/flag catalog as JSON. |
| `setup claude` | Install NeetoRecord plugin into Claude Code (`plugin.json`, hooks, this SKILL.md). |
| `setup cursor` / `windsurf` / `copilot` / `gemini` / `codex` | Write NeetoRecord rule files into the current project directory; re-run after an upgrade to refresh them. |

## Environment variable override

Set `NEETORECORD_BASE_URL` to point the CLI at a staging or local server:

```bash
export NEETORECORD_BASE_URL=http://acme.lvh.me:8980
neetorecord login --subdomain acme
```

## Error surface

Every command exits non-zero on failure and writes a single-line message to
stderr. Common errors the agent should expect:

- `not logged in. Run 'neetorecord login' to authenticate` — empty credential store.
- `multiple subdomains logged in (acme, beta); specify --subdomain` — pick one.
- `not logged in to "foo". Logged in subdomains: acme, beta` — bad `--subdomain`.
- `required flag(s) "xxx" not set` (from cobra) — missing required flag.
- API errors come through with the server's message body; inspect the
  JSON envelope (or the `--quiet` payload) for `error` / `errors` / `notice`
  keys and any suggestions the API returns.

## Product-specific commands

Every resource below takes the global flags above. Positional `<id>` arguments
accept either the recording's UUID or its `public_link_id` (the short code in a
watch URL). Run `neetorecord commands` for the authoritative tree, including any
flags added after this file was written.

| Resource | Commands |
|---|---|
| `recordings` | `list`, `show <id>`, `update <id>`, `delete <id>`, `search`, `search-by-transcript` |
| `recordings` (transcript) | `transcript <id>`, `transcript-status <id>`, `trigger-transcript <id>` |
| `recordings` (chapters) | `chapters <id>`, `chapter-status <id>`, `trigger-chapters <id>` |
| `recordings` (sharing) | `share-link <id>`, `embed-code <id>`, `download-url <id>`, `trigger-mp4 <id>`, `screenshot <id>` |
| `recordings` (CTAs) | `ctas <id>`, `create-cta <id>` |
| `recordings` (analytics) | `analytics <id>` |
| `folders` | `list`, `create` |
| `tags` | `list` |
| `team-members` | `list`, `show <id>`, `create`, `update <id>`, `delete <id>` |
| `recording-requests` | `create` |
| `analytics` | `show` |

Notes that matter when driving these:

- `recordings update --tag` **replaces** every tag on the recording. To add one
  without losing the others, run `recordings show <id>` first, read its `tags`
  array, and pass the existing names plus the new one. Passing only the new tag
  deletes the rest. Tag names that do not exist yet are created.
- `recordings update --folder-id ""` removes the recording from its folder.
  `--folder-id` accepts the id returned by `folders list`.
- `search` matches titles; `search-by-transcript` matches what was said. Both
  require `--query` and support pagination.
- `analytics show` is workspace-wide and takes no id; it accepts optional
  `--from-date` / `--to-date`. For one recording use `recordings analytics <id>`.
- `delete` is immediate and irreversible — there is no confirmation prompt, and
  `recordings delete` takes the transcript, chapters and CTAs with it. Show the
  user the title and id you are about to delete and get an explicit yes first.

### Waiting for generation

`trigger-transcript`, `trigger-chapters` and `trigger-mp4` return as soon as the
job is queued, so never treat the response as completion.

Transcripts and chapters have status commands. Poll `transcript-status <id>` or
`chapter-status <id>` and stop on a terminal state:

| Command | `status` values |
|---|---|
| `transcript-status` | `not_applicable`, `unexecuted`, `in_progress`, `failed`, `success` |
| `chapter-status` | `unexecuted`, `in_progress`, `failed`, `success` |

`success` and `failed` are terminal — stop polling and report a failure rather
than retrying forever. `in_progress` means keep waiting; `unexecuted` means
nothing has been triggered yet. These commands also return `has_transcript` /
`has_chapters`, which is the reliable check for whether content exists.

**There is no `mp4-status` command.** MP4 readiness surfaces through the
commands that need it: `download-url` returns `is_download_file_ready` and
`screenshot` returns `is_screenshot_ready`, both `false` while the file is
stale. Re-running `trigger-mp4` also reports where it stands — `ready` means the
MP4 is current, `generating` means it was just queued, `already_generating`
means a conversion is under way. So the loop is: call `download-url` or
`screenshot`, and if the readiness flag is `false`, run `trigger-mp4` and retry
until it flips.

- `screenshot` needs `--timestamp` within the recording's duration, and an up to
  date MP4.
- `download-url` and `screenshot` return short lived presigned URLs; fetch them
  immediately before use rather than caching them.

Full reference: https://apidocs.neetorecord.com/cli-reference/overview
