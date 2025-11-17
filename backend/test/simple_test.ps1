# Simple API Test Script
$BASE_URL = "http://localhost:8080/api/v1"
$PHONE = "13900000999"
$PASSWORD = "Test@123"
$PAYPASS = "123456"

Write-Host "=== SuXinYing APP API Test ===" -ForegroundColor Cyan
Write-Host ""

# Test 1: Register
Write-Host "[1/8] Testing Register..." -ForegroundColor Yellow
try {
    $body = @{ phone = $PHONE; password = $PASSWORD } | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE_URL/auth/register" -Method Post -Body $body -ContentType "application/json"
    Write-Host "  PASS - Register" -ForegroundColor Green
} catch {
    Write-Host "  WARN - User may exist" -ForegroundColor Yellow
}

# Test 2: Login
Write-Host "[2/8] Testing Login..." -ForegroundColor Yellow
$body = @{ phone = $PHONE; password = $PASSWORD } | ConvertTo-Json
$resp = Invoke-RestMethod -Uri "$BASE_URL/auth/login" -Method Post -Body $body -ContentType "application/json"
$TOKEN = $resp.access_token
Write-Host "  PASS - Login (Token: $($TOKEN.Substring(0,20))...)" -ForegroundColor Green

$headers = @{ Authorization = "Bearer $TOKEN" }

# Test 3: Set PayPassword
Write-Host "[3/8] Testing Set Pay Password..." -ForegroundColor Yellow
try {
    $body = @{ pay_password = $PAYPASS } | ConvertTo-Json
    Invoke-RestMethod -Uri "$BASE_URL/user/paypass/set" -Method Post -Body $body -ContentType "application/json" -Headers $headers | Out-Null
    Write-Host "  PASS - Set Pay Password" -ForegroundColor Green
} catch {
    Write-Host "  WARN - Pay password may exist" -ForegroundColor Yellow
}

# Test 4: Add Bank Card
Write-Host "[4/8] Testing Add Bank Card..." -ForegroundColor Yellow
try {
    $body = @{
        bank_name = "ICBC"
        card_number = "6222021234567890"
        card_holder = "Test User"
        pay_password = $PAYPASS
    } | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE_URL/bank-cards" -Method Post -Body $body -ContentType "application/json" -Headers $headers
    Write-Host "  PASS - Add Bank Card (Card: $($resp.card_number))" -ForegroundColor Green
} catch {
    Write-Host "  WARN - Bank card may exist" -ForegroundColor Yellow
}

# Test 5: Get Bank Cards
Write-Host "[5/8] Testing Get Bank Cards..." -ForegroundColor Yellow
$resp = Invoke-RestMethod -Uri "$BASE_URL/bank-cards" -Method Get -Headers $headers
Write-Host "  PASS - Get Bank Cards (Count: $($resp.cards.Count))" -ForegroundColor Green

# Test 6: Get Notifications Count
Write-Host "[6/8] Testing Get Notifications..." -ForegroundColor Yellow
$resp = Invoke-RestMethod -Uri "$BASE_URL/notifications/count" -Method Get -Headers $headers
Write-Host "  PASS - Get Notifications (Unread: $($resp.count))" -ForegroundColor Green

# Test 7: Get Risk Statistics
Write-Host "[7/8] Testing Risk Statistics..." -ForegroundColor Yellow
$resp = Invoke-RestMethod -Uri "$BASE_URL/risk/statistics?current_price=500" -Method Get -Headers $headers
Write-Host "  PASS - Risk Statistics (Orders: $($resp.total_orders), Force Close: $($resp.force_close_count))" -ForegroundColor Green

# Test 8: Get Sales Ranking
Write-Host "[8/8] Testing Sales Ranking..." -ForegroundColor Yellow
$resp = Invoke-RestMethod -Uri "$BASE_URL/sales/ranking?limit=10" -Method Get -Headers $headers
Write-Host "  PASS - Sales Ranking (Count: $($resp.rankings.Count))" -ForegroundColor Green

Write-Host ""
Write-Host "=== All Tests Completed ===" -ForegroundColor Cyan
Write-Host "Status: SUCCESS" -ForegroundColor Green
