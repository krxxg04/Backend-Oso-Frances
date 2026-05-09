param(
  [string]$HealthUrl = $env:HEALTH_URL,
  [string]$DeployHookUrl = $env:RENDER_DEPLOY_HOOK_URL,
  [int]$IntervalSeconds = 300,
  [int]$TimeoutSeconds = 20,
  [int]$FailureThreshold = 3
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

if ([string]::IsNullOrWhiteSpace($HealthUrl)) {
  throw "Define HEALTH_URL o pasa -HealthUrl. Ejemplo: https://tu-servicio.onrender.com/health"
}
if ([string]::IsNullOrWhiteSpace($DeployHookUrl)) {
  throw "Define RENDER_DEPLOY_HOOK_URL o pasa -DeployHookUrl."
}

$failures = 0
Write-Host "Iniciando monitor."
Write-Host "Health: $HealthUrl"
Write-Host "Deploy hook: configurado"
Write-Host "Intervalo: $IntervalSeconds s | Threshold: $FailureThreshold"

while ($true) {
  $now = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
  try {
    $resp = Invoke-WebRequest -Uri $HealthUrl -Method GET -TimeoutSec $TimeoutSeconds -UseBasicParsing
    if ($resp.StatusCode -ge 200 -and $resp.StatusCode -lt 300) {
      if ($failures -gt 0) {
        Write-Host "[$now] Health recuperado ($($resp.StatusCode))."
      } else {
        Write-Host "[$now] Health OK ($($resp.StatusCode))."
      }
      $failures = 0
    } else {
      $failures++
      Write-Warning "[$now] Health fallo HTTP $($resp.StatusCode). Intentos fallidos: $failures/$FailureThreshold"
    }
  } catch {
    $failures++
    Write-Warning "[$now] Health error: $($_.Exception.Message). Intentos fallidos: $failures/$FailureThreshold"
  }

  if ($failures -ge $FailureThreshold) {
    $deployNow = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    try {
      $deployResp = Invoke-WebRequest -Uri $DeployHookUrl -Method POST -TimeoutSec $TimeoutSeconds -UseBasicParsing
      Write-Warning "[$deployNow] Redeploy disparado. Status: $($deployResp.StatusCode)"
    } catch {
      Write-Warning "[$deployNow] No se pudo disparar redeploy: $($_.Exception.Message)"
    }
    $failures = 0
  }

  Start-Sleep -Seconds $IntervalSeconds
}

