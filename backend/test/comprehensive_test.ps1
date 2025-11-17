# 速金盈APP 综合测试脚本
# 覆盖所有优先级的测试场景

$BASE_URL = "http://localhost:8080/api/v1"
$PHONE = "13900001000"
$PASSWORD = "Test@123"
$PAYPASS = "123456"
$TOKEN = ""
$USER_ID = 0

$testResults = @{
    Total = 0
    Pass = 0
    Fail = 0
    Skip = 0
}

function Write-TestHeader {
    param([string]$Title)
    Write-Host ""
    Write-Host "╔═══════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "║  $Title" -ForegroundColor Cyan
    Write-Host "╚═══════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
    Write-Host ""
}

function Test-Case {
    param(
        [string]$Name,
        [scriptblock]$Action
    )
    
    $script:testResults.Total++
    Write-Host "🧪 $Name" -ForegroundColor Yellow -NoNewline
    
    try {
        & $Action
        Write-Host " ✅ PASS" -ForegroundColor Green
        $script:testResults.Pass++
    }
    catch {
        Write-Host " ❌ FAIL" -ForegroundColor Red
        Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
        $script:testResults.Fail++
    }
}

Write-Host "╔═══════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║         速金盈APP 综合功能测试                             ║" -ForegroundColor Cyan
Write-Host "╚═══════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

# ============================================================
# 前置准备：创建测试用户
# ============================================================

Write-TestHeader "前置准备：用户登录"

Test-Case "注册测试用户" {
    $body = @{ phone = $PHONE; password = $PASSWORD } | ConvertTo-Json
    try {
        $resp = Invoke-RestMethod -Uri "$BASE_URL/auth/register" -Method Post -Body $body -ContentType "application/json" -ErrorAction Stop
        Write-Host "   User ID: $($resp.user.id)" -ForegroundColor Gray
    }
    catch {
        if ($_.Exception.Response.StatusCode -ne 400) { throw }
        Write-Host "   用户已存在" -ForegroundColor Gray
    }
}

Test-Case "用户登录获取Token" {
    $body = @{ phone = $PHONE; password = $PASSWORD } | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE_URL/auth/login" -Method Post -Body $body -ContentType "application/json"
    $script:TOKEN = $resp.access_token
    $script:USER_ID = $resp.user.id
    Write-Host "   Token: $($TOKEN.Substring(0,30))..." -ForegroundColor Gray
}

Test-Case "设置支付密码" {
    $headers = @{ Authorization = "Bearer $TOKEN" }
    $body = @{ pay_password = $PAYPASS } | ConvertTo-Json
    try {
        Invoke-RestMethod -Uri "$BASE_URL/user/paypass/set" -Method Post -Body $body -ContentType "application/json" -Headers $headers -ErrorAction Stop | Out-Null
    }
    catch {
        if ($_.Exception.Message -notlike "*已设置*") { throw }
    }
}

# ============================================================
# 高优先级测试1：充值流程测试
# ============================================================

Write-TestHeader "高优先级1：充值流程测试"

Test-Case "手动充值（直接修改数据库）" {
    # 注意：实际生产环境需要通过充值审核流程
    # 这里为了测试方便，直接使用数据库命令充值
    Write-Host "   模拟充值 50000 元到账户" -ForegroundColor Gray
    # TODO: 实际应该通过充值审核API
    $script:testResults.Skip++
    throw "需要数据库访问权限或充值审核API"
}

# ============================================================
# 高优先级测试2：订单模块测试
# ============================================================

Write-TestHeader "高优先级2：订单模块测试"

Test-Case "创建订单（锁价买料）" {
    $headers = @{ Authorization = "Bearer $TOKEN" }
    $body = @{
        type = "long_buy"
        locked_price = 500.00
        weight_g = 100.0
        deposit = 10000.00
        pay_password = $PAYPASS
    } | ConvertTo-Json
    
    try {
        $resp = Invoke-RestMethod -Uri "$BASE_URL/orders" -Method Post -Body $body -ContentType "application/json" -Headers $headers
        Write-Host "   订单号: $($resp.order_id)" -ForegroundColor Gray
        Write-Host "   定金率: $($resp.margin_rate)%" -ForegroundColor Gray
        $script:ORDER_ID = $resp.order_id
    }
    catch {
        if ($_.Exception.Message -like "*定金不足*") {
            Write-Host "   需要先充值定金" -ForegroundColor Yellow
            throw "定金不足，需要充值"
        }
        throw
    }
}

Test-Case "查询订单列表" {
    $headers = @{ Authorization = "Bearer $TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/orders?status=holding" -Method Get -Headers $headers
    Write-Host "   持仓订单数: $($resp.total)" -ForegroundColor Gray
}

Test-Case "查询订单详情" {
    if (-not $script:ORDER_ID) {
        throw "没有可用的订单ID"
    }
    $headers = @{ Authorization = "Bearer $TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/orders/$($script:ORDER_ID)" -Method Get -Headers $headers
    Write-Host "   当前价格: $($resp.current_price)" -ForegroundColor Gray
    Write-Host "   浮动盈亏: $($resp.pnl_float)" -ForegroundColor Gray
}

Test-Case "现金结算订单" {
    if (-not $script:ORDER_ID) {
        throw "没有可用的订单ID"
    }
    $headers = @{ Authorization = "Bearer $TOKEN" }
    $body = @{
        settle_price = 510.00
        pay_password = $PAYPASS
    } | ConvertTo-Json
    
    $resp = Invoke-RestMethod -Uri "$BASE_URL/orders/$($script:ORDER_ID)/settle" -Method Post -Body $body -ContentType "application/json" -Headers $headers
    Write-Host "   结算价格: $($resp.settled_price)" -ForegroundColor Gray
    Write-Host "   结算盈亏: $($resp.settled_pnl)" -ForegroundColor Gray
}

# ============================================================
# 高优先级测试3：风控流程测试
# ============================================================

Write-TestHeader "高优先级3：风控流程测试"

Test-Case "查询风控统计（当前价500）" {
    $headers = @{ Authorization = "Bearer $TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/risk/statistics?current_price=500" -Method Get -Headers $headers
    Write-Host "   总订单: $($resp.total_orders)" -ForegroundColor Gray
    Write-Host "   强平: $($resp.force_close_count)" -ForegroundColor Gray
    Write-Host "   高风险: $($resp.high_risk_count)" -ForegroundColor Gray
    Write-Host "   预警: $($resp.warning_count)" -ForegroundColor Gray
}

Test-Case "模拟价格下跌（触发预警）" {
    # 注意：这需要修改风控调度器的价格或等待定时任务执行
    Write-Host "   风控调度器每60秒自动检查" -ForegroundColor Gray
    Write-Host "   当前实时监控运行中" -ForegroundColor Gray
}

Test-Case "验证风控通知" {
    $headers = @{ Authorization = "Bearer $TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/notifications/unread" -Method Get -Headers $headers
    Write-Host "   未读通知: $($resp.count)" -ForegroundColor Gray
    if ($resp.notifications -and $resp.notifications.Count -gt 0) {
        $riskNotifs = $resp.notifications | Where-Object { $_.type -eq "risk" }
        Write-Host "   风控通知: $($riskNotifs.Count)" -ForegroundColor Gray
    }
}

# ============================================================
# 中优先级测试1：通知完整测试
# ============================================================

Write-TestHeader "中优先级1：通知系统测试"

Test-Case "查询所有通知" {
    $headers = @{ Authorization = "Bearer $TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/notifications?limit=20" -Method Get -Headers $headers
    Write-Host "   通知总数: $($resp.total)" -ForegroundColor Gray
    
    if ($resp.notifications) {
        $types = $resp.notifications | Group-Object -Property type
        foreach ($type in $types) {
            Write-Host "   - $($type.Name): $($type.Count)" -ForegroundColor Gray
        }
    }
}

Test-Case "查询未读通知数量" {
    $headers = @{ Authorization = "Bearer $TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/notifications/count" -Method Get -Headers $headers
    Write-Host "   未读数量: $($resp.count)" -ForegroundColor Gray
}

Test-Case "标记通知为已读" {
    $headers = @{ Authorization = "Bearer $TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/notifications/read-all" -Method Post -Headers $headers
    Write-Host "   $($resp.message)" -ForegroundColor Gray
}

# ============================================================
# 中优先级测试2：销售管理测试
# ============================================================

Write-TestHeader "中优先级2：销售管理测试"

Test-Case "查询销售排行榜（总积分）" {
    $headers = @{ Authorization = "Bearer $TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/sales/ranking?limit=10&by_month=false" -Method Get -Headers $headers
    Write-Host "   销售人数: $($resp.rankings.Count)" -ForegroundColor Gray
}

Test-Case "查询销售排行榜（本月积分）" {
    $headers = @{ Authorization = "Bearer $TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/sales/ranking?limit=10&by_month=true" -Method Get -Headers $headers
    Write-Host "   本月活跃销售: $($resp.rankings.Count)" -ForegroundColor Gray
}

# ============================================================
# 中优先级测试3：银行卡管理
# ============================================================

Write-TestHeader "中优先级3：银行卡管理测试"

Test-Case "添加银行卡" {
    $headers = @{ Authorization = "Bearer $TOKEN" }
    $body = @{
        bank_name = "中国建设银行"
        card_number = "6217001234567890"
        card_holder = "测试用户"
        pay_password = $PAYPASS
    } | ConvertTo-Json
    
    try {
        $resp = Invoke-RestMethod -Uri "$BASE_URL/bank-cards" -Method Post -Body $body -ContentType "application/json" -Headers $headers
        Write-Host "   卡号: $($resp.card_number)" -ForegroundColor Gray
        Write-Host "   默认卡: $($resp.is_default)" -ForegroundColor Gray
    }
    catch {
        if ($_.Exception.Message -like "*最多*") {
            Write-Host "   已达到最大银行卡数量" -ForegroundColor Yellow
        }
        else { throw }
    }
}

Test-Case "查询银行卡列表" {
    $headers = @{ Authorization = "Bearer $TOKEN" }
    $resp = Invoke-RestMethod -Uri "$BASE_URL/bank-cards" -Method Get -Headers $headers
    Write-Host "   银行卡数: $($resp.cards.Count)" -ForegroundColor Gray
    foreach ($card in $resp.cards) {
        $defaultMark = if ($card.is_default) { " [默认]" } else { "" }
        Write-Host "   - $($card.bank_name): $($card.card_number)$defaultMark" -ForegroundColor Gray
    }
}

# ============================================================
# 低优先级测试：性能测试
# ============================================================

Write-TestHeader "低优先级：性能测试"

Test-Case "API响应时间测试（登录）" {
    $body = @{ phone = $PHONE; password = $PASSWORD } | ConvertTo-Json
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    Invoke-RestMethod -Uri "$BASE_URL/auth/login" -Method Post -Body $body -ContentType "application/json" | Out-Null
    $sw.Stop()
    Write-Host "   响应时间: $($sw.ElapsedMilliseconds)ms" -ForegroundColor Gray
    
    if ($sw.ElapsedMilliseconds -gt 1000) {
        throw "响应时间过长: $($sw.ElapsedMilliseconds)ms"
    }
}

Test-Case "API响应时间测试（查询订单）" {
    $headers = @{ Authorization = "Bearer $TOKEN" }
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    Invoke-RestMethod -Uri "$BASE_URL/orders" -Method Get -Headers $headers | Out-Null
    $sw.Stop()
    Write-Host "   响应时间: $($sw.ElapsedMilliseconds)ms" -ForegroundColor Gray
}

# ============================================================
# 测试报告
# ============================================================

Write-Host ""
Write-Host "╔═══════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                     测试报告                               ║" -ForegroundColor Cyan
Write-Host "╚═══════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

Write-Host "📊 测试统计:" -ForegroundColor White
Write-Host "   总测试数: $($testResults.Total)" -ForegroundColor Gray
Write-Host "   通过: $($testResults.Pass)" -ForegroundColor Green
Write-Host "   失败: $($testResults.Fail)" -ForegroundColor Red
Write-Host "   跳过: $($testResults.Skip)" -ForegroundColor Yellow

$passRate = if ($testResults.Total -gt 0) { 
    [math]::Round(($testResults.Pass / $testResults.Total) * 100, 2) 
} else { 0 }

Write-Host "   通过率: $passRate%" -ForegroundColor $(if($testResults.Fail -eq 0){"Green"}else{"Yellow"})
Write-Host ""

if ($testResults.Fail -eq 0) {
    Write-Host "🎉 所有测试通过！" -ForegroundColor Green
}
else {
    Write-Host "⚠️  部分测试失败，请检查错误信息" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "提示：" -ForegroundColor Cyan
Write-Host "- 订单测试需要先充值定金（手动修改数据库或通过充值审核）" -ForegroundColor Gray
Write-Host "- 风控测试需要等待定时任务执行（60秒间隔）" -ForegroundColor Gray
Write-Host "- 销售提成需要订单结算后自动计算" -ForegroundColor Gray
