param(
    [Parameter(Mandatory = $true)]
    [string]$ImageRepository,

    [string]$Tag,

    [switch]$AlsoTagLatest
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$versionFile = Join-Path $repoRoot "VERSION"

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    throw "docker is required"
}

$gitAvailable = Get-Command git -ErrorAction SilentlyContinue
$shortSha = "nogit"
if ($gitAvailable) {
    try {
        $shortSha = (git -c safe.directory="$repoRoot" -C $repoRoot rev-parse --short HEAD).Trim()
    } catch {
        $shortSha = "nogit"
    }
}

if ([string]::IsNullOrWhiteSpace($Tag)) {
    $Tag = "{0}-{1}" -f (Get-Date -Format "yyyyMMdd-HHmmss"), $shortSha
}

$image = "{0}:{1}" -f $ImageRepository, $Tag
$latestImage = "{0}:latest" -f $ImageRepository

$previousVersion = ""
if (Test-Path $versionFile) {
    $previousVersion = Get-Content $versionFile -Raw
}

try {
    Set-Content -LiteralPath $versionFile -Value "$Tag`n" -NoNewline

    Write-Host "Building $image"
    docker build --pull -t $image $repoRoot

    if ($AlsoTagLatest) {
        docker tag $image $latestImage
    }

    Write-Host "Pushing $image"
    docker push $image

    if ($AlsoTagLatest) {
        Write-Host "Pushing $latestImage"
        docker push $latestImage
    }
}
finally {
    Set-Content -LiteralPath $versionFile -Value $previousVersion -NoNewline
}

Write-Host ""
Write-Host "Release image:"
Write-Host "  $image"
Write-Host ""
Write-Host "Server update commands:"
Write-Host "  export NEW_API_IMAGE=$image"
Write-Host "  docker-compose --env-file deploy/.env.prod -f deploy/compose.prod.yml pull new-api"
Write-Host "  docker-compose --env-file deploy/.env.prod -f deploy/compose.prod.yml up -d new-api"
