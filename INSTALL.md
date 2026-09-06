# Installing prune_backups

`prune_backups` is a single, self-contained, dependency-free binary (no
runtime, no libraries, no cgo) - "installing" it is mostly just "get the
right file onto your `PATH`". Pick whichever of the following fits how you
manage software.

## Linux: package repository (recommended - auto-updating)

`prune_backups` is available through the
[TomTonic package repository](https://pkg.tomtonic.de), a signed
`apt`/`dnf`/`zypper`/`pacman`/`apk` repository built directly from this
project's own GitHub Releases (see
[TomTonic/pkg-repo](https://github.com/TomTonic/pkg-repo) for how it's
built and why only the index, not every package, is signed). Once added,
`apt upgrade`/`dnf upgrade`/`pacman -Syu`/`apk upgrade` pick up new
releases automatically - no manual downloads needed.

* **Debian / Ubuntu** (and derivatives such as Linux Mint, Pop!_OS, Raspberry Pi OS):

  ```Shell
  curl -fsSL https://pkg.tomtonic.de/pubkey.gpg | sudo tee /etc/apt/keyrings/tomtonic.asc
  echo "deb [signed-by=/etc/apt/keyrings/tomtonic.asc] https://pkg.tomtonic.de/apt stable main" | \
    sudo tee /etc/apt/sources.list.d/tomtonic.list
  sudo apt update && sudo apt install prune_backups
  ```

* **Fedora / RHEL / CentOS / Rocky / AlmaLinux / openSUSE / SLES**:

  ```Shell
  sudo curl -fsSL -o /etc/yum.repos.d/tomtonic.repo https://pkg.tomtonic.de/rpm/tomtonic.repo
  sudo dnf install prune_backups   # or: sudo zypper install prune_backups
  ```

* **Arch Linux** (and derivatives such as Manjaro, EndeavourOS):

  ```Shell
  # add to /etc/pacman.conf:
  #   [tomtonic]
  #   SigLevel = Optional TrustedOnly DatabaseRequired
  #   Server = https://pkg.tomtonic.de/pacman/$arch
  curl -fsSL https://pkg.tomtonic.de/pubkey.gpg | sudo pacman-key --add -
  sudo pacman-key --lsign-key AA9C6D63B7B6C0BC18A89693E3725F71EDDAEC03
  sudo pacman -Sy prune_backups
  ```

* **Alpine Linux**:

  ```Shell
  sudo curl -fsSL -o /etc/apk/keys/tomtonic.rsa.pub https://pkg.tomtonic.de/alpine/tomtonic.rsa.pub
  echo "https://pkg.tomtonic.de/apk" | sudo tee -a /etc/apk/repositories
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

## macOS: Homebrew (recommended - auto-updating)

prune_backups is available via a [Homebrew tap](https://github.com/TomTonic/homebrew-tap)
maintained alongside this project (not homebrew-core), published automatically
by each release:

```Shell
brew install TomTonic/tap/prune_backups
```

`brew upgrade` picks up new releases automatically, and the binary lands on
your `PATH` at `/opt/homebrew/bin/prune_backups` (Apple silicon) or
`/usr/local/bin/prune_backups` (Intel).

The cask clears the `com.apple.quarantine` attribute that Homebrew Cask sets
on everything it stages. Without that, macOS does not merely warn about the
un-notarized binary - Gatekeeper kills it outright, and the process dies with
exit code 137 and no output at all. Integrity is not weakened: Homebrew has
already verified the download against the SHA-256 pinned in the cask before
the attribute is removed.

## Linux / macOS: manual archive download

For anyone who would rather not use a package manager, every release also
ships plain archives named
`prune_backups_<version>_<linux|darwin>_<amd64|arm64|386|arm>.tar.gz`, plus a
`checksums.txt` covering every archive. Verify before installing:

```Shell
sha256sum --check --ignore-missing checksums.txt   # shasum -a 256 -c on macOS
tar xzf prune_backups_<version>_<os>_<arch>.tar.gz
sudo mv prune_backups /usr/local/bin/prune_backups   # or anywhere on your PATH
```

On macOS a manually downloaded binary is not notarized, so Gatekeeper will
refuse to run it. Clear the quarantine attribute yourself:

```Shell
xattr -d com.apple.quarantine prune_backups
```

Alternatively, right-click the file in Finder → **Open**, then confirm in
the Gatekeeper dialog on first run. The Homebrew tap above does this step for
you, which is why it is the recommended route on macOS.

## Windows

Download `prune_backups_<version>_windows_<amd64|arm64|386>.zip` from the
[releases page](https://github.com/TomTonic/prune_backups/releases), unpack
it, and place `prune_backups.exe` wherever you like - it's fully
self-contained, no installer needed. Since the binary isn't code-signed,
Windows SmartScreen will warn on first run: click **More info** → **Run
anyway** (or right-click the file → **Properties** → **Unblock**).

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
