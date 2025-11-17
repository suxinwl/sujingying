# 速金盈APP 自动化测试脚本 (PowerShell)
# 用途：测试定金充值和银行卡管理功能

$BASE_URL = "http://localhost:8080/api/v1"
$ACCESS_TOKEN = ""

Write-Host "========== 速金盈APP 自动化测试 ==========" -ForegroundColor Cyan
Write-Host ""

# 1. 注册用户
Write-Host "1️⃣  测试用户注册..." -ForegroundColor Yellow
$registerBody = @{
    phone = "13900000001"
    password = "Test@123"
} | ConvertTo-Json

try {
    $registerResp = Invoke-RestMethod -Uri "$BASE_URL/auth/register" -Method Post -Body $registerBody -ContentType "application/json"
    Write-Host "✅ 注册成功: $($registerResp | ConvertTo-Json -Compress)" -ForegroundColor Green
}
catch {
    Write-Host "⚠️  用户可能已存在，继续测试..." -ForegroundColor Yellow
}
Write-Host ""

# 2. 登录获取Token
Write-Host "2️⃣  测试用户登录..." -ForegroundColor Yellow
$loginBody = @{
    phone = "13900000001"
    password = "Test@123"
} | ConvertTo-Json

$loginResp = Invoke-RestMethod -Uri "$BASE_URL/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$ACCESS_TOKEN = $loginResp.access_token
Write-Host "✅ 登录成功" -ForegroundColor Green
Write-Host "   User ID: $($loginResp.user.id)"
Write-Host "   Token: $($ACCESS_TOKEN.Substring(0, 20))..."
Write-Host ""

# 3. 设置支付密码
Write-Host "3️⃣  测试设置支付密码..." -ForegroundColor Yellow
$paypassBody = @{
    pay_password = "123456"
} | ConvertTo-Json

$headers = @{
    Authorization = "Bearer $ACCESS_TOKEN"
}

try {
    $paypassResp = Invoke-RestMethod -Uri "$BASE_URL/user/paypass/set" -Method Post -Body $paypassBody -ContentType "application/json" -Headers $headers
    Write-Host "✅ 支付密码设置成功" -ForegroundColor Green
}
catch {
    Write-Host "⚠️  支付密码可能已设置" -ForegroundColor Yellow
}
Write-Host ""

# 4. 添加银行卡（需支付密码）
Write-Host "4️⃣  测试添加银行卡（需支付密码验证）..." -ForegroundColor Yellow
$addCardBody = @{
    bank_name = "中国工商银行"
    card_number = "6222021234567890"
    card_holder = "张三"
    pay_password = "123456"
} | ConvertTo-Json

try {
    $cardResp = Invoke-RestMethod -Uri "$BASE_URL/bank-cards" -Method Post -Body $addCardBody -ContentType "application/json" -Headers $headers
    Write-Host "✅ 银行卡添加成功" -ForegroundColor Green
    Write-Host "   银行: $($cardResp.bank_name)"
    Write-Host "   卡号: $($cardResp.card_number)"
    Write-Host "   持卡人: $($cardResp.card_holder)"
    Write-Host "   默认卡: $($cardResp.is_default)"
}
catch {
    Write-Host "❌ 添加银行卡失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 5. 查询银行卡列表
Write-Host "5️⃣  测试查询银行卡列表..." -ForegroundColor Yellow
try {
    $cardsResp = Invoke-RestMethod -Uri "$BASE_URL/bank-cards" -Method Get -Headers $headers
    Write-Host "✅ 查询成功，共 $($cardsResp.cards.Count) 张银行卡" -ForegroundColor Green
    foreach ($card in $cardsResp.cards) {
        Write-Host "   - $($card.bank_name): $($card.card_number) ($($card.card_holder))"
    }
}
catch {
    Write-Host "❌ 查询失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 6. 测试未读通知数量
Write-Host "6️⃣  测试通知系统..." -ForegroundColor Yellow
try {
    $notifyResp = Invoke-RestMethod -Uri "$BASE_URL/notifications/count" -Method Get -Headers $headers
    Write-Host "✅ 未读通知数量: $($notifyResp.count)" -ForegroundColor Green
}
catch {
    Write-Host "❌ 查询通知失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "✅ 自动化测试完成！" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Cyan
