# Installing prune_backups

`prune_backups` is a single, self-contained, dependency-free binary (no
runtime, no libraries, no cgo) - "installing" it is mostly just "get the
right file onto your `PATH`". Pick whichever of the following fits how you
manage software.

## Linux: package repository (recommended - auto-updating)

`prune_backups` is available through the
[acmelab package repository](https://pkg.acmelab.de), a signed
`apt`/`dnf`/`zypper`/`pacman`/`apk` repository built directly from this
project's own GitHub Releases (see
[TomTonic/pkg-repo](https://github.com/TomTonic/pkg-repo) for how it's
built and why only the index, not every package, is signed). Once added,
`apt upgrade`/`dnf upgrade`/`pacman -Syu`/`apk upgrade` pick up new
releases automatically - no manual downloads needed.

* **Debian / Ubuntu** (and derivatives such as Linux Mint, Pop!_OS, Raspberry Pi OS):

  ```Shell
  curl -fsSL https://pkg.acmelab.de/pubkey.gpg | sudo tee /etc/apt/keyrings/acmelab.asc
  echo "deb [signed-by=/etc/apt/keyrings/acmelab.asc] https://pkg.acmelab.de/apt stable main" | \
    sudo tee /etc/apt/sources.list.d/acmelab.list
  sudo apt update && sudo apt install prune_backups
  ```

* **Fedora / RHEL / CentOS / Rocky / AlmaLinux / openSUSE / SLES**:

  ```Shell
  sudo curl -fsSL -o /etc/yum.repos.d/acmelab.repo https://pkg.acmelab.de/rpm/acmelab.repo
  sudo dnf install prune_backups   # or: sudo zypper install prune_backups
  ```

* **Arch Linux** (and derivatives such as Manjaro, EndeavourOS):

  ```Shell
  # add to /etc/pacman.conf:
  #   [acmelab]
  #   SigLevel = Optional TrustedOnly DatabaseRequired
  #   Server = https://pkg.acmelab.de/pacman/$arch
  curl -fsSL https://pkg.acmelab.de/pubkey.gpg | sudo pacman-key --add -
  sudo pacman-key --lsign-key 284B3557CDC44D25509DA0A3B9C061B9627E9BB0
  sudo pacman -Sy prune_backups
  ```

* **Alpine Linux**:

  ```Shell
  sudo curl -fsSL -o /etc/apk/keys/acmelab.rsa.pub https://pkg.acmelab.de/alpine/acmelab.rsa.pub
  echo "https://pkg.acmelab.de/apk" | sudo tee -a /etc/apk/repositories
  sudo apk update && sudo apk add prune_backups
  ```

## Linux: manual package download

Every [release](https://github.com/TomTonic/prune_backups/releases) also
ships the same packages as plain assets, for anyone who'd rather not add
the repository above. Since the binary has no C library dependency (no
cgo), one package per architecture works across all versions of the
corresponding distro family that are still supported by the Go toolchain
used to build it - there is no need to pick a package per Ubuntu/Debian/
Fedora version.

* **Debian / Ubuntu** (and derivatives): download the `.deb` asset, then run:

  ```Shell
  sudo apt install ./prune_backups_<version>_<amd64|arm64>.deb
  ```

* **Fedora / RHEL / CentOS / Rocky / AlmaLinux**: download the `.rpm` asset, then run:

  ```Shell
  sudo dnf install ./prune_backups-<version>-1.<x86_64|aarch64>.rpm
  ```

* **openSUSE / SLES**: same `.rpm` asset, installed with:

  ```Shell
  sudo zypper install ./prune_backups-<version>-1.<x86_64|aarch64>.rpm
  ```

* **Alpine Linux**: download the `.apk` asset, then run:

  ```Shell
  sudo apk add --allow-untrusted ./prune_backups_<version>_<x86_64|aarch64>.apk
  ```

* **Arch Linux** (and derivatives): download the `.pkg.tar.zst` asset, then run:

  ```Shell
  sudo pacman -U ./prune_backups-<version>-1-<x86_64|aarch64>.pkg.tar.zst
  ```

All of these install the `prune_backups` binary to `/usr/bin/prune_backups`,
so it is immediately available on your `PATH`.

## Windows

Download `prune_backups-<amd64|arm64|x86>-win.exe` from the
[releases page](https://github.com/TomTonic/prune_backups/releases) and
place it wherever you like - it's fully self-contained, no installer
needed. Since the binary isn't code-signed, Windows SmartScreen will warn
on first run: click **More info** → **Run anyway** (or right-click the
file → **Properties** → **Unblock**).

## macOS

Download `prune_backups-<amd64|arm64>-mac` from the
[releases page](https://github.com/TomTonic/prune_backups/releases), then:

```Shell
chmod +x prune_backups-<amd64|arm64>-mac
xattr -d com.apple.quarantine prune_backups-<amd64|arm64>-mac   # since it's not notarized
mv prune_backups-<amd64|arm64>-mac /usr/local/bin/prune_backups  # or anywhere on your PATH
```

Alternatively, right-click the file in Finder → **Open**, then confirm in
the Gatekeeper dialog on first run instead of using `xattr`.

## Build your own executable

* Install the Go compiler, if you do not yet have it (check with
  `go version`). Available for free for many operating systems and
  architectures (x86, x64, ARM32, ARM64, Windows, Linux, macOS, etc.), see
  <https://go.dev/dl/>.
* Download source code: `git clone https://github.com/TomTonic/prune_backups.git`
* Compile the source: `go build`
* Optionally run the tests: `go test -cover`
* The executable is named `prune_backups` or `prune_backups.exe`, depending
  on your system. You can place it anywhere you like, it is fully
  self-contained.

These are the same binaries the GitHub Actions release workflow in this
repository builds - if you don't trust the pre-built ones, building your
own from source gives you the identical result.
