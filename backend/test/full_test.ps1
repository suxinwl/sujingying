# 速金盈APP 完整功能测试脚本
# 测试所有核心模块功能

$BASE_URL = "http://localhost:8080/api/v1"
$ACCESS_TOKEN = ""
$TEST_PHONE = "13900000999"
$TEST_PASSWORD = "Test@123"
$TEST_PAYPASS = "123456"

Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║         速金盈APP 完整功能测试                              ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

$testResults = @()

# 测试函数
function Test-API {
    param(
        [string]$Name,
        [scriptblock]$TestBlock
    )
    
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
    Write-Host "🧪 测试: $Name" -ForegroundColor Yellow
    
    try {
        & $TestBlock
        Write-Host "✅ $Name - 通过" -ForegroundColor Green
        $script:testResults += @{Name=$Name; Status="PASS"}
    }
    catch {
        Write-Host "❌ $Name - 失败: $($_.Exception.Message)" -ForegroundColor Red
        $script:testResults += @{Name=$Name; Status="FAIL"; Error=$_.Exception.Message}
    }
    Write-Host ""
}

# ============================================================
# 模块1: 用户认证测试
# ============================================================

Write-Host "┌──────────────────────────────────────────────────────────┐" -ForegroundColor Cyan
Write-Host "│  模块1: 用户认证与安全                                   │" -ForegroundColor Cyan
Write-Host "└──────────────────────────────────────────────────────────┘" -ForegroundColor Cyan
Write-Host ""

Test-API "1.1 用户注册" {
    $body = @{
        phone = $TEST_PHONE
        password = $TEST_PASSWORD
    } | ConvertTo-Json
    
    try {
        $resp = Invoke-RestMethod -Uri "$BASE_URL/auth/register" -Method Post -Body $body -ContentType "application/json"
        Write-Host "   注册成功: $($resp.message)" -ForegroundColor Gray
    }
    catch {
        if ($_.Exception.Response.StatusCode -eq 400) {
            Write-Host "   用户已存在，跳过注册" -ForegroundColor Gray
        }
        else { throw }
    }
}

Test-API "1.2 用户登录" {
    $body = @{
        phone = $TEST_PHONE
        password = $TEST_PASSWORD
    } | ConvertTo-Json
    
    $resp = Invoke-RestMethod -Uri "$BASE_URL/auth/login" -Method Post -Body $body -ContentType "application/json"
    $script:ACCESS_TOKEN = $resp.access_token
    Write-Host "   用户ID: $($resp.user.id)" -ForegroundColor Gray
    Write-Host "   Token: $($ACCESS_TOKEN.Substring(0,30))..." -ForegroundColor Gray
}

Test-API "1.3 设置支付密码" {
    $headers = @{ Authorization = "Bearer $ACCESS_TOKEN" }
    $body = @{ pay_password = $TEST_PAYPASS } | ConvertTo-Json
    
    try {
        $resp = Invoke-RestMethod -Uri "$BASE_URL/user/paypass/set" -Method Post -Body $body -ContentType "application/json" -Headers $headers
        Write-Host "   支付密码设置成功" -ForegroundColor Gray
    }
    catch {
        if ($_.Exception.Message -like "*已设置*") {
            Write-Host "   支付密码已存在" -ForegroundColor Gray
        }
        else { throw }
    }
}

# ============================================================
# 模块2: 银行卡管理测试
# ============================================================

Write-Host "┌──────────────────────────────────────────────────────────┐" -ForegroundColor Cyan
Write-Host "│  模块2: 银行卡管理                                       │" -ForegroundColor Cyan
Write-Host "└──────────────────────────────────────────────────────────┘" -ForegroundColor Cyan
Write-Host ""

Test-API "2.1 添加银行卡" {
    $headers = @{ Authorization = "Bearer $ACCESS_TOKEN" }
    $body = @{
        bank_name = "中国工商银行"
        card_number = "6222021234567890"
        card_holder = "测试用户"
        pay_password = $TEST_PAYPASS
    } | ConvertTo-Json
    
    $resp = Invoke-RestMethod -Uri "$BASE_URL/bank-cards" -Method Post -Body $body -ContentType "application/json" -Headers $headers
    Write-Host "   卡号: $($resp.card_number)" -ForegroundColor Gray
    Write-Host "   默认卡: $($resp.is_default)" -ForegroundColor Gray
}

Test-API "2.2 查询银行卡列表" {
    $headers = @{ Authorization = "Bearer $ACCESS_TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/bank-cards" -Method Get -Headers $headers
    Write-Host "   银行卡数量: $($resp.cards.Count)" -ForegroundColor Gray
}

# ============================================================
# 模块3: 通知系统测试
# ============================================================

Write-Host "┌──────────────────────────────────────────────────────────┐" -ForegroundColor Cyan
Write-Host "│  模块3: 通知系统                                         │" -ForegroundColor Cyan
Write-Host "└──────────────────────────────────────────────────────────┘" -ForegroundColor Cyan
Write-Host ""

Test-API "3.1 查询未读通知数量" {
    $headers = @{ Authorization = "Bearer $ACCESS_TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/notifications/count" -Method Get -Headers $headers
    Write-Host "   未读通知: $($resp.count) 条" -ForegroundColor Gray
}

Test-API "3.2 查询通知列表" {
    $headers = @{ Authorization = "Bearer $ACCESS_TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/notifications?limit=5" -Method Get -Headers $headers
    Write-Host "   通知总数: $($resp.total)" -ForegroundColor Gray
}

# ============================================================
# 模块4: 风控统计测试
# ============================================================

Write-Host "┌──────────────────────────────────────────────────────────┐" -ForegroundColor Cyan
Write-Host "│  模块4: 风控引擎                                         │" -ForegroundColor Cyan
Write-Host "└──────────────────────────────────────────────────────────┘" -ForegroundColor Cyan
Write-Host ""

Test-API "4.1 查询风控统计" {
    $headers = @{ Authorization = "Bearer $ACCESS_TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/risk/statistics?current_price=500" -Method Get -Headers $headers
    Write-Host "   总订单数: $($resp.total_orders)" -ForegroundColor Gray
    Write-Host "   强平订单: $($resp.force_close_count)" -ForegroundColor Gray
    Write-Host "   高风险订单: $($resp.high_risk_count)" -ForegroundColor Gray
    Write-Host "   预警订单: $($resp.warning_count)" -ForegroundColor Gray
}

# ============================================================
# 模块5: 销售看板测试
# ============================================================

Write-Host "┌──────────────────────────────────────────────────────────┐" -ForegroundColor Cyan
Write-Host "│  模块5: 销售管理                                         │" -ForegroundColor Cyan
Write-Host "└──────────────────────────────────────────────────────────┘" -ForegroundColor Cyan
Write-Host ""

Test-API "5.1 查询销售排行榜" {
    $headers = @{ Authorization = "Bearer $ACCESS_TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/sales/ranking?limit=10" -Method Get -Headers $headers
    Write-Host "   排行榜人数: $($resp.rankings.Count)" -ForegroundColor Gray
}

# ============================================================
# 测试报告
# ============================================================

Write-Host ""
Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                     测试报告                               ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

$passCount = ($testResults | Where-Object { $_.Status -eq "PASS" }).Count
$failCount = ($testResults | Where-Object { $_.Status -eq "FAIL" }).Count
$totalCount = $testResults.Count

Write-Host "📊 测试统计:" -ForegroundColor White
Write-Host "   总测试数: $totalCount" -ForegroundColor Gray
Write-Host "   通过: $passCount" -ForegroundColor Green
Write-Host "   失败: $failCount" -ForegroundColor Red
Write-Host "   通过率: $([math]::Round($passCount/$totalCount*100, 2))%" -ForegroundColor $(if($failCount -eq 0){"Green"}else{"Yellow"})
Write-Host ""

if ($failCount -gt 0) {
    Write-Host "❌ 失败的测试:" -ForegroundColor Red
    $testResults | Where-Object { $_.Status -eq "FAIL" } | ForEach-Object {
        Write-Host "   - $($_.Name): $($_.Error)" -ForegroundColor Red
    }
}
else {
    Write-Host "🎉 所有测试通过！" -ForegroundColor Green
}

Write-Host ""
Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                 测试完成！                                 ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
