# Quick Start

## 1. Add your profiles

```bash
git-profile add --id work     --name "Jane Dev" --email jane@company.com
git-profile add --id personal --name "Jane Doe" --email jane@example.com \
                               --ssh-key ~/.ssh/id_ed25519_personal
```

## 2. Apply a profile

Apply a profile to the **current repository**:

```bash
git-profile use work
```

Or apply one **globally** (sets `~/.gitconfig`):

```bash
git-profile use personal --global
```

## 3. Check the active identity

```bash
git-profile current
```

```text
Current git identity

  user.name   Jane Dev
  user.email  jane@company.com
  ssh-key     (default)

  Matched profile: work
```

## 4. List all profiles

```bash
git-profile list
```

```text
  PROFILE     USER         EMAIL                  SSH KEY
  ──────────────────────────────────────────────────────────
  personal    Jane Doe     jane@example.com        ~/.ssh/id_ed25519_personal
● work        Jane Dev     jane@company.com        (default)
```

The `●` marker shows the profile that matches the current git identity.

## 5. Auto-apply with git hooks (optional)

Never forget to switch profiles again:

```bash
cd my-work-project
git-profile install-hooks
git-profile set-default work
```

Now every `git commit` and `git push` in that repo automatically uses the `work` profile.

See the [Auto-Apply Hooks guide](../guides/hooks.md) for full details.

## Next steps

- [Full command reference](commands.md) — all flags for every command
- [GPG commit signing](../guides/gpg.md) — sign commits per identity
- [Configuration](../configuration.md) — config file format and location
