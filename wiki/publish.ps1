# One-time: open https://github.com/alexshescp/axam/wiki , click
# "Create the first page", save an empty Home, then run this script.
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $MyInvocation.MyCommand.Path
# Avoid 8.3 TEMP paths (e.g. AD0B8~1.SHA) that break Remove-Item on this host.
$Dest = "C:\Dev\axam-wiki-publish"
if (Test-Path -LiteralPath $Dest) {
    Remove-Item -LiteralPath $Dest -Recurse -Force
}
git clone https://github.com/alexshescp/axam.wiki.git $Dest
if ($LASTEXITCODE -ne 0) {
    throw @"
Wiki git repository is not ready yet.
Open https://github.com/alexshescp/axam/wiki , click Create the first page,
save Home once, then run this script again.
"@
}
Get-ChildItem -LiteralPath $Root -Filter "*.md" | ForEach-Object {
    Copy-Item -LiteralPath $_.FullName -Destination $Dest -Force
}
Push-Location -LiteralPath $Dest
try {
    git add .
    $pending = git status --porcelain
    if (-not $pending) {
        Write-Host "Wiki already up to date."
        return
    }
    git -c user.name="Aleksandr Shaman" -c user.email="axam@codcl.com" commit -m "Publish Axam Studio wiki"
    if ($LASTEXITCODE -ne 0) { throw "git commit failed" }
    git push
    if ($LASTEXITCODE -ne 0) { throw "git push failed" }
} finally {
    Pop-Location
}
Write-Host "Wiki: https://github.com/alexshescp/axam/wiki"
