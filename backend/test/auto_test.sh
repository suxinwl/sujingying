#!/bin/bash
# 自动化测试脚本
# 用途：测试定金充值和银行卡管理功能

BASE_URL="http://localhost:8080/api/v1"
ACCESS_TOKEN=""
USER_ID=""

echo "========== 速金盈APP 自动化测试 =========="
echo ""

# 1. 注册用户
echo "1️⃣ 测试用户注册..."
REGISTER_RESP=$(curl -s -X POST ${BASE_URL}/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13900000001",
    "password": "Test@123"
  }')
echo "注册响应: $REGISTER_RESP"
echo ""

# 2. 登录获取Token
echo "2️⃣ 测试用户登录..."
LOGIN_RESP=$(curl -s -X POST ${BASE_URL}/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13900000001",
    "password": "Test@123"
  }')
echo "登录响应: $LOGIN_RESP"
ACCESS_TOKEN=$(echo $LOGIN_RESP | grep -o '"access_token":"[^"]*' | sed 's/"access_token":"//')
USER_ID=$(echo $LOGIN_RESP | grep -o '"id":[0-9]*' | sed 's/"id"://')
echo "Access Token: $ACCESS_TOKEN"
echo "User ID: $USER_ID"
echo ""

# 3. 设置支付密码
echo "3️⃣ 测试设置支付密码..."
PAYPASS_RESP=$(curl -s -X POST ${BASE_URL}/user/paypass/set \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"pay_password": "123456"}')
echo "支付密码设置响应: $PAYPASS_RESP"
echo ""

# 4. 添加银行卡（需支付密码）
echo "4️⃣ 测试添加银行卡..."
ADDCARD_RESP=$(curl -s -X POST ${BASE_URL}/bank-cards \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "bank_name": "中国工商银行",
    "card_number": "6222021234567890",
    "card_holder": "张三",
    "pay_password": "123456"
  }')
echo "添加银行卡响应: $ADDCARD_RESP"
echo ""

# 5. 查询银行卡列表
echo "5️⃣ 测试查询银行卡列表..."
CARDS_RESP=$(curl -s -X GET ${BASE_URL}/bank-cards \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo "银行卡列表: $CARDS_RESP"
echo ""

# 6. 测试完成
echo "=========================================="
echo "✅ 自动化测试完成！"
echo "=========================================="
