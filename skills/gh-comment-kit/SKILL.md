---
name: gh-comment-kit
description: gh-comment-kit GitHub CLI extension for posting and managing trackable comments on GitHub pull requests, and listing comments on issues or pull requests. Use when posting review comments that should be tracked across runs (CI bots, scheduled jobs), updating or replacing previous comments by group, hiding/resolving outdated PR comments, or listing comments on an issue or PR. Posting/updating/hiding trackable comments is supported on pull requests only; the `list` command additionally accepts issues.
license: MIT
compatibility:
  - Requires gh CLI (https://cli.github.com) with gh-comment-kit extension installed (`gh extension install srz-zumix/gh-comment-kit`)
---

# gh-comment-kit

`gh-comment-kit` is a GitHub CLI extension for posting trackable comments on GitHub **pull requests**, and providing the latest updates by managing previously posted comments (update, delete, hide, resolve). The plain `list` command also works on issues.

Trackable comments embed metadata (a user-supplied **group identifier** passed via `--group`, default `gh-comment-kit`) so subsequent runs can find, update, hide, or resolve the previous comments in the same group without cluttering the conversation. Only comments posted through `gh comment-kit review comment` carry this metadata; regular GitHub comments are not tracked.

**Scope at a glance**:
- `gh comment-kit list` — issue **or** PR (read-only)
- `gh comment-kit review comment` / `review list` / `review hide` — **pull requests only**

## Prerequisites

```bash
# Install gh CLI
brew install gh          # macOS
# or: https://cli.github.com/

# Install gh-comment-kit extension
gh extension install srz-zumix/gh-comment-kit

# Authenticate
gh auth login

# Verify
gh comment-kit --version
```

## CLI Structure

```
gh comment-kit                  # Root command
├── list                        # List all comments on an issue or pull request
└── review                      # Manage trackable pull request comments
    ├── comment                 # Post (or update/delete/hide/resolve previous) review comment
    ├── list                    # List trackable comments posted by gh-comment-kit
    └── hide                    # Hide trackable comments by group
```

## Persistent Global Flags

| Flag | Description |
| --- | --- |
| `--read-only` | Prevent any write operations |
| `-L`, `--log-level` | Log level (debug, info, warn, error) |

Common per-command flags:

| Flag | Short | Description |
| --- | --- | --- |
| `--repo` | `-R` | Repository in the format `owner/repo` (defaults to the current repo) |
| `--json` | | Output as JSON with the specified fields |

`<target>` arguments accept an issue/PR number, URL, or branch name (the latter resolves to the open PR for that branch).

---

## Comment Commands

### `list` (alias: `ls`)

```bash
# List all comments on an issue or pull request
gh comment-kit list <target>

# Specify repository
gh comment-kit list <target> --repo owner/repo

# Resolve from a URL
gh comment-kit list https://github.com/owner/repo/pull/123

# Resolve from a branch name (open PR for that branch)
gh comment-kit list my-feature-branch

# JSON output
gh comment-kit list <target> --json id,user,body,createdAt
```

Lists every comment on the target issue or pull request (not limited to gh-comment-kit-tracked comments).

---

## Review Subcommands (`gh comment-kit review`)

Commands under `review` operate on pull request comments and use the `--group` identifier embedded in metadata to track related comments across runs.

### `review comment` (alias: `c`)

```bash
# Post a comment on a pull request
gh comment-kit review comment <target> --body "Build succeeded"

# Read body from a file
gh comment-kit review comment <target> --body-file report.md

# Read body from stdin
cat report.md | gh comment-kit review comment <target> --body-file -

# Use a custom group identifier (default: "gh-comment-kit")
gh comment-kit review comment <target> --group ci-build --body "..."

# Inline review comment on a specific line of a file
gh comment-kit review comment <target> \
  --path src/main.go --line 42 --body "Consider extracting this"

# Update the previous comment in the same group instead of posting a new one
gh comment-kit review comment <target> --update --body "Updated report"

# Delete previous comments in the group, then post a new one
gh comment-kit review comment <target> --delete --body "Latest run"

# Hide previous comments in the group with a reason, then post a new one
gh comment-kit review comment <target> --hide OUTDATED --body "Latest run"

# Resolve previous review threads in the group, then post a new one
gh comment-kit review comment <target> --resolve --body "Issues addressed"

# Truncate body if it exceeds GitHub's size limit (default: split into multiple)
gh comment-kit review comment <target> --body-file huge.md --truncate

# Dry run (print what would be posted)
gh comment-kit review comment <target> --body "Preview" --dryrun
```

`--update`, `--delete`, `--hide`, and `--resolve` are mutually exclusive. `--body` and `--body-file` are mutually exclusive.

Valid `--hide` reasons: `ABUSE`, `DUPLICATE`, `OFF_TOPIC`, `OUTDATED`, `RESOLVED`, `SPAM`.

If the body exceeds GitHub's 65,536-character limit, the comment is automatically split into multiple comments unless `--truncate` is set.

### `review list`

```bash
# List trackable comments posted by gh-comment-kit on a PR
gh comment-kit review list <target>

# Filter by group identifier (default: list all groups)
gh comment-kit review list <target> --group ci-build

# JSON output
gh comment-kit review list <target> --json id,group,body,url
```

Lists only comments that contain `gh-comment-kit` metadata. Use this to inspect what previous runs posted before deciding to update/hide/resolve.

### `review hide`

```bash
# Hide all gh-comment-kit comments in the default group
gh comment-kit review hide <target>

# Hide a specific group
gh comment-kit review hide <target> --group ci-build

# Specify reason (default: OUTDATED)
gh comment-kit review hide <target> --reason RESOLVED
```

Inline review comments are hidden via the GraphQL `minimizeComment` mutation. Issue-level PR comments that cannot be resolved as threads are also minimized with the supplied reason.

Valid reasons: `ABUSE`, `DUPLICATE`, `OFF_TOPIC`, `OUTDATED`, `RESOLVED`, `SPAM`.

---

## Common Workflows

### Post a CI status comment that updates on every run

```bash
gh comment-kit review comment "$PR" \
  --group ci-status \
  --update \
  --body-file ci-summary.md
```

The first run creates the comment; subsequent runs edit the same comment in place.

### Post a fresh report and hide previous ones

```bash
gh comment-kit review comment "$PR" \
  --group lint-report \
  --hide OUTDATED \
  --body-file lint.md
```

Old reports stay in the conversation history but are collapsed as outdated.

### Inline code review from a tool

```bash
gh comment-kit review comment "$PR" \
  --group static-analysis \
  --path src/foo.go --line 12 \
  --resolve \
  --body "Detected potential nil dereference"
```

Previous unresolved threads from the same tool are resolved, then a new inline comment is posted.

### Cleanup at the end of a workflow

```bash
gh comment-kit review hide "$PR" --group ci-status --reason RESOLVED
```

### Inspect tracked comments before acting

```bash
gh comment-kit review list "$PR" --group ci-status --json id,body,url
```

---

## Tips

- `<target>` accepts a number, a full URL (which also resolves the repository), or a branch name. When a URL is given, `--repo` is inferred from it.
- Use distinct `--group` values per tool/job so different sources do not interfere with each other (e.g. `ci-status`, `lint-report`, `coverage`).
- Use `--dryrun` to preview the body and resolved metadata before posting from a workflow.
- Combine with `--read-only` to verify a workflow path performs no writes.

## References

- Repository: https://github.com/srz-zumix/gh-comment-kit
- gh CLI manual: https://cli.github.com/manual/
