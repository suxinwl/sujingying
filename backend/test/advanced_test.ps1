# Advanced API Test Script
$BASE_URL = "http://localhost:8080/api/v1"
$PHONE = "13900001000"
$PASSWORD = "Test@123"
$PAYPASS = "123456"
$TOKEN = ""
$ORDER_ID = ""

$stats = @{ Total=0; Pass=0; Fail=0 }

function Test-API {
    param([string]$Name, [scriptblock]$Action)
    $script:stats.Total++
    Write-Host "TEST: $Name" -ForegroundColor Yellow -NoNewline
    try {
        & $Action
        Write-Host " [PASS]" -ForegroundColor Green
        $script:stats.Pass++
    } catch {
        Write-Host " [FAIL]" -ForegroundColor Red
        Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Red
        $script:stats.Fail++
    }
}

Write-Host "=== Advanced Function Test ===" -ForegroundColor Cyan
Write-Host ""

# Setup
Write-Host "--- Setup: User Authentication ---" -ForegroundColor Cyan

Test-API "Register User" {
    $body = @{ phone=$PHONE; password=$PASSWORD } | ConvertTo-Json
    try {
        Invoke-RestMethod -Uri "$BASE_URL/auth/register" -Method Post -Body $body -ContentType "application/json" | Out-Null
    } catch {
        if ($_.Exception.Response.StatusCode -ne 400) { throw }
    }
}

Test-API "Login" {
    $body = @{ phone=$PHONE; password=$PASSWORD } | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE_URL/auth/login" -Method Post -Body $body -ContentType "application/json"
    $script:TOKEN = $resp.access_token
    Write-Host "  Token: $($TOKEN.Substring(0,20))..." -ForegroundColor Gray
}

Test-API "Set PayPassword" {
    $headers = @{ Authorization="Bearer $TOKEN" }
    $body = @{ pay_password=$PAYPASS } | ConvertTo-Json
    try {
        Invoke-RestMethod -Uri "$BASE_URL/user/paypass/set" -Method Post -Body $body -ContentType "application/json" -Headers $headers | Out-Null
    } catch { if ($_.Exception.Message -notlike "*already*") { throw } }
}

Write-Host ""
Write-Host "--- Priority 1: Order & Settlement ---" -ForegroundColor Cyan

Test-API "Create Order (will fail without deposit)" {
    $headers = @{ Authorization="Bearer $TOKEN" }
    $body = @{
        type="long_buy"
        locked_price=500.00
        weight_g=100.0
        deposit=10000.00
        pay_password=$PAYPASS
    } | ConvertTo-Json
    
    try {
        $resp = Invoke-RestMethod -Uri "$BASE_URL/orders" -Method Post -Body $body -ContentType "application/json" -Headers $headers
        $script:ORDER_ID = $resp.order_id
        Write-Host "  Order ID: $ORDER_ID" -ForegroundColor Gray
    } catch {
        if ($_.Exception.Message -like "*insufficient*" -or $_.Exception.Message -like "*不足*") {
            Write-Host "  Note: Need deposit first" -ForegroundColor Yellow
            throw "Insufficient deposit (Expected)"
        }
        throw
    }
}

Test-API "Query Order List" {
    $headers = @{ Authorization="Bearer $TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/orders" -Method Get -Headers $headers
    Write-Host "  Total Orders: $($resp.total)" -ForegroundColor Gray
}

Write-Host ""
Write-Host "--- Priority 2: Risk Control ---" -ForegroundColor Cyan

Test-API "Risk Statistics (Price=500)" {
    $headers = @{ Authorization="Bearer $TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/risk/statistics?current_price=500" -Method Get -Headers $headers
    Write-Host "  Total: $($resp.total_orders), Force Close: $($resp.force_close_count)" -ForegroundColor Gray
    Write-Host "  High Risk: $($resp.high_risk_count), Warning: $($resp.warning_count)" -ForegroundColor Gray
}

Test-API "Risk Statistics (Price=400 - Simulate Drop)" {
    $headers = @{ Authorization="Bearer $TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/risk/statistics?current_price=400" -Method Get -Headers $headers
    Write-Host "  After Price Drop: Force Close=$($resp.force_close_count), High Risk=$($resp.high_risk_count)" -ForegroundColor Gray
}

Write-Host ""
Write-Host "--- Priority 3: Notifications ---" -ForegroundColor Cyan

Test-API "Get All Notifications" {
    $headers = @{ Authorization="Bearer $TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/notifications?limit=20" -Method Get -Headers $headers
    Write-Host "  Total Notifications: $($resp.total)" -ForegroundColor Gray
}

Test-API "Get Unread Count" {
    $headers = @{ Authorization="Bearer $TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/notifications/count" -Method Get -Headers $headers
    Write-Host "  Unread: $($resp.count)" -ForegroundColor Gray
}

Test-API "Mark All As Read" {
    $headers = @{ Authorization="Bearer $TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/notifications/read-all" -Method Post -Headers $headers
    Write-Host "  Result: $($resp.message)" -ForegroundColor Gray
}

Write-Host ""
Write-Host "--- Priority 4: Sales Management ---" -ForegroundColor Cyan

Test-API "Sales Ranking (Total Points)" {
    $headers = @{ Authorization="Bearer $TOKEN" }
    $uri = "$BASE_URL/sales/ranking?limit=10&by_month=false"
    $resp = Invoke-RestMethod -Uri $uri -Method Get -Headers $headers
    Write-Host "  Salespersons: $($resp.rankings.Count)" -ForegroundColor Gray
}

Test-API "Sales Ranking (Monthly Points)" {
    $headers = @{ Authorization="Bearer $TOKEN" }
    $uri = "$BASE_URL/sales/ranking?limit=10&by_month=true"
    $resp = Invoke-RestMethod -Uri $uri -Method Get -Headers $headers
    Write-Host "  Active This Month: $($resp.rankings.Count)" -ForegroundColor Gray
}

Write-Host ""
Write-Host "--- Priority 5: Bank Cards ---" -ForegroundColor Cyan

Test-API "Add Bank Card" {
    $headers = @{ Authorization="Bearer $TOKEN" }
    $body = @{
        bank_name="CCB"
        card_number="6217001234567890"
        card_holder="Test User"
        pay_password=$PAYPASS
    } | ConvertTo-Json
    
    try {
        $resp = Invoke-RestMethod -Uri "$BASE_URL/bank-cards" -Method Post -Body $body -ContentType "application/json" -Headers $headers
        Write-Host "  Card: $($resp.card_number)" -ForegroundColor Gray
    } catch {
        if ($_.Exception.Message -like "*5*") {
            Write-Host "  Max cards reached" -ForegroundColor Yellow
        } else { throw }
    }
}

Test-API "List Bank Cards" {
    $headers = @{ Authorization="Bearer $TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/bank-cards" -Method Get -Headers $headers
    Write-Host "  Total Cards: $($resp.cards.Count)" -ForegroundColor Gray
}

Write-Host ""
Write-Host "--- Performance Test ---" -ForegroundColor Cyan

Test-API "Login Response Time" {
    $body = @{ phone=$PHONE; password=$PASSWORD } | ConvertTo-Json
    $sw = [Diagnostics.Stopwatch]::StartNew()
    Invoke-RestMethod -Uri "$BASE_URL/auth/login" -Method Post -Body $body -ContentType "application/json" | Out-Null
    $sw.Stop()
    Write-Host "  Response Time: $($sw.ElapsedMilliseconds)ms" -ForegroundColor $(if($sw.ElapsedMilliseconds -lt 500){"Green"}else{"Yellow"})
}

Test-API "Query Response Time" {
    $headers = @{ Authorization="Bearer $TOKEN" }
    $sw = [Diagnostics.Stopwatch]::StartNew()
    Invoke-RestMethod -Uri "$BASE_URL/orders" -Method Get -Headers $headers | Out-Null
    $sw.Stop()
    Write-Host "  Response Time: $($sw.ElapsedMilliseconds)ms" -ForegroundColor $(if($sw.ElapsedMilliseconds -lt 200){"Green"}else{"Yellow"})
}

# Summary
Write-Host ""
Write-Host "=== Test Summary ===" -ForegroundColor Cyan
Write-Host "Total: $($stats.Total)" -ForegroundColor White
Write-Host "Pass:  $($stats.Pass)" -ForegroundColor Green
Write-Host "Fail:  $($stats.Fail)" -ForegroundColor Red
$rate = [math]::Round(($stats.Pass/$stats.Total)*100, 2)
Write-Host "Pass Rate: $rate%" -ForegroundColor $(if($stats.Fail -eq 0){"Green"}else{"Yellow"})

Write-Host ""
Write-Host "Notes:" -ForegroundColor Cyan
Write-Host "- Order creation requires deposit (manual DB update or deposit approval)" -ForegroundColor Gray
Write-Host "- Risk scheduler runs every 60 seconds automatically" -ForegroundColor Gray
Write-Host "- Commission calculated automatically after order settlement" -ForegroundColor Gray
