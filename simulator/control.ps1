param(
    [Parameter(Position = 0)]
    [ValidateSet("start", "stop", "restart", "status", "logs")]
    [string]$Action = "start",

    [ValidateSet("cycle", "normal", "degradation", "failure")]
    [string]$Mode = "cycle",

    [ValidateRange(0.0, 1.0)]
    [double]$Severity = 1.0,

    [ValidateRange(0.1, 3600.0)]
    [double]$Interval = 5.0,

    [ValidateRange(1, 3600)]
    [int]$RefreshInterval = 15,

    [ValidateRange(0, 1000000)]
    [int]$NormalReadings = 45,

    [ValidateRange(0, 1000000)]
    [int]$DegradationReadings = 35,

    [ValidateRange(0, 1000000)]
    [int]$FailureReadings = 30
)

$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent $PSScriptRoot
$culture = [System.Globalization.CultureInfo]::InvariantCulture

function Invoke-Compose {
    param([string[]]$Arguments)
    & docker compose --profile simulator @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "O comando Docker falhou com código $LASTEXITCODE."
    }
}

Push-Location $projectRoot
try {
    switch ($Action) {
        "start" {
            $env:SIMULATION_MODE = $Mode
            $env:SIMULATION_SEVERITY = $Severity.ToString($culture)
            $env:SIMULATION_INTERVAL = $Interval.ToString($culture)
            $env:MACHINE_REFRESH_INTERVAL = $RefreshInterval.ToString($culture)
            $env:SIMULATION_NORMAL_READINGS = $NormalReadings.ToString($culture)
            $env:SIMULATION_DEGRADATION_READINGS = $DegradationReadings.ToString($culture)
            $env:SIMULATION_FAILURE_READINGS = $FailureReadings.ToString($culture)
            Invoke-Compose @("up", "-d", "--build", "--force-recreate", "simulator")
            Write-Host "Simulador iniciado: modo=$Mode intensidade=$Severity intervalo=${Interval}s"
        }
        "restart" {
            $env:SIMULATION_MODE = $Mode
            $env:SIMULATION_SEVERITY = $Severity.ToString($culture)
            $env:SIMULATION_INTERVAL = $Interval.ToString($culture)
            $env:MACHINE_REFRESH_INTERVAL = $RefreshInterval.ToString($culture)
            $env:SIMULATION_NORMAL_READINGS = $NormalReadings.ToString($culture)
            $env:SIMULATION_DEGRADATION_READINGS = $DegradationReadings.ToString($culture)
            $env:SIMULATION_FAILURE_READINGS = $FailureReadings.ToString($culture)
            Invoke-Compose @("up", "-d", "--build", "--force-recreate", "simulator")
            Write-Host "Simulador reiniciado: modo=$Mode intensidade=$Severity intervalo=${Interval}s"
        }
        "stop" {
            Invoke-Compose @("stop", "simulator")
            Write-Host "Simulador parado. Os demais serviços continuam ativos."
        }
        "status" {
            Invoke-Compose @("ps", "simulator")
        }
        "logs" {
            Invoke-Compose @("logs", "-f", "simulator")
        }
    }
} finally {
    Pop-Location
}
