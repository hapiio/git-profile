# Configuration

## Config file location

Profiles are stored at:

```text
$XDG_CONFIG_HOME/git-profile/config.json
```

Which defaults to `~/.config/git-profile/config.json` if `XDG_CONFIG_HOME` is not set.

## Custom config path

Use `--config` to point at a different file:

```bash
git-profile --config ~/dotfiles/git-profile.json list
```

## Config file format

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

## Profile fields

| Field          | Type   | Description                               |
| -------------- | ------ | ----------------------------------------- |
| `id`           | string | Unique profile identifier                 |
| `git_user`     | string | `user.name` written to git config         |
| `git_email`    | string | `user.email` written to git config        |
| `ssh_key_path` | string | Path to SSH private key (`~/` supported)  |
| `gpg_key_id`   | string | GPG key ID written to `user.signingkey`   |
| `sign_commits` | bool   | Sets `commit.gpgsign = true` when applied |

## How git config keys are mapped

When you run `git-profile use <id>`, only fields present in the profile are written:

| Profile field  | git config key    |
| -------------- | ----------------- |
| `git_user`     | `user.name`       |
| `git_email`    | `user.email`      |
| `ssh_key_path` | `core.sshCommand` |
| `gpg_key_id`   | `user.signingkey` |
| `sign_commits` | `commit.gpgsign`  |

Fields absent from the profile are left unchanged in git config.

## Environment variables

| Variable          | Description                                           |
| ----------------- | ----------------------------------------------------- |
| `XDG_CONFIG_HOME` | Base directory for config file (default: `~/.config`) |

## Dotfiles integration

Because the config file is plain JSON, you can symlink or copy it from your dotfiles:

```bash
# symlink from dotfiles
ln -s ~/dotfiles/git-profile.json ~/.config/git-profile/config.json

# or use --config flag in a shell alias
alias gp='git-profile --config ~/dotfiles/git-profile.json'
```
