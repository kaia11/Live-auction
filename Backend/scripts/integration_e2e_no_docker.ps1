param(
    [string]$BaseUrl = "http://127.0.0.1:8080",
    [string]$RoomId = "room-001",
    [string]$ViewerPhone = "viewer",
    [string]$AnchorPhone = "anchor",
    [string]$AdminPhone = "admin",
    [string]$Password = "123456",
    [int]$SessionDurationSeconds = 600,
    [int]$StartPrice = 100,
    [int]$IncrementStep = 10
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Write-Step {
    param([string]$Text)
    Write-Host ""
    Write-Host ("=" * 80)
    Write-Host $Text
    Write-Host ("=" * 80)
}

function Read-ErrorResponseBody {
    param($ExceptionResponse)
    if ($null -eq $ExceptionResponse) {
        return ""
    }
    $stream = $ExceptionResponse.GetResponseStream()
    if ($null -eq $stream) {
        return ""
    }
    $reader = New-Object IO.StreamReader($stream)
    $content = $reader.ReadToEnd()
    $reader.Close()
    return $content
}

function Invoke-Api {
    param(
        [string]$Method,
        [string]$Path,
        [AllowNull()]$BodyObject,
        [AllowNull()][string]$Token
    )

    $uri = "$BaseUrl$Path"
    $headers = @{}
    if ($Token) {
        $headers["Authorization"] = "Bearer $Token"
    }

    $bodyText = $null
    if ($null -ne $BodyObject) {
        $bodyText = $BodyObject | ConvertTo-Json -Compress -Depth 20
    }

    $status = 0
    $content = ""
    try {
        if ($null -ne $bodyText) {
            $resp = Invoke-WebRequest -UseBasicParsing -Uri $uri -Method $Method -Headers $headers -ContentType "application/json; charset=utf-8" -Body $bodyText
        } else {
            $resp = Invoke-WebRequest -UseBasicParsing -Uri $uri -Method $Method -Headers $headers
        }
        $status = [int]$resp.StatusCode
        $content = $resp.Content
    } catch {
        $res = $_.Exception.Response
        if ($null -eq $res) {
            throw
        }
        $status = [int]$res.StatusCode
        $content = Read-ErrorResponseBody -ExceptionResponse $res
    }

    $json = $null
    if ($content) {
        try {
            $json = $content | ConvertFrom-Json
        } catch {
            $json = $null
        }
    }

    [pscustomobject]@{
        Method  = $Method
        Path    = $Path
        Status  = $status
        BodyRaw = $content
        Body    = $json
    }
}

function Assert-Status {
    param(
        [string]$Name,
        [int[]]$ExpectedStatus,
        $Response
    )

    $ok = $ExpectedStatus -contains $Response.Status
    $expectText = ($ExpectedStatus -join "/")
    $actualText = $Response.Status
    $result = if ($ok) { "PASS" } else { "FAIL" }
    Write-Host ("[{0}] {1} expected={2} actual={3}" -f $result, $Name, $expectText, $actualText)
    if ($Response.BodyRaw) {
        Write-Host $Response.BodyRaw
    }

    [pscustomobject]@{
        Name     = $Name
        Pass     = $ok
        Expected = $expectText
        Actual   = "$actualText"
    }
}

$results = New-Object System.Collections.Generic.List[object]

Write-Step "0) Health check"
$health = Invoke-Api -Method "GET" -Path "/health" -BodyObject $null -Token $null
$results.Add((Assert-Status -Name "GET /health" -ExpectedStatus @(200) -Response $health))

Write-Step "1) Login as viewer / anchor / admin"
$viewerLogin = Invoke-Api -Method "POST" -Path "/auth/login" -BodyObject @{ phone = $ViewerPhone; password = $Password } -Token $null
$results.Add((Assert-Status -Name "POST /auth/login viewer" -ExpectedStatus @(200) -Response $viewerLogin))
$anchorLogin = Invoke-Api -Method "POST" -Path "/auth/login" -BodyObject @{ phone = $AnchorPhone; password = $Password } -Token $null
$results.Add((Assert-Status -Name "POST /auth/login anchor" -ExpectedStatus @(200) -Response $anchorLogin))
$adminLogin = Invoke-Api -Method "POST" -Path "/auth/login" -BodyObject @{ phone = $AdminPhone; password = $Password } -Token $null
$results.Add((Assert-Status -Name "POST /auth/login admin" -ExpectedStatus @(200) -Response $adminLogin))

$viewerToken = $viewerLogin.Body.data.token
$anchorToken = $anchorLogin.Body.data.token
$adminToken = $adminLogin.Body.data.token

if (-not $viewerToken -or -not $anchorToken -or -not $adminToken) {
    throw "Failed to retrieve one or more tokens."
}

Write-Step "2) Permission matrix quick checks"
$rooms = Invoke-Api -Method "GET" -Path "/rooms" -BodyObject $null -Token $null
$results.Add((Assert-Status -Name "GET /rooms anonymous" -ExpectedStatus @(200) -Response $rooms))

$meViewer = Invoke-Api -Method "GET" -Path "/users/me" -BodyObject $null -Token $viewerToken
$results.Add((Assert-Status -Name "GET /users/me viewer" -ExpectedStatus @(200) -Response $meViewer))

$statsAnchor = Invoke-Api -Method "GET" -Path "/admin/stats/overview" -BodyObject $null -Token $anchorToken
$results.Add((Assert-Status -Name "GET /admin/stats/overview anchor" -ExpectedStatus @(200) -Response $statsAnchor))

$statsViewer = Invoke-Api -Method "GET" -Path "/admin/stats/overview" -BodyObject $null -Token $viewerToken
$results.Add((Assert-Status -Name "GET /admin/stats/overview viewer forbidden" -ExpectedStatus @(403) -Response $statsViewer))

Write-Step "3) Create isolated test item"
$createItemBody = @{
    title                   = "integration-e2e-item"
    coverImage              = "https://images.unsplash.com/photo-1617038220319-276d3cfab638?w=800&q=80"
    description             = "created by integration_e2e_no_docker.ps1"
    startPrice              = $StartPrice
    incrementStep           = $IncrementStep
    durationSeconds         = $SessionDurationSeconds
    extensionSeconds        = 30
    extensionTriggerSeconds = 30
}
$createItem = Invoke-Api -Method "POST" -Path "/admin/rooms/$RoomId/items" -BodyObject $createItemBody -Token $anchorToken
$results.Add((Assert-Status -Name "POST /admin/rooms/{roomId}/items" -ExpectedStatus @(200, 201) -Response $createItem))

Write-Step "4) Queue next and start session"
$next = Invoke-Api -Method "POST" -Path "/admin/rooms/$RoomId/queue/next" -BodyObject $null -Token $anchorToken
$results.Add((Assert-Status -Name "POST /admin/rooms/{roomId}/queue/next" -ExpectedStatus @(200) -Response $next))

$sessionId = $next.Body.data.nextSessionId
$itemId = $next.Body.data.nextItemId
if (-not $sessionId -or -not $itemId) {
    throw "Failed to resolve nextSessionId or nextItemId."
}

$start = Invoke-Api -Method "POST" -Path "/admin/sessions/$sessionId/start" -BodyObject $null -Token $anchorToken
$results.Add((Assert-Status -Name "POST /admin/sessions/{sessionId}/start" -ExpectedStatus @(200) -Response $start))

Write-Step "5) Bid flow"
$current = Invoke-Api -Method "GET" -Path "/rooms/$RoomId/current-session" -BodyObject $null -Token $null
$results.Add((Assert-Status -Name "GET /rooms/{roomId}/current-session" -ExpectedStatus @(200) -Response $current))

$currentPrice = [int64]$current.Body.data.currentPrice
$increment = [int64]$current.Body.data.incrementStep
$bidPrice = $currentPrice + $increment
$requestId = "req-" + [DateTimeOffset]::Now.ToUnixTimeMilliseconds()

$bidBody = @{
    roomId    = $RoomId
    sessionId = $sessionId
    itemId    = $itemId
    bidPrice  = $bidPrice
    requestId = $requestId
}

$bid = Invoke-Api -Method "POST" -Path "/bids" -BodyObject $bidBody -Token $viewerToken
$results.Add((Assert-Status -Name "POST /bids first attempt" -ExpectedStatus @(200) -Response $bid))

$ranking = Invoke-Api -Method "GET" -Path "/sessions/$sessionId/ranking" -BodyObject $null -Token $null
$results.Add((Assert-Status -Name "GET /sessions/{sessionId}/ranking" -ExpectedStatus @(200) -Response $ranking))

$myStatus = Invoke-Api -Method "GET" -Path "/sessions/$sessionId/my-status" -BodyObject $null -Token $viewerToken
$results.Add((Assert-Status -Name "GET /sessions/{sessionId}/my-status" -ExpectedStatus @(200) -Response $myStatus))

Write-Step "6) Idempotency check with duplicate requestId"
$dup = Invoke-Api -Method "POST" -Path "/bids" -BodyObject $bidBody -Token $viewerToken
$results.Add((Assert-Status -Name "POST /bids duplicate requestId" -ExpectedStatus @(409) -Response $dup))

Write-Step "7) Final summary"
$passCount = @($results | Where-Object { $_.Pass }).Count
$failCount = @($results | Where-Object { -not $_.Pass }).Count
$results | Format-Table -AutoSize
Write-Host ""
Write-Host ("Total: {0}, Pass: {1}, Fail: {2}" -f $results.Count, $passCount, $failCount)

if ($failCount -gt 0) {
    exit 1
}

exit 0
