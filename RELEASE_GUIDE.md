# Paso Release & Distribution Guide

This guide explains how to distribute Paso through package managers instead of requiring users to build from source.

## Table of Contents

- [The Problem with Building from Source](#the-problem-with-building-from-source)
- [Package Distribution: The Solution](#package-distribution-the-solution)
- [How Different Package Managers Work](#how-different-package-managers-work)
- [Service Management: systemd vs launchd](#service-management-systemd-vs-launchd)
- [The Modern Approach: GoReleaser](#the-modern-approach-goreleaser)
- [Full Release Pipeline](#full-release-pipeline)
- [Trade-offs](#trade-offs)

## The Problem with Building from Source

Our current `scripts/install.sh`:
1. Requires Go to be installed
2. Runs `go build` to compile the binaries
3. Copies them to `~/.local/bin`
4. Sets up systemd service (Linux only)

This creates friction for users who just want to use the tool.

## Package Distribution: The Solution

Package managers distribute **pre-built binaries** along with metadata about dependencies, installation locations, and service files.

**Key insight**: You build once (per architecture), upload to a repository, and users download the compiled binary.

You become responsible for building for different:
- Architectures: `amd64`, `arm64`
- Operating systems: `linux`, `darwin` (macOS)

## How Different Package Managers Work

### AUR (Arch Linux) - `yay -S paso`

The AUR has two approaches:

#### 1. Source-based (still requires Go)

`PKGBUILD`:
```bash
# Maintainer: Your Name <email>
pkgname=paso
pkgver=1.0.0
pkgrel=1
pkgdesc="Terminal Kanban board"
arch=('x86_64' 'aarch64')
url="https://github.com/yourusername/paso"
license=('MIT')
depends=('glibc')
makedepends=('go')
source=("$pkgname-$pkgver.tar.gz::https://github.com/yourusername/paso/archive/v$pkgver.tar.gz")
sha256sums=('...')

build() {
    cd "$pkgname-$pkgver"
    export CGO_ENABLED=0
    go build -o bin/paso -ldflags="-s -w" .
    go build -o bin/paso-daemon -ldflags="-s -w" ./cmd/daemon
}

package() {
    cd "$pkgname-$pkgver"

    # Install binaries
    install -Dm755 bin/paso "$pkgdir/usr/bin/paso"
    install -Dm755 bin/paso-daemon "$pkgdir/usr/bin/paso-daemon"

    # Install systemd service
    install -Dm644 systemd/paso.service "$pkgdir/usr/lib/systemd/user/paso.service"

    # Install license
    install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
}
```

#### 2. Binary-based (NO Go needed!)

`PKGBUILD` for `paso-bin`:
```bash
pkgname=paso-bin
pkgver=1.0.0
pkgrel=1
pkgdesc="Terminal Kanban board (binary release)"
arch=('x86_64' 'aarch64')
url="https://github.com/yourusername/paso"
license=('MIT')
depends=('glibc')
provides=('paso')
conflicts=('paso')
source_x86_64=("https://github.com/user/paso/releases/download/v$pkgver/paso-linux-amd64.tar.gz")
source_aarch64=("https://github.com/user/paso/releases/download/v$pkgver/paso-linux-arm64.tar.gz")
sha256sums_x86_64=('...')
sha256sums_aarch64=('...')

package() {
    # Install binaries
    install -Dm755 paso "$pkgdir/usr/bin/paso"
    install -Dm755 paso-daemon "$pkgdir/usr/bin/paso-daemon"

    # Install systemd service
    install -Dm644 paso.service "$pkgdir/usr/lib/systemd/user/paso.service"

    # Install license
    install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
}
```

Users install with:
```bash
yay -S paso-bin
systemctl --user enable paso.service
systemctl --user start paso.service
```

### Debian/Ubuntu - `apt install paso`

Requires building `.deb` packages.

`debian/control`:
```
Source: paso
Section: utils
Priority: optional
Maintainer: Your Name <email>
Build-Depends: golang-go
Standards-Version: 4.5.0

Package: paso
Architecture: amd64 arm64
Depends: ${shlibs:Depends}, ${misc:Depends}, systemd
Description: Terminal Kanban board
 A beautiful terminal-based kanban board for task management
 with real-time sync across terminal sessions.
```

`debian/paso.service` (systemd unit):
```ini
[Unit]
Description=Paso Daemon - Terminal Kanban Board Real-time Sync
After=network.target

[Service]
Type=simple
ExecStart=/usr/bin/paso-daemon
Restart=on-failure
RestartSec=5

# Security hardening
PrivateTmp=yes
NoNewPrivileges=yes
ProtectSystem=strict
ReadWritePaths=%h/.paso

# Logging
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=default.target
```

Building:
```bash
dpkg-buildpackage -us -uc  # Creates .deb file
```

**Note**: To support `apt install`, you need a repository. Options:
- Create a PPA (Personal Package Archive) on Launchpad
- Host your own apt repository
- Use packagecloud.io or similar services

### Homebrew (macOS) - `brew install paso`

Homebrew formula (Ruby DSL):

```ruby
class Paso < Formula
  desc "Terminal Kanban board"
  homepage "https://github.com/yourusername/paso"
  url "https://github.com/yourusername/paso/archive/v1.0.0.tar.gz"
  sha256 "..."
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-o", "bin/paso", "."
    system "go", "build", "-o", "bin/paso-daemon", "./cmd/daemon"
    bin.install "bin/paso"
    bin.install "bin/paso-daemon"

    # Install launchd plist
    (prefix/"LaunchAgents").install "launchd/com.paso.daemon.plist"
  end

  def caveats
    <<~EOS
      To start the paso daemon automatically on login:
        ln -sfv #{prefix}/LaunchAgents/com.paso.daemon.plist ~/Library/LaunchAgents/
        launchctl load ~/Library/LaunchAgents/com.paso.daemon.plist
    EOS
  end

  test do
    system "#{bin}/paso", "--version"
  end
end
```

For a binary-only formula (using pre-built releases):
```ruby
class Paso < Formula
  desc "Terminal Kanban board"
  homepage "https://github.com/yourusername/paso"
  version "1.0.0"
  license "MIT"

  if Hardware::CPU.arm?
    url "https://github.com/yourusername/paso/releases/download/v1.0.0/paso-darwin-arm64.tar.gz"
    sha256 "..."
  else
    url "https://github.com/yourusername/paso/releases/download/v1.0.0/paso-darwin-amd64.tar.gz"
    sha256 "..."
  end

  def install
    bin.install "paso"
    bin.install "paso-daemon"
    (prefix/"LaunchAgents").install "com.paso.daemon.plist"
  end

  def caveats
    <<~EOS
      To start the paso daemon automatically on login:
        ln -sfv #{prefix}/LaunchAgents/com.paso.daemon.plist ~/Library/LaunchAgents/
        launchctl load ~/Library/LaunchAgents/com.paso.daemon.plist
    EOS
  end

  test do
    system "#{bin}/paso", "--version"
  end
end
```

## Service Management: systemd vs launchd

**Important**: macOS does NOT have systemd. It uses launchd.

### Linux: systemd

`systemd/paso.service`:
```ini
[Unit]
Description=Paso Daemon - Terminal Kanban Board Real-time Sync
After=network.target

[Service]
Type=simple
ExecStart=/usr/bin/paso-daemon
Restart=on-failure
RestartSec=5

# Security hardening
PrivateTmp=yes
NoNewPrivileges=yes
ProtectSystem=strict
ReadWritePaths=%h/.paso

# Logging
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=default.target
```

**User commands**:
```bash
# Enable and start
systemctl --user enable paso.service
systemctl --user start paso.service

# Check status
systemctl --user status paso.service

# View logs
journalctl --user -u paso -f

# Stop
systemctl --user stop paso.service
```

### macOS: launchd

`launchd/com.paso.daemon.plist`:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.paso.daemon</string>

    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/paso-daemon</string>
    </array>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <dict>
        <key>SuccessfulExit</key>
        <false/>
    </dict>

    <key>StandardOutPath</key>
    <string>/tmp/paso-daemon.log</string>

    <key>StandardErrorPath</key>
    <string>/tmp/paso-daemon.error.log</string>

    <key>ThrottleInterval</key>
    <integer>5</integer>
</dict>
</plist>
```

**User commands**:
```bash
# Install (copy to LaunchAgents)
cp com.paso.daemon.plist ~/Library/LaunchAgents/

# Load and start
launchctl load ~/Library/LaunchAgents/com.paso.daemon.plist

# Check status
launchctl list | grep paso

# View logs
tail -f /tmp/paso-daemon.log

# Unload and stop
launchctl unload ~/Library/LaunchAgents/com.paso.daemon.plist
```

## The Modern Approach: GoReleaser

Most Go projects use **GoReleaser** to automate everything.

### Installation

```bash
# macOS
brew install goreleaser

# Linux
go install github.com/goreleaser/goreleaser@latest
```

### Configuration

`.goreleaser.yaml`:
```yaml
project_name: paso

before:
  hooks:
    - go mod tidy

builds:
  - id: paso
    binary: paso
    main: .
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64
    env:
      - CGO_ENABLED=0
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

  - id: paso-daemon
    binary: paso-daemon
    main: ./cmd/daemon
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64
    env:
      - CGO_ENABLED=0
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - id: default
    format: tar.gz
    name_template: >-
      {{ .ProjectName }}-
      {{- .Os }}-
      {{- if eq .Arch "amd64" }}x86_64
      {{- else }}{{ .Arch }}{{ end }}
    files:
      - LICENSE
      - README.md
      - src: systemd/paso.service
        dst: paso.service
        strip_parent: true
        info:
          mode: 0644
      - src: launchd/com.paso.daemon.plist
        dst: com.paso.daemon.plist
        strip_parent: true
        info:
          mode: 0644

nfpms:
  - id: paso
    package_name: paso
    homepage: https://github.com/yourusername/paso
    description: Terminal Kanban board with real-time sync
    maintainer: Your Name <email@example.com>
    license: MIT

    formats:
      - deb
      - rpm
      - archlinux

    bindir: /usr/bin

    contents:
      # Install systemd service (Linux only)
      - src: systemd/paso.service
        dst: /usr/lib/systemd/user/paso.service
        file_info:
          mode: 0644

    scripts:
      postinstall: scripts/postinstall.sh
      preremove: scripts/preremove.sh

    dependencies:
      - systemd

    archlinux:
      pkgbase: paso
      packager: Your Name <email@example.com>

brews:
  - name: paso
    homepage: https://github.com/yourusername/paso
    description: Terminal Kanban board with real-time sync
    license: MIT

    repository:
      owner: yourusername
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"

    folder: Formula

    install: |
      bin.install "paso"
      bin.install "paso-daemon"
      (prefix/"LaunchAgents").install "com.paso.daemon.plist"

    caveats: |
      To start the paso daemon automatically on login:
        ln -sfv #{prefix}/LaunchAgents/com.paso.daemon.plist ~/Library/LaunchAgents/
        launchctl load ~/Library/LaunchAgents/com.paso.daemon.plist

      To start manually:
        paso-daemon &

aurs:
  - name: paso-bin
    homepage: https://github.com/yourusername/paso
    description: Terminal Kanban board with real-time sync (binary release)
    license: MIT

    maintainers:
      - 'Your Name <email@example.com>'

    contributors:
      - 'Your Name <email@example.com>'

    provides:
      - paso

    conflicts:
      - paso

    depends:
      - glibc

    git_url: ssh://aur@aur.archlinux.org/paso-bin.git

    private_key: '{{ .Env.AUR_SSH_KEY }}'

    package: |
      # Install binaries
      install -Dm755 paso "${pkgdir}/usr/bin/paso"
      install -Dm755 paso-daemon "${pkgdir}/usr/bin/paso-daemon"

      # Install systemd service
      install -Dm644 paso.service "${pkgdir}/usr/lib/systemd/user/paso.service"

      # Install license
      install -Dm644 LICENSE "${pkgdir}/usr/share/licenses/${pkgname}/LICENSE"

checksum:
  name_template: 'checksums.txt'

snapshot:
  name_template: "{{ incpatch .Version }}-next"

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'
```

### Post-install Scripts

`scripts/postinstall.sh`:
```bash
#!/bin/bash
# Post-install script for .deb/.rpm packages

echo "Paso has been installed!"
echo ""
echo "To enable the daemon to start on login:"
echo "  systemctl --user enable paso.service"
echo "  systemctl --user start paso.service"
echo ""
echo "To start manually:"
echo "  paso-daemon &"
```

`scripts/preremove.sh`:
```bash
#!/bin/bash
# Pre-remove script

# Stop the daemon if running
if systemctl --user is-active paso.service >/dev/null 2>&1; then
    systemctl --user stop paso.service
    systemctl --user disable paso.service
fi
```

### GitHub Actions Workflow

`.github/workflows/release.yml`:
```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write
  packages: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
          AUR_SSH_KEY: ${{ secrets.AUR_SSH_KEY }}
```

## Full Release Pipeline

### 1. Development
Write code, commit to git normally.

### 2. Prepare Release
```bash
# Update version, changelog, etc.
git add .
git commit -m "chore: prepare v1.0.0 release"
```

### 3. Tag Release
```bash
git tag v1.0.0
git push origin main
git push origin v1.0.0
```

### 4. Automated Build (GitHub Actions)
When you push a tag, GitHub Actions automatically:
- Runs GoReleaser
- Builds for all OS/arch combinations:
  - `paso-linux-amd64.tar.gz`
  - `paso-linux-arm64.tar.gz`
  - `paso-darwin-amd64.tar.gz`
  - `paso-darwin-arm64.tar.gz`
- Creates packages:
  - `paso_1.0.0_amd64.deb`
  - `paso_1.0.0_arm64.deb`
  - `paso-1.0.0-1.x86_64.rpm`
  - `paso-1.0.0-1-x86_64.pkg.tar.zst` (Arch)
- Uploads to GitHub Releases
- Pushes to Homebrew tap
- Pushes to AUR repository

### 5. Users Install

**Arch Linux**:
```bash
yay -S paso-bin
systemctl --user enable paso.service
systemctl --user start paso.service
```

**Ubuntu/Debian** (with PPA or direct download):
```bash
# From .deb file
wget https://github.com/you/paso/releases/latest/download/paso_1.0.0_amd64.deb
sudo dpkg -i paso_1.0.0_amd64.deb

# Enable daemon
systemctl --user enable paso.service
systemctl --user start paso.service
```

**macOS**:
```bash
brew install yourname/tap/paso

# Enable daemon
ln -sfv /usr/local/Cellar/paso/1.0.0/LaunchAgents/com.paso.daemon.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.paso.daemon.plist
```

**Direct download** (any Linux):
```bash
# Detect architecture
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    PASO_ARCH="amd64"
elif [ "$ARCH" = "aarch64" ]; then
    PASO_ARCH="arm64"
fi

# Download and install
curl -L "https://github.com/you/paso/releases/latest/download/paso-linux-${PASO_ARCH}.tar.gz" | tar xz
sudo mv paso paso-daemon /usr/local/bin/

# Install systemd service
mkdir -p ~/.config/systemd/user
curl -L "https://github.com/you/paso/releases/latest/download/paso.service" \
    -o ~/.config/systemd/user/paso.service

# Enable and start
systemctl --user daemon-reload
systemctl --user enable paso.service
systemctl --user start paso.service
```

## Trade-offs

### Pros
- Users don't need Go installed
- Faster installation (no compilation)
- More "professional" distribution
- Can sign binaries for security
- Easier updates via package managers
- Service management included

### Cons
- You build for multiple platforms
- Need CI/CD setup (GitHub Actions)
- Slightly more complex release process
- Need to maintain package recipes
- Need to manage service files for different init systems

## Recommended Approach

### Phase 1: GitHub Releases (Start Here)
Set up GoReleaser to build and publish to GitHub Releases. Users can download binaries directly.

**Effort**: Low
**User reach**: Medium
**Commands**: Manual download + install

### Phase 2: AUR (Easy Win for Arch Users)
Create `paso-bin` package in AUR.

**Effort**: Low (just a PKGBUILD file)
**User reach**: All Arch/Manjaro users
**Commands**: `yay -S paso-bin`

### Phase 3: Homebrew Tap
Create your own tap for macOS users.

**Effort**: Medium (need separate repo)
**User reach**: All macOS users
**Commands**: `brew install yourname/tap/paso`

### Phase 4: Official Repos (Long-term)
Get into official Homebrew, create PPA for Ubuntu, etc.

**Effort**: High
**User reach**: Maximum
**Commands**: `brew install paso`, `apt install paso`

## Next Steps

1. Create service files:
   - `systemd/paso.service`
   - `launchd/com.paso.daemon.plist`

2. Create `.goreleaser.yaml` configuration

3. Set up GitHub Actions workflow

4. Create first release:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

5. Test installation from release artifacts

6. Create AUR package

7. Create Homebrew tap

8. Update documentation with installation instructions
