# Contributing to git-profile

Thank you for taking the time to contribute! This document covers everything you
need to get started — from reporting a bug to shipping a new feature.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Features](#suggesting-features)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Tests](#tests)
- [Commit Convention](#commit-convention)
- [Pull Request Process](#pull-request-process)
- [Release Process](#release-process)
- [Secret Configuration (maintainers)](#secret-configuration-maintainers)

---

## Code of Conduct

Be respectful and constructive. We follow the
[Contributor Covenant](https://www.contributor-covenant.org/version/2/1/code_of_conduct/).

---

## Reporting Bugs

1. Search [existing issues](https://github.com/hapiio/git-profile/issues) first.
2. If not found, open a new issue and include:
   - Your OS and architecture (`git-profile version`)
   - The exact command you ran
   - Expected vs actual behaviour
   - Any relevant error output

---

## Suggesting Features

Open an issue with the `enhancement` label. Describe:

- The problem you are trying to solve
- Your proposed solution
- Alternatives you considered

For significant changes, discuss first before writing code — it avoids wasted
effort if the direction doesn't align with the project goals.

---

## Development Setup

**Prerequisites:** Go 1.22+, git

```bash
git clone https://github.com/hapiio/git-profile
cd git-profile
go mod download

# Build
make build

# Install locally
make install   # installs to $GOPATH/bin
```

### Useful make targets

| Target | Description |
|---|---|
| `make build` | Compile `./git-profile` |
| `make test` | Run all tests with `-race` |
| `make cover` | Run tests and open HTML coverage report |
| `make lint` | Run `golangci-lint` (requires installation) |
| `make snapshot` | Build a GoReleaser snapshot locally |
| `make clean` | Remove build artefacts |

### Installing golangci-lint

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

---

## Making Changes

### Project structure

```
git-profile/
├── cmd/            # One file per cobra command + shared helpers
├── internal/
│   ├── config/     # Config file load/save (Manager, Profile, Config)
│   ├── git/        # Thin wrappers over the git binary
│   └── ui/         # Lipgloss styles, Input/Confirm/Select prompts
├── main.go         # Entry point — calls cmd.Execute()
```

### Guidelines

- **One concern per package.** `internal/git` never imports `internal/config`,
  and `internal/ui` never imports either. Commands in `cmd/` are the only place
  where all three meet.
- **No interactive prompts in non-TTY contexts.** `ui.IsTTY()` guards all
  `Input`, `Confirm`, and `Select` calls. Never add a blocking read without
  checking `IsTTY()` first.
- **Atomic config writes.** Config is always written via a temp-file rename.
  Do not use `os.WriteFile` directly on the config path.
- **Minimal dependencies.** The project intentionally keeps its dependency tree
  small (`cobra` + `lipgloss` only). Discuss in an issue before adding a new
  module.

---

## Tests

```bash
make test
```

All tests use `t.TempDir()` for isolation — no test should read or write the
real `~/.config/git-profile/config.json`.

### Writing new tests

- **`internal/config`** — use `config.NewManager(path)` with a temp path.
- **`internal/git`** — use the `initRepo(t)` helper to create a real git repo
  in a temp dir, and `chdir(t, dir)` to switch into it.
- **`cmd`** — use `cmd.RunArgs([]string{"--config", tmpPath, ...})` and assert
  on the resulting config file contents, not on stdout.

### Coverage

After running `make test`, open the coverage report:

```bash
make cover
```

New code should maintain or improve the existing coverage level. CI enforces a
60 % overall floor and 50 % per-patch minimum via Codecov.

---

## Commit Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short summary>

[optional body]

[optional footer]
```

| Type | When to use |
|---|---|
| `feat` | New feature visible to users |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `test` | Adding or fixing tests |
| `refactor` | Code restructuring with no behaviour change |
| `chore` | Dependency bumps, CI tweaks, etc. |
| `ci` | Changes to GitHub Actions workflows |

**Examples:**

```
feat(cmd): add `import` command to bootstrap profiles from git config
fix(config): ensure config directory has 0700 permissions on creation
docs: document shell completion setup in README
```

GoReleaser uses commit messages to auto-generate the changelog, so following
this convention directly improves release notes.

---

## Pull Request Process

1. Fork the repo and create a branch from `main`:
   ```bash
   git checkout -b feat/my-feature
   ```
2. Make your changes and write tests.
3. Run the full test suite locally:
   ```bash
   make test && make lint
   ```
4. Push and open a Pull Request against `main`.
5. Fill in the PR template — describe what changed and why.
6. A maintainer will review and may request changes.
7. Once approved and CI is green, the PR is merged by a maintainer.

### PR checklist

- [ ] Tests added or updated
- [ ] `make test` passes
- [ ] `make lint` passes (no new warnings)
- [ ] Commit messages follow the convention above
- [ ] Documentation updated if behaviour changed

---

## Release Process

Releases are fully automated via GoReleaser and GitHub Actions.

```bash
# Maintainers only:
git tag v1.2.3
git push origin v1.2.3
```

This triggers `.github/workflows/release.yml`, which:

1. Runs the full test suite
2. Builds binaries for all platforms (linux/amd64, linux/arm64, linux/armv7,
   darwin/amd64, darwin/arm64, windows/amd64)
3. Creates GitHub Release with checksums
4. Publishes archives (`.tar.gz` / `.zip`)
5. Publishes `.deb` and `.rpm` packages
6. Updates the Homebrew formula (macOS + Linux)
7. Updates the Scoop bucket (Windows) — if `SCOOP_BUCKET_TOKEN` is set
8. Publishes to AUR (Arch Linux) — if `AUR_KEY` is set

---

## Secret Configuration (maintainers)

Add these under **Settings → Secrets and variables → Actions**:

| Secret | Required | Description |
|---|---|---|
| `HOMEBREW_TAP_TOKEN` | Recommended | PAT with `repo` write access to `hapiio/homebrew-tap` |
| `SCOOP_BUCKET_TOKEN` | Optional | PAT with `repo` write access to `hapiio/scoop-bucket` |
| `AUR_KEY` | Optional | Base64-encoded ed25519 private key registered on AUR |
| `CODECOV_TOKEN` | Optional | Codecov upload token for coverage reporting |

Channels without a configured secret are skipped automatically — the release
still completes for all other channels.

### Creating the Homebrew tap

```bash
# Create a new repo named homebrew-tap under the hapiio org, then:
mkdir -p Formula
# GoReleaser will push the formula here on every release tag.
```

### Generating an AUR key

```bash
ssh-keygen -t ed25519 -C "git-profile AUR" -f aur_key -N ""
# Register aur_key.pub on https://aur.archlinux.org/account/
# Base64-encode the private key for the secret:
base64 -w0 aur_key
```
