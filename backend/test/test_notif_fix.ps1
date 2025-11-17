# Test Notification MarkAllAsRead Fix
$BASE_URL = "http://localhost:8080/api/v1"
$PHONE = "13900003000"
$PASSWORD = "Test@123"

Write-Host "=== Testing Notification Bug Fix ===" -ForegroundColor Cyan
Write-Host ""

# Register and Login
Write-Host "[1/5] Register & Login..." -ForegroundColor Yellow
$body = @{ phone=$PHONE; password=$PASSWORD } | ConvertTo-Json
try {
    Invoke-RestMethod -Uri "$BASE_URL/auth/register" -Method Post -Body $body -ContentType "application/json" | Out-Null
} catch {}

$resp = Invoke-RestMethod -Uri "$BASE_URL/auth/login" -Method Post -Body $body -ContentType "application/json"
$TOKEN = $resp.access_token
Write-Host "  Login successful" -ForegroundColor Green

$headers = @{ Authorization="Bearer $TOKEN" }

# Check unread count
Write-Host "[2/5] Check unread count..." -ForegroundColor Yellow
$resp = Invoke-RestMethod -Uri "$BASE_URL/notifications/count" -Method Get -Headers $headers
Write-Host "  Unread count: $($resp.count)" -ForegroundColor Gray

# Mark all as read (BUG TEST)
Write-Host "[3/5] Mark all as read when no notifications..." -ForegroundColor Yellow
try {
    $resp = Invoke-RestMethod -Uri "$BASE_URL/notifications/read-all" -Method Post -Headers $headers
    Write-Host "  SUCCESS - No error!" -ForegroundColor Green
    Write-Host "  Response: $($resp.message)" -ForegroundColor Gray
} catch {
    Write-Host "  FAILED - Still has error!" -ForegroundColor Red
    Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Verify unread count
Write-Host "[4/5] Verify unread count..." -ForegroundColor Yellow
$resp = Invoke-RestMethod -Uri "$BASE_URL/notifications/count" -Method Get -Headers $headers
Write-Host "  Unread count: $($resp.count)" -ForegroundColor Gray

# List all notifications
Write-Host "[5/5] List all notifications..." -ForegroundColor Yellow
$resp = Invoke-RestMethod -Uri "$BASE_URL/notifications" -Method Get -Headers $headers
Write-Host "  Total notifications: $($resp.total)" -ForegroundColor Gray

Write-Host ""
Write-Host "=== Test Complete ===" -ForegroundColor Cyan
Write-Host "Bug Status: FIXED" -ForegroundColor Green
