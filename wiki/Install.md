# Install

Release **1.1.0** uses GitHub tag `v1.1.0`. You do not need a Go toolchain for these paths.

Assets: [https://github.com/alexshescp/axam/releases/tag/v1.1.0](https://github.com/alexshescp/axam/releases/tag/v1.1.0)

## Debian / Ubuntu, 64-bit Intel/AMD

```bash
curl -fL -o axam_1.1.0_amd64.deb \
  https://github.com/alexshescp/axam/releases/download/v1.1.0/axam_1.1.0_amd64.deb
sudo apt install ./axam_1.1.0_amd64.deb
```

## Debian / Ubuntu, 64-bit ARM

```bash
curl -fL -o axam_1.1.0_arm64.deb \
  https://github.com/alexshescp/axam/releases/download/v1.1.0/axam_1.1.0_arm64.deb
sudo apt install ./axam_1.1.0_arm64.deb
```

## Other Linux, 64-bit Intel/AMD

```bash
curl -fL -o axam \
  https://github.com/alexshescp/axam/releases/download/v1.1.0/axam_linux_amd64
chmod +x axam
./axam
```

Optional install:

```bash
sudo install -m 0755 axam /usr/local/bin/axam
```

## Other Linux, 64-bit ARM

```bash
curl -fL -o axam \
  https://github.com/alexshescp/axam/releases/download/v1.1.0/axam_linux_arm64
chmod +x axam
./axam
```

Optional install:

```bash
sudo install -m 0755 axam /usr/local/bin/axam
```

## macOS, Intel

```bash
curl -fL -o axam \
  https://github.com/alexshescp/axam/releases/download/v1.1.0/axam_darwin_amd64
chmod +x axam
xattr -d com.apple.quarantine axam || true
./axam
```

Optional install:

```bash
sudo install -m 0755 axam /usr/local/bin/axam
```

## macOS, Apple silicon

```bash
curl -fL -o axam \
  https://github.com/alexshescp/axam/releases/download/v1.1.0/axam_darwin_arm64
chmod +x axam
xattr -d com.apple.quarantine axam || true
./axam
```

Optional install:

```bash
sudo install -m 0755 axam /usr/local/bin/axam
```

## With a Go toolchain

```bash
go install github.com/alexshescp/axam/cmd/axam@latest
```

Or from source:

```bash
git clone https://github.com/alexshescp/axam.git
cd axam
go test ./...
go vet ./...
go build -o axam ./cmd/axam
```

## After install

```bash
axam version
axam doctor
axam
axam notes1.txt
```

Wire F4 in Midnight Commander: [Midnight Commander](Midnight-Commander).

Update by installing a newer `.deb` or replacing the binary. `config.toml` and `state.json` are user files and are not removed by package upgrades.
