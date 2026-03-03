# Installation

## Homebrew (macOS / Linux)

```bash
brew install hapiio/tap/git-profile
```

## go install

```bash
go install github.com/hapiio/git-profile@latest
```

## Pre-built binaries

Download the latest binary for your platform from the [Releases](https://github.com/hapiio/git-profile/releases/latest) page.

### macOS (Apple Silicon)

 ```bash
curl -L https://github.com/hapiio/git-profile/releases/latest/download/git-profile_darwin_arm64.tar.gz \
      | tar -xz && mv git-profile /usr/local/bin/
```

### macOS (Intel)

```bash
curl -L https://github.com/hapiio/git-profile/releases/latest/download/git-profile_darwin_x86_64.tar.gz \
      | tar -xz && mv git-profile /usr/local/bin/
```

### Linux (amd64)

```bash
curl -L https://github.com/hapiio/git-profile/releases/latest/download/git-profile_linux_x86_64.tar.gz \
      | tar -xz && mv git-profile /usr/local/bin/
```

### Linux (arm64)

```bash
curl -L https://github.com/hapiio/git-profile/releases/latest/download/git-profile_linux_arm64.tar.gz \
      | tar -xz && mv git-profile /usr/local/bin/
```

## Linux packages

`.deb` and `.rpm` packages are available on the [Releases](https://github.com/hapiio/git-profile/releases/latest) page.

### Debian / Ubuntu

```bash
curl -LO https://github.com/hapiio/git-profile/releases/latest/download/git-profile_linux_amd64.deb
    sudo dpkg -i git-profile_linux_amd64.deb
```

### Fedora / RHEL

```bash
curl -LO https://github.com/hapiio/git-profile/releases/latest/download/git-profile_linux_amd64.rpm
    sudo rpm -i git-profile_linux_amd64.rpm
```

## Arch Linux (AUR)

```bash
yay -S git-profile-bin
```

## Shell completions

After installing, enable shell completions for a better experience:

### bash

```bash
git-profile completion bash > /etc/bash_completion.d/git-profile
```

### zsh

```bash
git-profile completion zsh > "${fpath[1]}/_git-profile"
```

### fish

```bash
git-profile completion fish > ~/.config/fish/completions/git-profile.fish
```

## Verify installation

```bash
git-profile version
```

You should see the version number, commit hash, and build date.
