# Test Login API
Write-Host "Testing Login API..." -ForegroundColor Cyan

$body = @{
    username = "13800000001"
    password = "123456"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" `
        -Method POST `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $body
    
    Write-Host "`n✅ Login Success!" -ForegroundColor Green
    Write-Host "`nFull Response:" -ForegroundColor Cyan
    $response | ConvertTo-Json -Depth 3
    
    if ($response.access_token) {
        Write-Host "`nToken: $($response.access_token.Substring(0, 30))..." -ForegroundColor Gray
    }
} catch {
    Write-Host "`n❌ Login Failed!" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Yellow
    
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response Body: $responseBody" -ForegroundColor Yellow
    }
}
