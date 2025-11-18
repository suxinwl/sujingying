# Quick API Test Script

$baseUrl = "http://localhost:8080/api/v1"

Write-Host "Testing Backend API..." -ForegroundColor Cyan

# Test 1: Register admin user
Write-Host "`n1. Registering admin user..." -ForegroundColor Yellow
try {
    $body = '{"username":"admin","password":"123456","real_name":"Admin","phone":"13800000001","invite_code":"SYSTEM"}'
    $response = Invoke-WebRequest -Uri "$baseUrl/auth/register" -Method POST -Headers @{"Content-Type"="application/json"} -Body $body
    Write-Host "Admin registered successfully" -ForegroundColor Green
} catch {
    Write-Host "Admin may already exist (this is OK)" -ForegroundColor Yellow
}

# Test 2: Register customer user  
Write-Host "`n2. Registering customer user..." -ForegroundColor Yellow
try {
    $body = '{"username":"customer","password":"123456","real_name":"Customer","phone":"13800000002","invite_code":"TEST2025"}'
    $response = Invoke-WebRequest -Uri "$baseUrl/auth/register" -Method POST -Headers @{"Content-Type"="application/json"} -Body $body
    Write-Host "Customer registered successfully" -ForegroundColor Green
} catch {
    Write-Host "Customer may already exist (this is OK)" -ForegroundColor Yellow
}

# Test 3: Try to login
Write-Host "`n3. Testing login..." -ForegroundColor Yellow
try {
    $body = '{"username":"admin","password":"123456"}'
    $response = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Headers @{"Content-Type"="application/json"} -Body $body
    Write-Host "Login successful!" -ForegroundColor Green
    Write-Host "Token: $($response.data.access_token.Substring(0,20))..." -ForegroundColor Gray
} catch {
    Write-Host "Login failed - User may need approval" -ForegroundColor Yellow
}

Write-Host "`n================================================" -ForegroundColor Cyan
Write-Host "Test Accounts:" -ForegroundColor Cyan
Write-Host "Username: admin    Password: 123456" -ForegroundColor White
Write-Host "Username: customer Password: 123456" -ForegroundColor White
Write-Host "`nFrontend: http://localhost:5173" -ForegroundColor Green
Write-Host "Backend:  http://localhost:8080" -ForegroundColor Green
Write-Host "================================================`n" -ForegroundColor Cyan
