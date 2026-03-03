# Stop Committing with the Wrong Git Identity

We've all done it. You finish a feature, push it, and then realize you just committed to your client's repo as `Jane Doe <jane@personal.com>` instead of `Jane Dev <jane@company.com>`.

Or worse — you opened a PR to an open-source project and your work email is now public forever.

Managing multiple git identities is a solved problem that nobody has solved cleanly. Until now.

---

## The old ways

**Option 1: Edit `~/.gitconfig` before every project**

```bash
git config --global user.email jane@company.com
# ... do some work ...
git config --global user.email jane@personal.com
# ... forget to switch back ...
```

You will forget. Every time.

**Option 2: `includeIf` in `~/.gitconfig`**

```ini
[includeIf "gitdir:~/work/"]
  path = ~/.gitconfig-work
[includeIf "gitdir:~/personal/"]
  path = ~/.gitconfig-personal
```

Works if every project lives in a tidy folder structure. Mine don't. Yours probably don't either.

**Option 3: Set it per-repo, manually, every time**

```bash
git clone git@github.com:company/project.git
cd project
git config user.name "Jane Dev"
git config user.email "jane@company.com"
# also need to set core.sshCommand if using per-identity SSH keys...
# and user.signingkey if GPG signing...
```

Four commands every single time. That's not a workflow, that's a tax.

---

## A better way: `git-profile`

`git-profile` is a small CLI tool I built in Go that lets you define named identity profiles and apply them to any repo with a single command.

```bash
# One-time setup
git-profile add --id work     --name "Jane Dev"  --email jane@company.com
git-profile add --id personal --name "Jane Doe"  --email jane@example.com \
                               --ssh-key ~/.ssh/id_ed25519_personal

# Per repo, one command
git-profile use work
```

That's it. It sets `user.name`, `user.email`, `core.sshCommand`, `user.signingkey`, and `commit.gpgsign` all at once, only for that repo.

---

## What it looks like day-to-day

```bash
$ git-profile list

  PROFILE     USER         EMAIL                    SSH KEY
  ──────────────────────────────────────────────────────────────
  personal    Jane Doe     jane@example.com          ~/.ssh/id_ed25519_personal
● work        Jane Dev     jane@company.com          (default)

$ git-profile current

  Current git identity

    user.name   Jane Dev
    user.email  jane@company.com
    ssh-key     (default)

    Matched profile: work
```

The `●` shows which profile matches your current repo's identity.

---

## The "set it and forget it" mode

The best feature: git hooks that automatically apply the right profile before every commit and push.

```bash
cd my-work-project
git-profile install-hooks
git-profile set-default work
```

Now `git commit` in that repo always uses the `work` profile. No more thinking about it.

You can also set a global fallback:

```bash
git-profile set-default personal --global
```

Any repo without an explicit default falls back to `personal`.

---

## GPG signing per identity

If you sign commits, you probably have different GPG keys per identity. `git-profile` handles that too:

```bash
git-profile add --id work \
  --name "Jane Dev" \
  --email jane@company.com \
  --gpg-key "ABC123DEF456" \
  --sign-commits
```

When you run `git-profile use work`, it sets all four keys at once:

```ini
user.name       = Jane Dev
user.email      = jane@company.com
user.signingkey = ABC123DEF456
commit.gpgsign  = true
```

---

## Already have git configured? Import it

```bash
git-profile import --id work --global
```

Reads your current `~/.gitconfig` (name, email, SSH command, signing key) and creates a profile from it. Zero re-typing.

---

## Install

```bash
# Homebrew
brew install hapiio/tap/git-profile

# Go
go install github.com/hapiio/git-profile@latest
```

Binaries for Linux and macOS (Apple Silicon + Intel) are on the [releases page](https://github.com/hapiio/git-profile/releases).

Shell completions for bash, zsh, and fish are included.

---

## Why I built it

I switch between a full-time job, personal projects, and a few open-source repos every day. After the third time I pushed a commit with the wrong email to a client repo, I looked for a tool that handled all of it — SSH keys, GPG keys, auto-apply hooks — in one place. I didn't find one, so I built it.

The entire tool is a single static binary with no runtime dependencies. Profiles are stored as plain JSON in `~/.config/git-profile/config.json` so they're easy to back up or commit to a dotfiles repo.

---

Source and docs: **[github.com/hapiio/git-profile](https://github.com/hapiio/git-profile)**

Happy to hear feedback or feature requests in the issues.

---

> **Tags:** `git` `devtools` `golang` `productivity` `opensource`
