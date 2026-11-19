$ErrorActionPreference = "Stop"
$envPath = Join-Path $PSScriptRoot ".env"

if (Test-Path -LiteralPath $envPath) {
    Write-Host ".env já existe. Nenhuma alteração foi feita."
    exit 0
}

function New-Secret {
    param([int]$Bytes)
    $buffer = New-Object byte[] $Bytes
    $generator = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    try {
        $generator.GetBytes($buffer)
    } finally {
        $generator.Dispose()
    }
    return [Convert]::ToBase64String($buffer).TrimEnd("=").Replace("+", "-").Replace("/", "_")
}

$content = @"
DB_USER=admin
DB_PASSWORD=$(New-Secret 24)
DB_NAME=industrial_monitoring
JWT_SECRET=$(New-Secret 48)
AI_SERVICE_TIMEOUT=5s
SIMULATOR_TOKEN=$(New-Secret 48)
SIMULATION_INTERVAL=5
MACHINE_REFRESH_INTERVAL=15
TELEMETRY_RETENTION_DAYS=30
"@

[IO.File]::WriteAllText($envPath, $content, [Text.UTF8Encoding]::new($false))
Write-Host ".env criado com credenciais aleatórias fortes."
