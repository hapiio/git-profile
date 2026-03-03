# git-profile

![git-profile banner](/docs/images/banner.png)
[![CI](https://github.com/hapiio/git-profile/actions/workflows/ci.yml/badge.svg)](https://github.com/hapiio/git-profile/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/hapiio/git-profile/graph/badge.svg)](https://codecov.io/gh/hapiio/git-profile)
[![Go Report Card](https://goreportcard.com/badge/github.com/hapiio/git-profile)](https://goreportcard.com/report/github.com/hapiio/git-profile)
[![GoDoc](https://pkg.go.dev/badge/github.com/hapiio/git-profile.svg)](https://pkg.go.dev/github.com/hapiio/git-profile)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/hapiio/git-profile)](https://github.com/hapiio/git-profile/releases/latest)

**Switch between multiple git identities with a single command.**

Do you juggle work, personal, and open-source git accounts? `git-profile` lets
you define named identity profiles and apply them per-repo or globally — so you
never accidentally push a personal commit with your work email again.

---

## Features

- **Named profiles** — store `user.name`, `user.email`, SSH key, and GPG signing per identity
- **Per-repo or global** — apply any profile locally or to `~/.gitconfig`
- **Auto-apply via git hooks** — `prepare-commit-msg` and `pre-push` hooks enforce the right identity before every commit and push
- **Interactive picker** — numbered menu when you can't remember the ID
- **Import existing identity** — bootstrap a profile from your current git config in one command
- **Edit, rename, remove** — full profile lifecycle management
- **Shell completions** — bash, zsh, and fish
- **Zero runtime dependencies** — single static binary, no runtime required

---

## Installation

### Homebrew (macOS / Linux)

```bash
brew install hapiio/tap/git-profile
```

### go install

```bash
go install github.com/hapiio/git-profile@latest
```

### Pre-built binaries

Download the latest binary for your platform from the
[Releases](https://github.com/hapiio/git-profile/releases/latest) page.

```bash
# macOS arm64
curl -L https://github.com/hapiio/git-profile/releases/latest/download/git-profile_darwin_arm64.tar.gz \
  | tar -xz && mv git-profile /usr/local/bin/

# Linux amd64
curl -L https://github.com/hapiio/git-profile/releases/latest/download/git-profile_linux_x86_64.tar.gz \
  | tar -xz && mv git-profile /usr/local/bin/
```

### Linux packages (.deb / .rpm)

`.deb` and `.rpm` packages are available on the
[Releases](https://github.com/hapiio/git-profile/releases/latest) page.

---

## Quick Start

```bash
# 1. Add your profiles
git-profile add --id work     --name "Jane Dev" --email jane@company.com
git-profile add --id personal --name "Jane Doe" --email jane@example.com \
                               --ssh-key ~/.ssh/id_ed25519_personal

# 2. Apply a profile to the current repo
git-profile use work

# 3. Or apply globally
git-profile use personal --global

# 4. See what's active
git-profile current

# 5. Install hooks to auto-apply on every commit/push
git-profile install-hooks
git-profile set-default work
```

---

## Usage

```text
git-profile <command> [flags]
```

### Command Reference

| Command              | Description                                                |
| -------------------- | ---------------------------------------------------------- |
| `add`                | Add a new identity profile                                 |
| `list`               | List all configured profiles                               |
| `use <id>`           | Apply a profile to this repo (or globally with `--global`) |
| `current`            | Show the active git identity in this repo                  |
| `choose`             | Interactively pick a profile from a numbered menu          |
| `edit <id>`          | Edit an existing profile interactively                     |
| `rename <old> <new>` | Rename a profile                                           |
| `remove <id...>`     | Remove one or more profiles                                |
| `import`             | Create a profile from the current git identity             |
| `set-default <id>`   | Set a default profile for this repo or globally            |
| `ensure`             | Apply the correct profile (used by git hooks)              |
| `install-hooks`      | Install git hooks to auto-apply profiles on commit/push    |
| `version`            | Print version information                                  |
| `completion`         | Generate shell completion scripts                          |

---

### `add` — Add a profile

```bash
# Minimal
git-profile add --id work --name "Jane Dev" --email jane@company.com

# With SSH key
git-profile add --id oss \
  --name "Jane Dev" \
  --email jane@oss.dev \
  --ssh-key ~/.ssh/id_ed25519_oss

# With GPG signing
git-profile add --id secure \
  --name "Jane Dev" \
  --email jane@company.com \
  --gpg-key "ABC123DEF456" \
  --sign-commits

# Overwrite an existing profile
git-profile add --id work --name "Jane Dev" --email jane@company.com --force

# Interactive (prompts for each field)
git-profile add
```

**Flags:**

| Flag             | Description                              |
| ---------------- | ---------------------------------------- |
| `--id`           | Profile identifier (required)            |
| `--name`         | `user.name` value                        |
| `--email`        | `user.email` value                       |
| `--ssh-key`      | Path to SSH private key (`~/` supported) |
| `--gpg-key`      | GPG key ID for commit signing            |
| `--sign-commits` | Enable `commit.gpgsign = true`           |
| `--force`        | Overwrite if profile ID already exists   |

---

### `list` — View all profiles

```bash
git-profile list
```

```text
  PROFILE     USER         EMAIL                  SSH KEY
  ──────────────────────────────────────────────────────────
  personal    Jane Doe     jane@example.com        (default)
● work        Jane Dev     jane@company.com        ~/.ssh/id_work
```

The `●` marker indicates the profile whose `user.name` and `user.email` match
the current repository's git identity.

---

### `use` — Apply a profile

```bash
git-profile use work                 # sets identity in .git/config
git-profile use personal --global    # writes to ~/.gitconfig
```

Sets the following git config keys (only those present in the profile):

| Profile field  | git config key    |
| -------------- | ----------------- |
| `git_user`     | `user.name`       |
| `git_email`    | `user.email`      |
| `ssh_key_path` | `core.sshCommand` |
| `gpg_key_id`   | `user.signingkey` |
| `sign_commits` | `commit.gpgsign`  |

---

### `current` — Show active identity

```bash
git-profile current
```

```text
Current git identity

  user.name   Jane Dev
  user.email  jane@company.com
  ssh-key     (default)
  gpg-key     ABC123DEF456
  gpg-sign    true

  Matched profile: work
```

---

### `import` — Bootstrap from existing git config

Already have git configured? Import it as a profile in one command:

```bash
git-profile import --id personal --global   # imports from ~/.gitconfig
git-profile import --id work                # imports from current repo's .git/config
```

Any `core.sshCommand` containing `-i /path/to/key` is automatically parsed and
stored as `ssh_key_path`.

---

### `edit` — Modify a profile

```bash
git-profile edit work
```

Opens a pre-filled interactive prompt. Press Enter to keep the current value.

---

### `rename` — Rename a profile

```bash
git-profile rename work work-backup
```

All profile fields are preserved. Any repos with `gitprofile.default = work`
must be updated manually:

```bash
git config gitprofile.default work-backup
```

---

### `remove` — Delete profiles

```bash
git-profile remove personal              # prompts for confirmation
git-profile remove work personal --yes   # skip confirmation
```

---

### `install-hooks` and `set-default` — Auto-apply on commit/push

```bash
# Step 1: install hooks in a repo
cd my-project
git-profile install-hooks

# Step 2: set the default profile for this repo
git-profile set-default work

# Set a global fallback for all repos
git-profile set-default personal --global
```

From now on, `git commit` and `git push` in this repo automatically call
`git-profile ensure` before running:

```text
ensure resolution order:
  1. Local gitprofile.default  (set by set-default in this repo)
  2. Global gitprofile.default (set by set-default --global)
  3. Interactive picker        (terminal only)
```

---

## GPG Commit Signing

`git-profile` manages all GPG signing settings — `user.signingkey` and
`commit.gpgsign` — so you never have to edit them manually.

### 1. Find your GPG key ID

```bash
gpg --list-secret-keys --keyid-format=long
```

```text
sec   ed25519/ABC123DEF456 2024-01-01 [SC]
      ABCDEF1234567890ABCDEF1234567890ABC123DE
uid   [ultimate] Jane Dev <jane@company.com>
```

The key ID is the part after `/` on the `sec` line — `ABC123DEF456` here.

### 2. Create a profile with GPG signing

```bash
git-profile add \
  --id work \
  --name "Jane Dev" \
  --email jane@company.com \
  --gpg-key "ABC123DEF456" \
  --sign-commits
```

### 3. Apply the profile

```bash
git-profile use work
```

This writes four git config keys at once:

```ini
[user]
  name       = Jane Dev
  email      = jane@company.com
  signingkey = ABC123DEF456
[commit]
  gpgsign    = true
```

### 4. Verify signing works

```bash
git commit --allow-empty -m "test gpg signing"
git log --show-signature -1
```

### Using different keys per identity

```bash
git-profile add --id work \
  --name "Jane Dev" --email jane@company.com \
  --gpg-key "WORK_KEY_ID" --sign-commits

git-profile add --id personal \
  --name "Jane Doe" --email jane@example.com \
  --gpg-key "PERSONAL_KEY_ID" --sign-commits

git-profile add --id oss \
  --name "Jane Doe" --email jane@oss.dev
  # no GPG — leave unsigned for OSS contributions

cd ~/work-project  && git-profile use work
cd ~/personal-blog && git-profile use personal
cd ~/oss-project   && git-profile use oss
```

### Disabling GPG signing for a repo

Applying a profile that has no `gpg_key_id` does **not** unset signing keys —
it leaves any existing `user.signingkey` and `commit.gpgsign` values in place.
To explicitly disable signing:

```bash
git config --unset user.signingkey
git config --unset commit.gpgsign
```

---

## Configuration

Profiles are stored at `$XDG_CONFIG_HOME/git-profile/config.json`
(defaults to `~/.config/git-profile/config.json`).

Override the path with `--config`:

```bash
git-profile --config ~/dotfiles/git-profile.json list
```

Example config file:

```json
{
  "version": 1,
  "profiles": {
    "work": {
      "id": "work",
      "git_user": "Jane Dev",
      "git_email": "jane@company.com",
      "ssh_key_path": "~/.ssh/id_ed25519_work",
      "gpg_key_id": "ABC123DEF456",
      "sign_commits": true
    },
    "personal": {
      "id": "personal",
      "git_user": "Jane Doe",
      "git_email": "jane@example.com",
      "ssh_key_path": "~/.ssh/id_ed25519_personal"
    },
    "oss": {
      "id": "oss",
      "git_user": "Jane Doe",
      "git_email": "jane@oss.dev"
    }
  }
}
```

---

## Shell Completions

```bash
# bash
git-profile completion bash > /etc/bash_completion.d/git-profile

# zsh
git-profile completion zsh > "${fpath[1]}/_git-profile"

# fish
git-profile completion fish > ~/.config/fish/completions/git-profile.fish
```

---

## Development

```bash
git clone https://github.com/hapiio/git-profile
cd git-profile

make build    # compile to ./git-profile
make test     # run tests with -race
make cover    # open coverage report in browser
make install  # install to $GOPATH/bin
make lint     # run golangci-lint
make snapshot # build release snapshot via GoReleaser
```

### Release (maintainers)

Releases are fully automated via GoReleaser and GitHub Actions.

```bash
git tag v1.2.3
git push origin v1.2.3
# Publishes binaries, Homebrew formula, .deb/.rpm, AUR package
```

---

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for setup
instructions, coding conventions, and the PR process.

---

## License

[MIT](LICENSE) © 2026 hapiio
