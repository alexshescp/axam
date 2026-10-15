# Homebrew tap

Formula installs the prebuilt GitHub Release binary (no Go toolchain required).

## Tap

Live tap: [github.com/alexshescp/homebrew-axam](https://github.com/alexshescp/homebrew-axam)

Keep `Formula/axam.rb` in that repository in sync with
`packaging/homebrew/axam.rb` in the main axam repo when you cut a release.

## Install

```bash
brew install alexshescp/axam/axam
axam version
```

Or:

```bash
brew tap alexshescp/axam
brew install axam
```

## Update for a new release

1. Rebuild and publish GitHub Release assets.
2. Refresh `sha256` values from that release’s `SHA256SUMS`.
3. Bump `version` in `Formula/axam.rb`.
4. Commit and push the tap.

```bash
curl -fsSL https://github.com/alexshescp/axam/releases/download/vX.Y.Z/SHA256SUMS
```

## homebrew-core later

A source formula (`go build` from the tag tarball) is better for a core PR.
Source archive for `v1.1.0`:

```
https://github.com/alexshescp/axam/archive/refs/tags/v1.1.0.tar.gz
sha256: 81cbdbf16e0ae620512548c38126b6584455f4ada1355c382b5d73fa1b83ebdf
```
