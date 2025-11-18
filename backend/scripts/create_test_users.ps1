# 创建测试用户脚本
# 用于快速创建测试账号

$baseUrl = "http://localhost:8080/api/v1"

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "速金盈系统 - 创建测试用户" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# 用户列表
$users = @(
    @{
        username = "admin"
        password = "123456"
        real_name = "系统管理员"
        phone = "13800000001"
        role = "super_admin"
    },
    @{
        username = "customer"
        password = "123456"
        real_name = "测试客户"
        phone = "13800000002"
        role = "customer"
    },
    @{
        username = "sales"
        password = "123456"
        real_name = "测试销售"
        phone = "13800000003"
        role = "sales"
    },
    @{
        username = "support"
        password = "123456"
        real_name = "测试客服"
        phone = "13800000004"
        role = "support"
    }
)

# 创建邀请码（作为销售角色）
function Create-InviteCode {
    param($token)
    
    try {
        $body = @{
            "code" = "TEST2025"
            "max_uses" = 100
            "expires_at" = "2026-12-31T23:59:59Z"
        } | ConvertTo-Json
        
        $response = Invoke-WebRequest -Uri "$baseUrl/invite-codes" `
            -Method POST `
            -Headers @{
                "Content-Type" = "application/json"
                "Authorization" = "Bearer $token"
            } `
            -Body $body `
            -ErrorAction Stop
            
        Write-Host "✅ 创建邀请码成功: TEST2025" -ForegroundColor Green
        return $true
    } catch {
        Write-Host "⚠️  邀请码可能已存在" -ForegroundColor Yellow
        return $false
    }
}

# 注册用户
function Register-User {
    param($user)
    
    Write-Host "正在注册用户: $($user.username)..." -ForegroundColor Yellow
    
    try {
        $body = @{
            "username" = $user.username
            "password" = $user.password
            "real_name" = $user.real_name
            "phone" = $user.phone
            "invite_code" = "TEST2025"
        } | ConvertTo-Json
        
        $response = Invoke-WebRequest -Uri "$baseUrl/auth/register" `
            -Method POST `
            -Headers @{"Content-Type" = "application/json"} `
            -Body $body `
            -ErrorAction Stop
            
        Write-Host "✅ 用户 $($user.username) 注册成功" -ForegroundColor Green
        return $true
    } catch {
        $errorMessage = $_.Exception.Message
        if ($errorMessage -like "*用户名已存在*") {
            Write-Host "⚠️  用户 $($user.username) 已存在，跳过" -ForegroundColor Yellow
        } else {
            Write-Host "❌ 用户 $($user.username) 注册失败: $errorMessage" -ForegroundColor Red
        }
        return $false
    }
}

# 登录获取token
function Login-User {
    param($username, $password)
    
    try {
        $body = @{
            "username" = $username
            "password" = $password
        } | ConvertTo-Json
        
        $response = Invoke-RestMethod -Uri "$baseUrl/auth/login" `
            -Method POST `
            -Headers @{"Content-Type" = "application/json"} `
            -Body $body `
            -ErrorAction Stop
            
        return $response.data.access_token
    } catch {
        Write-Host "❌ 登录失败: $username" -ForegroundColor Red
        return $null
    }
}

# 审核用户（管理员操作）
function Approve-User {
    param($token, $userId)
    
    try {
        $body = @{
            "status" = "active"
        } | ConvertTo-Json
        
        $response = Invoke-WebRequest -Uri "$baseUrl/admin/users/$userId/status" `
            -Method PUT `
            -Headers @{
                "Content-Type" = "application/json"
                "Authorization" = "Bearer $token"
            } `
            -Body $body `
            -ErrorAction Stop
            
        return $true
    } catch {
        return $false
    }
}

# 主流程
Write-Host "步骤1: 创建管理员账号" -ForegroundColor Cyan
Write-Host "------------------------------"
Register-User -user $users[0]
Write-Host ""

Write-Host "步骤2: 登录管理员账号" -ForegroundColor Cyan
Write-Host "------------------------------"
$adminToken = Login-User -username "admin" -password "123456"
if ($adminToken) {
    Write-Host "✅ 管理员登录成功" -ForegroundColor Green
} else {
    Write-Host "❌ 管理员登录失败，请检查账号状态" -ForegroundColor Red
    Write-Host "提示: 可能需要手动在数据库中将admin用户状态设置为active" -ForegroundColor Yellow
    exit 1
}
Write-Host ""

Write-Host "步骤3: 创建邀请码" -ForegroundColor Cyan
Write-Host "------------------------------"
Create-InviteCode -token $adminToken
Write-Host ""

Write-Host "步骤4: 创建其他测试用户" -ForegroundColor Cyan
Write-Host "------------------------------"
for ($i = 1; $i -lt $users.Count; $i++) {
    Register-User -user $users[$i]
}
Write-Host ""

Write-Host "步骤5: 审核所有用户" -ForegroundColor Cyan
Write-Host "------------------------------"
# 这里需要获取所有待审核用户ID并审核
Write-Host "⚠️  请手动在管理后台审核用户，或直接在数据库中设置用户状态为active" -ForegroundColor Yellow
Write-Host ""

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "测试用户创建完成！" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "测试账号列表:" -ForegroundColor Cyan
Write-Host "------------------------------"
foreach ($user in $users) {
    Write-Host "用户名: $($user.username)" -ForegroundColor White
    Write-Host "密码: $($user.password)" -ForegroundColor White
    Write-Host "角色: $($user.role)" -ForegroundColor White
    Write-Host "------------------------------"
}
Write-Host ""
Write-Host "前端访问地址: http://localhost:5173" -ForegroundColor Green
Write-Host "后端API地址: http://localhost:8080" -ForegroundColor Green
Write-Host ""
Write-Host "开始测试吧！🎉" -ForegroundColor Cyan
