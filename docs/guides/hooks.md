# Auto-Apply Hooks

Git hooks let `git-profile` automatically enforce the correct identity before every commit and push — no more accidentally committing with the wrong email.

## Setup

```bash
# Step 1: go into your project
cd my-work-project

# Step 2: install the hooks
git-profile install-hooks

# Step 3: set the default profile for this repo
git-profile set-default work
```

That's it. From now on, `git commit` and `git push` in this repo automatically call `git-profile ensure` before running.

## How it works

`install-hooks` writes two hook scripts into `.git/hooks/`:

- **`prepare-commit-msg`** — runs before the commit editor opens
- **`pre-push`** — runs before a push is sent

Each hook calls `git-profile ensure`, which applies the profile following this resolution order:

| Priority | Source                   | How to set                          |
| -------- | ------------------------ | ----------------------------------- |
| 1        | Local `gitprofile.default`  | `git-profile set-default <id>`      |
| 2        | Global `gitprofile.default` | `git-profile set-default <id> --global` |
| 3        | Interactive picker        | Terminal only (not in CI)           |

## Global fallback

Set a personal profile as the fallback for repos that don't have a local default:

```bash
git-profile set-default personal --global
```

Now any repo without an explicit default uses `personal`, and repos with `set-default work` override it with `work`.

## Non-interactive environments (CI/CD)

When stdin is not a terminal (e.g., in CI), the interactive picker is skipped and `ensure` exits with an error if no default is configured. This prevents CI pipelines from hanging on interactive prompts.

```bash
# In CI: set the profile explicitly before committing
git-profile use ci-bot
```

## Removing hooks

To remove the installed hooks:

```bash
rm .git/hooks/prepare-commit-msg .git/hooks/pre-push
```

## Example: multi-repo workflow

```bash
# Work projects
for repo in ~/work/*; do
  cd "$repo"
  git-profile install-hooks
  git-profile set-default work
done

# Personal projects
for repo in ~/personal/*; do
  cd "$repo"
  git-profile install-hooks
  git-profile set-default personal
done

# Global fallback
git-profile set-default personal --global
```
