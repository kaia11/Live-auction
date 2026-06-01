param(
    [string]$BaseUrl = "http://127.0.0.1:8080",
    [ValidateSet("single", "multi")]
    [string]$Mode = "single",
    [int]$Concurrency = 5,
    [int]$RequestsPerWorker = 20,
    [string[]]$RoomIds = @("room-001"),
    [string]$SessionId = "session-001",
    [string]$ItemId = "item-001",
    [int]$StartPrice = 140
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
            } catch {
            }
        }
    } -ArgumentList $BaseUrl, $token, $worker, $RequestsPerWorker, $roomId, $SessionId, $ItemId, $StartPrice
}

$jobs | Wait-Job | Out-Null
$jobs | Remove-Job | Out-Null

Write-Host "load test completed mode=$Mode concurrency=$Concurrency requests_per_worker=$RequestsPerWorker"
