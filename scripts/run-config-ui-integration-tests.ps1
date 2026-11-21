param([ValidateRange(1, 1440)][int]$TimeoutMinutes = 2)
& (Join-Path $PSScriptRoot 'run-integration-tests.ps1') -TimeoutMinutes $TimeoutMinutes -TestPattern '^TestConfigUI_'
