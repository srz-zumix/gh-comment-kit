# gh-comment-kit

A tool for posting trackable comments and providing the latest updates.

## Installation

```sh
gh extension install srz-zumix/gh-comment-kit
```

## Shell Completion

**Workaround Available!** While gh CLI doesn't natively support extension completion, we provide a patch script that enables it.

**Prerequisites:** Before setting up gh-comment-kit completion, ensure gh CLI completion is configured for your shell. See [gh completion documentation](https://cli.github.com/manual/gh_completion) for setup instructions.

For detailed installation instructions and setup for each shell, see the [Shell Completion Guide](https://github.com/srz-zumix/go-gh-extension/blob/main/docs/shell-completion.md).

## Commands

### list

#### List comments for an issue or pull request

```sh
gh comment-kit list <target> [--repo <owner/repo>] [--json <fields>]
```

List all comments for the specified issue or pull request.
`<target>` accepts an issue/PR number, URL, or branch name.

| Flag | Short | Default | Description |
| ---- | ----- | ------- | ----------- |
| `--repo` | `-R` | (current repo) | Repository in the format `owner/repo` |
| `--json` | | | Output as JSON with specified fields |

---

### review

#### Post a review comment to a pull request

```sh
gh comment-kit review comment <target> [flags]
```

Post a trackable review comment to the specified pull request.
`<target>` accepts a PR number, URL, or branch name.
The comment body is supplied via `--body` or `--body-file` (mutually exclusive).
Pass `-` to `--body-file` to read from stdin.

If the comment body exceeds GitHub's size limit (65,536 characters), it is automatically split into multiple comments. Use `--truncate` to truncate instead of splitting.

| Flag | Short | Default | Description |
| ---- | ----- | ------- | ----------- |
| `--body` | `-b` | `""` | Comment body text |
| `--body-file` | `-F` | `""` | Path to a file containing the comment body (`-` for stdin) |
| `--group` | `-g` | `"gh-comment-kit"` | Comment group identifier used to track related comments |
| `--path` | `-p` | `""` | File path to attach the review comment to |
| `--line` | `-l` | `0` | Line number to comment on (requires `--path`) |
| `--update` | | `false` | Update (edit) the last comment in the group instead of creating a new one |
| `--delete` | | `false` | Delete previous comments in the same group before posting |
| `--resolve` | | `false` | Resolve previous review comments in the same group |
| `--truncate` | | `false` | Truncate the comment body if it exceeds the size limit instead of splitting |
| `--dryrun` | `-n` | `false` | Print what would be posted without actually posting |
| `--repo` | `-R` | (current repo) | Repository in the format `owner/repo` |
| `--json` | | | Output as JSON with specified fields |

#### List trackable comments for a pull request

```sh
gh comment-kit review list <target> [--group <group>] [--repo <owner/repo>] [--json <fields>]
```

List comments that were posted with `gh comment-kit review comment` for the specified pull request.
`<target>` accepts a PR number, URL, or branch name.

| Flag | Short | Default | Description |
| ---- | ----- | ------- | ----------- |
| `--group` | `-g` | `""` (all groups) | Filter by comment group identifier |
| `--repo` | `-R` | (current repo) | Repository in the format `owner/repo` |
| `--json` | | | Output as JSON with specified fields |

#### Hide trackable comments on a pull request

```sh
gh comment-kit review hide <target> [--group <group>] [--reason <reason>] [--repo <owner/repo>]
```

Hide (minimize) pull request comments that contain gh-comment-kit metadata.
`<target>` accepts a PR number, URL, or branch name.
Use `--group` to target comments for a specific group; omit to match all groups.
Review comments (inline) are hidden via the GraphQL `minimizeComment` mutation.
Issue-level comments that cannot be resolved as threads are also hidden with the specified reason.

| Flag | Short | Default | Description |
| ---- | ----- | ------- | ----------- |
| `--group` | `-g` | `"gh-comment-kit"` | Comment group identifier to target |
| `--reason` | `-r` | `OUTDATED` | Reason for hiding: `ABUSE`, `DUPLICATE`, `OFF_TOPIC`, `OUTDATED`, `RESOLVED`, `SPAM` |
| `--repo` | `-R` | (current repo) | Repository in the format `owner/repo` |
