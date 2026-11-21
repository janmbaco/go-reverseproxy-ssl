param([ValidateRange(1, 1440)][int]$TimeoutMinutes = 10, [string]$TestPattern = '.')
$ErrorActionPreference = 'Stop'
Push-Location (Join-Path $PSScriptRoot '..')
$compose = @('compose', '--project-name', "reverseproxy-tests-$PID", '-f', 'integration/docker-compose.test.yml')
$previous = $env:INTEGRATION_TEST
try {
    & docker @compose up --detach --build
    if ($LASTEXITCODE -ne 0) { throw 'Compose startup failed' }
    foreach ($url in @('https://localhost/', 'http://localhost:8081/api/config')) {
        $ready = $false
        for ($attempt = 0; $attempt -lt 40; $attempt++) {
            & curl.exe --noproxy '*' --insecure --fail --silent --max-time 2 $url -o NUL
            if ($LASTEXITCODE -eq 0) { $ready = $true; break }
            Start-Sleep -Seconds 1
        }
        if (-not $ready) { throw "Service did not become ready at $url" }
    }
    $env:INTEGRATION_TEST = 'true'
    & go test -run $TestPattern -count=1 "-timeout=${TimeoutMinutes}m" -v ./integration/...
    if ($LASTEXITCODE -ne 0) { throw 'Integration tests failed' }
} catch {
    & docker @compose logs --no-color
    throw
} finally {
    $env:INTEGRATION_TEST = $previous
    & docker @compose down --volumes --remove-orphans
    Pop-Location
}
