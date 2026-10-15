# Build axam on Windows. Cross-compiles Linux and macOS binaries and local .deb packages.
$ErrorActionPreference = "Stop"
$Version = "1.1.0"
$Ldflags = "-s -w -X main.version=$Version"
New-Item -ItemType Directory -Force -Path dist | Out-Null

function Build-Target($goos, $goarch, $out) {
    $env:CGO_ENABLED = "0"
    $env:GOOS = $goos
    $env:GOARCH = $goarch
    go build -trimpath -ldflags $Ldflags -o $out ./cmd/axam
    if ($LASTEXITCODE -ne 0) { throw "build failed for $goos/$goarch" }
}

Build-Target linux amd64 dist/axam_linux_amd64
Build-Target linux arm64 dist/axam_linux_arm64
Build-Target darwin amd64 dist/axam_darwin_amd64
Build-Target darwin arm64 dist/axam_darwin_arm64
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
$env:CGO_ENABLED = "0"
go build -trimpath -ldflags $Ldflags -o dist/axam.exe ./cmd/axam

go run packaging/mkdeb.go -bin dist/axam_linux_amd64 -arch amd64 -version $Version -out "dist/axam_${Version}_amd64.deb"
go run packaging/mkdeb.go -bin dist/axam_linux_arm64 -arch arm64 -version $Version -out "dist/axam_${Version}_arm64.deb"
Write-Host "Built dist/"
