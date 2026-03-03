# Command Reference

```text
git-profile <command> [flags]
```

## Command overview

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

## `add` — Add a profile

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

## `list` — View all profiles

```bash
git-profile list
```

```text
  PROFILE     USER         EMAIL                  SSH KEY
  ──────────────────────────────────────────────────────────
  personal    Jane Doe     jane@example.com        (default)
● work        Jane Dev     jane@company.com        ~/.ssh/id_work
```

The `●` marker indicates the profile whose `user.name` and `user.email` match the current repository's git identity.

---

## `use` — Apply a profile

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

**Flags:**

| Flag       | Description                              |
| ---------- | ---------------------------------------- |
| `--global` | Write to `~/.gitconfig` instead of local |

---

## `current` — Show active identity

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

## `choose` — Interactive picker

```bash
git-profile choose
```

Opens a numbered menu listing all profiles. Type the number and press Enter to apply.

**Flags:**

| Flag       | Description                              |
| ---------- | ---------------------------------------- |
| `--global` | Apply selected profile to `~/.gitconfig` |

---

## `edit` — Modify a profile

```bash
git-profile edit work
```

Opens a pre-filled interactive prompt. Press Enter to keep the current value.

---

## `rename` — Rename a profile

```bash
git-profile rename work work-backup
```

All profile fields are preserved. Any repos with `gitprofile.default = work` must be updated manually:

```bash
git config gitprofile.default work-backup
```

---

## `remove` — Delete profiles

```bash
git-profile remove personal              # prompts for confirmation
git-profile remove work personal --yes   # skip confirmation
```

**Flags:**

| Flag    | Description                  |
| ------- | ---------------------------- |
| `--yes` | Skip the confirmation prompt |

---

## `import` — Bootstrap from existing config

Already have git configured? Import it as a profile in one command:

```bash
git-profile import --id personal --global   # imports from ~/.gitconfig
git-profile import --id work                # imports from current repo's .git/config
```

Any `core.sshCommand` containing `-i /path/to/key` is automatically parsed and stored as `ssh_key_path`.

**Flags:**

| Flag       | Description                                 |
| ---------- | ------------------------------------------- |
| `--id`     | Profile ID to create                        |
| `--global` | Import from `~/.gitconfig` instead of local |

---

## `set-default` — Set default profile

```bash
git-profile set-default work             # default for this repo
git-profile set-default personal --global  # fallback for all repos
```

**Flags:**

| Flag       | Description                              |
| ---------- | ---------------------------------------- |
| `--global` | Set as global fallback in `~/.gitconfig` |

---

## `ensure` — Apply the correct profile (hooks)

```bash
git-profile ensure
```

Resolution order:

1. Local `gitprofile.default` (set by `set-default` in this repo)
2. Global `gitprofile.default` (set by `set-default --global`)
3. Interactive picker (terminal only)

This command is called automatically by the installed git hooks.

---

## `install-hooks` — Install git hooks

```bash
git-profile install-hooks
```

Installs `prepare-commit-msg` and `pre-push` hooks into `.git/hooks/`. Each hook calls `git-profile ensure` before proceeding.

---

## `version` — Print version

```bash
git-profile version
```

---

## `completion` — Shell completions

```bash
git-profile completion bash
git-profile completion zsh
git-profile completion fish
```

See [Installation](../installation.md#shell-completions) for setup instructions.
