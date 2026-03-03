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
- **Interactive picker** — fuzzy-free numbered menu when you can't remember the ID
- **Import existing identity** — bootstrap a profile from your current git config in one command
- **Edit, rename, remove** — full profile lifecycle management
- **Shell completions** — bash, zsh, fish, and PowerShell
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
# macOS arm64 example
curl -L https://github.com/hapiio/git-profile/releases/latest/download/git-profile_darwin_arm64.tar.gz \
  | tar -xz && mv git-profile /usr/local/bin/
```

---

## Quick Start

```bash
# 1. Add your profiles
git-profile add --id work     --name "Jane Dev"  --email jane@company.com
git-profile add --id personal --name "Jane Doe"  --email jane@example.com \
                               --ssh-key ~/.ssh/id_ed25519_personal

# 2. Apply a profile to the current repo
git-profile use work

# 3. Or apply globally
git-profile use personal --global

# 4. See what's active
git-profile current
```

---

## Usage

```sh
git-profile <command> [flags]
```

### Commands

| Command              | Description                                                |
| -------------------- | ---------------------------------------------------------- |
| `add`                | Add a new identity profile                                 |
| `list`               | List all configured profiles                               |
| `use <id>`           | Apply a profile to this repo (or globally with `--global`) |
| `current`            | Show the active git identity in this repo                  |
| `choose`             | Interactively pick a profile from a numbered menu          |
| `edit <id>`          | Edit an existing profile interactively                     |
| `rename <old> <new>` | Rename a profile                                           |
| `remove <id>`        | Remove a profile                                           |
| `import`             | Create a profile from the current git identity             |
| `set-default <id>`   | Set a default profile for this repo or globally            |
| `ensure`             | Apply the correct profile (used by git hooks)              |
| `install-hooks`      | Install git hooks to auto-apply profiles                   |
| `version`            | Print version information                                  |
| `completion`         | Generate shell completion scripts                          |

---

### `add` — Add a profile

```bash
# With flags (scriptable)
git-profile add --id work --name "Jane Dev" --email jane@company.com
git-profile add --id oss  --name "Jane Dev" --email jane@oss.dev \
                           --ssh-key ~/.ssh/id_oss \
                           --gpg-key ABC123DEF --sign-commits

# Interactive (run without flags)
git-profile add
```

### `list` — View all profiles

```sh
  PROFILE     USER         EMAIL                  SSH KEY
  ──────────────────────────────────────────────────────────
  personal    Jane Doe     jane@example.com        (default)
● work        Jane Dev     jane@company.com        ~/.ssh/id_work
```

The `●` marker indicates the profile matching the current repo's git identity.

### `use` — Apply a profile

```bash
git-profile use work            # applies to current repo only
git-profile use personal --global  # overwrites ~/.gitconfig
```

### `import` — Bootstrap from existing git config

Already have git configured? Import it as a profile in one command:

```bash
git-profile import --id personal --global   # from ~/.gitconfig
git-profile import --id work                # from current repo's .git/config
```

### `install-hooks` — Auto-apply on commit/push

```bash
cd my-repo
git-profile install-hooks
git-profile set-default work   # apply "work" automatically in this repo
```

From now on, `git commit` and `git push` in this repo will automatically ensure
the correct identity is active. If no default is set, you'll be prompted
interactively.

### `ensure` — Used by hooks

```bash
git-profile ensure
```

Resolution order:

1. Local `gitprofile.default` (set by `set-default`)
2. Global `gitprofile.default`
3. Interactive picker (TTY only)

---

## Configuration

Profiles are stored in a JSON file at:

| Platform      | Path                                                                                         |
| ------------- | -------------------------------------------------------------------------------------------- |
| macOS / Linux | `$XDG_CONFIG_HOME/git-profile/config.json` (defaults to `~/.config/git-profile/config.json`) |
| Windows       | `%AppData%\git-profile\config.json`                                                          |

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
      "ssh_key_path": "/home/jane/.ssh/id_work"
    },
    "personal": {
      "id": "personal",
      "git_user": "Jane Doe",
      "git_email": "jane@example.com",
      "gpg_key_id": "ABC123DEF456",
      "sign_commits": true
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

# PowerShell
git-profile completion powershell | Out-String | Invoke-Expression
```

---

## GPG Commit Signing

```bash
# Add a profile with GPG signing
git-profile add \
  --id work \
  --name "Jane Dev" \
  --email jane@company.com \
  --gpg-key "YOUR_GPG_KEY_ID" \
  --sign-commits

# Apply it — git-profile also sets user.signingkey and commit.gpgsign
git-profile use work
```

---

## Development

```bash
git clone https://github.com/hapiio/git-profile
cd git-profile

make build    # compile to ./git-profile
make test     # run tests with -race
make cover    # open coverage report
make install  # install to $GOPATH/bin
make lint     # run golangci-lint (requires installation)
```

### Release (maintainers)

Releases are fully automated via GoReleaser and GitHub Actions.

```bash
git tag v1.2.3
git push origin v1.2.3
# GitHub Actions publishes binaries, Homebrew formula, .deb/.rpm packages
```

---

## Contributing

Contributions are welcome! Please:

1. Fork the repo and create a feature branch
2. Add tests for new functionality
3. Run `make test` and `make lint` before submitting
4. Open a pull request with a clear description

See [CONTRIBUTING.md](CONTRIBUTING.md) if present, or open an issue to discuss
larger changes first.

---

## License

[MIT](LICENSE) © 2026 hapiio
