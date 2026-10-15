# Arch User Repository

Two packages:

| Directory | AUR name | What it installs |
| --- | --- | --- |
| `axam-bin/` | `axam-bin` | Prebuilt Linux binary from GitHub Releases |
| `axam/` | `axam` | Builds from the release source tarball |

They `conflicts` with each other and both `provides=axam`.

## Publish (first time)

1. Create an AUR account and add your SSH key.
2. For the binary package (fastest for users):

```bash
git clone ssh://aur@aur.archlinux.org/axam-bin.git
cd axam-bin
cp /path/to/axam/packaging/aur/axam-bin/PKGBUILD .
cp /path/to/axam/packaging/aur/axam-bin/.SRCINFO .
git add PKGBUILD .SRCINFO
git commit -m "Initial import: axam 1.1.0"
git push
```

3. Optionally the source package:

```bash
git clone ssh://aur@aur.archlinux.org/axam.git
cd axam
cp /path/to/axam/packaging/aur/axam/PKGBUILD .
cp /path/to/axam/packaging/aur/axam/.SRCINFO .
git add PKGBUILD .SRCINFO
git commit -m "Initial import: axam 1.1.0"
git push
```

On Arch, regenerate `.SRCINFO` instead of copying when you edit:

```bash
makepkg --printsrcinfo > .SRCINFO
```

## Install (users)

```bash
# helper, e.g. yay / paru
yay -S axam-bin
# or
yay -S axam
```

## Update for a new release

1. Bump `pkgver` / reset `pkgrel` to `1`.
2. Refresh `sha256sums*` from the release `SHA256SUMS` (bin) or the tag tarball (source).
3. `makepkg --printsrcinfo > .SRCINFO`
4. Commit and push the AUR git repo.

Note: building `axam` from source needs a Go toolchain new enough for `go 1.26.3` in `go.mod`.
