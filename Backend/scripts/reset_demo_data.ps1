param(
    [string]$Distro = ""
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$wslRoot = $root -replace "\\", "/"
$wslRoot = "/mnt/" + $wslRoot.Substring(0,1).ToLower() + $wslRoot.Substring(2)

$prefix = @()
if ($Distro -ne "") {
    $prefix = @("-d", $Distro)
}

$command = "cd '$wslRoot' && sh ./Backend/scripts/reset_demo_data.sh"
wsl @prefix -- bash -lc $command
