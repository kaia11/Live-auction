param(
    [string]$BaseUrl = "http://127.0.0.1:8080",
    [ValidateSet("single", "multi")]
    [string]$Mode = "single",
    [int]$Concurrency = 5,
    [int]$RequestsPerWorker = 20,
    [string[]]$RoomIds = @("room-001"),
    [string]$SessionId = "session-001",
    [string]$ItemId = "item-001",
    [int]$StartPrice = 140,
    [switch]$CheckHealth,
    [switch]$SnapshotMetrics
)

$login = Invoke-RestMethod -Uri "$BaseUrl/auth/login" -Method POST -ContentType "application/json" -Body '{"phone":"viewer","password":"demo"}'
$token = $login.data.token
if (-not $token) {
    throw "failed to get token"
}

$jobs = @()
for ($worker = 0; $worker -lt $Concurrency; $worker++) {
    $roomId = $RoomIds[0]
    if ($Mode -eq "multi") {
        $roomId = $RoomIds[$worker % $RoomIds.Count]
    }

    $jobs += Start-Job -ScriptBlock {
        param($BaseUrl, $Token, $Worker, $RequestsPerWorker, $RoomId, $SessionId, $ItemId, $StartPrice)
        $success = 0
        $failed = 0
        for ($i = 0; $i -lt $RequestsPerWorker; $i++) {
            $price = $StartPrice + ($Worker * 100) + $i
            $body = @{
                roomId    = $RoomId
                sessionId = $SessionId
                itemId    = $ItemId
                bidPrice  = $price
                requestId = "load-$Worker-$i"
            } | ConvertTo-Json

            try {
                Invoke-RestMethod -Uri "$BaseUrl/bids" -Method POST -Headers @{ Authorization = "Bearer $Token" } -ContentType "application/json" -Body $body | Out-Null
                $success++
            } catch {
                $failed++
            }
        }
        [pscustomobject]@{
            Worker  = $Worker
            Success = $success
            Failed  = $failed
        }
    } -ArgumentList $BaseUrl, $token, $worker, $RequestsPerWorker, $roomId, $SessionId, $ItemId, $StartPrice
}

$jobs | Wait-Job | Out-Null
$workerResults = $jobs | Receive-Job
$jobs | Remove-Job | Out-Null

$totalSuccess = ($workerResults | Measure-Object -Property Success -Sum).Sum
$totalFailed = ($workerResults | Measure-Object -Property Failed -Sum).Sum
$totalRequests = $Concurrency * $RequestsPerWorker

Write-Host "load test completed mode=$Mode concurrency=$Concurrency requests_per_worker=$RequestsPerWorker total_requests=$totalRequests success=$totalSuccess failed=$totalFailed"

if ($CheckHealth) {
    try {
        $health = Invoke-RestMethod -Uri "$BaseUrl/health" -Method GET
        Write-Host "health ok=$($health.data.ok) persistentLedgerRequired=$($health.data.persistentLedgerRequired) mysqlConfigured=$($health.data.mysqlConfigured)"
    } catch {
        Write-Warning "failed to fetch /health: $($_.Exception.Message)"
    }
}

if ($SnapshotMetrics) {
    try {
        $metrics = Invoke-WebRequest -Uri "$BaseUrl/metrics" -Method GET
        Write-Host "metrics snapshot:"
        $metrics.Content -split "`n" |
            Where-Object { $_ -match '^auction_(bid|settlements|ws|errors)' } |
            ForEach-Object { Write-Host $_ }
    } catch {
        Write-Warning "failed to fetch /metrics: $($_.Exception.Message)"
    }
}
