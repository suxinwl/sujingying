# Test Unified Response Format
Write-Host "测试统一响应格式" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""

# Test 1: Customer Login
Write-Host "测试1: 普通客户登录" -ForegroundColor Yellow
Write-Host "-------------------------------"
$body = @{
    username = "13800000001"
    password = "123456"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" `
        -Method POST `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $body
    
    Write-Host "✅ 登录成功!" -ForegroundColor Green
    Write-Host "Response结构:" -ForegroundColor Cyan
    $response | ConvertTo-Json -Depth 3
    
    if ($response.data) {
        Write-Host "`n✅ 响应包含data字段 (统一格式)" -ForegroundColor Green
        $token = $response.data.access_token
        Write-Host "Token: $($token.Substring(0, 30))..." -ForegroundColor Gray
        
        # Save token for next test
        $script:customerToken = $token
    } else {
        Write-Host "`n❌ 响应未包含data字段 (旧格式)" -ForegroundColor Red
    }
} catch {
    Write-Host "❌ 登录失败" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Yellow
}

Write-Host "`n"

# Test 2: Admin Login
Write-Host "测试2: 超级管理员登录" -ForegroundColor Yellow
Write-Host "-------------------------------"
$body2 = @{
    username = "13900000000"
    password = "admin123"
} | ConvertTo-Json

try {
    $response2 = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" `
        -Method POST `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $body2
    
    Write-Host "✅ 管理员登录成功!" -ForegroundColor Green
    Write-Host "Response结构:" -ForegroundColor Cyan
    $response2 | ConvertTo-Json -Depth 3
    
    if ($response2.data) {
        Write-Host "`n✅ 响应包含data字段 (统一格式)" -ForegroundColor Green
        Write-Host "用户角色: $($response2.data.user.role)" -ForegroundColor Magenta
    }
} catch {
    Write-Host "❌ 管理员登录失败" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Yellow
}

Write-Host "`n"
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "测试账号信息:" -ForegroundColor Cyan
Write-Host "-------------------------------"
Write-Host "【普通客户】" -ForegroundColor White
Write-Host "  用户名: 13800000001" -ForegroundColor Gray
Write-Host "  密码: 123456" -ForegroundColor Gray
Write-Host "  角色: customer" -ForegroundColor Gray
Write-Host ""
Write-Host "【超级管理员】" -ForegroundColor White
Write-Host "  用户名: 13900000000" -ForegroundColor Gray
Write-Host "  密码: admin123" -ForegroundColor Gray
Write-Host "  角色: super_admin" -ForegroundColor Gray
Write-Host "===============================================" -ForegroundColor Cyan
