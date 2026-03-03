# GPG Commit Signing

`git-profile` manages all GPG signing settings — `user.signingkey` and `commit.gpgsign` — so you never have to edit them manually.

## 1. Find your GPG key ID

```bash
gpg --list-secret-keys --keyid-format=long
```

```text
sec   ed25519/ABC123DEF456 2024-01-01 [SC]
      ABCDEF1234567890ABCDEF1234567890ABC123DE
uid   [ultimate] Jane Dev <jane@company.com>
```

The key ID is the part after `/` on the `sec` line — `ABC123DEF456` in this example.

## 2. Create a profile with GPG signing

```bash
git-profile add \
  --id work \
  --name "Jane Dev" \
  --email jane@company.com \
  --gpg-key "ABC123DEF456" \
  --sign-commits
```

## 3. Apply the profile

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

## 4. Verify signing works

```bash
git commit --allow-empty -m "test gpg signing"
git log --show-signature -1
```

## Using different keys per identity

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

## Disabling GPG signing for a repo

Applying a profile that has no `gpg_key_id` does **not** unset signing keys — it leaves any existing `user.signingkey` and `commit.gpgsign` values in place. To explicitly disable signing:

```bash
git config --unset user.signingkey
git config --unset commit.gpgsign
```

## Troubleshooting

### `gpg: skipped "KEY_ID": No secret key`

The signing key stored in the profile doesn't match any key in your GPG keyring. Either:

- Import the key: `gpg --import private-key.asc`
- Or update the profile with the correct key ID: `git-profile edit work`

### `error: gpg failed to sign the data`

Try running `gpg --status-fd=2 -bsau KEY_ID </dev/null` to test signing directly. Common causes:

- GPG agent is not running: `gpg-agent --daemon`
- On macOS: ensure `pinentry-mac` is installed: `brew install pinentry-mac`
