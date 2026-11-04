[CmdletBinding()]
param(
    [string]$Config,
    [string]$Operation = 'listen',
    [string[]]$Ids = @(),
    [object]$Body,
    [string]$JsonFile,
    [string]$ImageFile,
    [switch]$Write,
    [switch]$List,
    [switch]$Check
)

$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '../..')).Path
$hasBody = $PSBoundParameters.ContainsKey('Body')
if ($hasBody -and $JsonFile) {
    throw 'Use either -Body or -JsonFile.'
}
if (-not $Config) {
    $Config = Join-Path $repo 'config.yaml'
}
if ($ImageFile) {
    $ImageFile = (Resolve-Path -LiteralPath $ImageFile).Path
}
if ($JsonFile) {
    $JsonFile = (Resolve-Path -LiteralPath $JsonFile).Path
}
if (-not $List) {
    $Config = (Resolve-Path -LiteralPath $Config).Path
}

$tempFile = $null
$pushed = $false
try {
    if ($hasBody) {
        $tempFile = [IO.Path]::GetTempFileName()
        $json = ConvertTo-Json -InputObject $Body -Depth 30 -Compress
        [IO.File]::WriteAllText($tempFile, $json, [Text.UTF8Encoding]::new($false))
        $JsonFile = $tempFile
    }
    $goArgs = @('run', './examples/management', '-config', $Config, '-op', $Operation)
    if ($List) { $goArgs += '-list' }
    if ($Check) { $goArgs += '-check' }
    if ($Write) { $goArgs += '-write' }
    if ($ImageFile) { $goArgs += @('-image-file', $ImageFile) }
    if ($JsonFile) { $goArgs += @('-input', $JsonFile) }
    $goArgs += '--'
    $goArgs += $Ids

    Push-Location -LiteralPath $repo
    $pushed = $true
    & go @goArgs
    $runExitCode = $LASTEXITCODE
}
finally {
    if ($pushed) { Pop-Location }
    if ($tempFile) { Remove-Item -LiteralPath $tempFile -ErrorAction SilentlyContinue }
}
exit $runExitCode
