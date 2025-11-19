#!/bin/bash

# 测试支付密码API
# 使用方法：bash test_paypass_api.sh YOUR_JWT_TOKEN

if [ -z "$1" ]; then
  echo "请提供JWT Token"
  echo "用法: bash test_paypass_api.sh YOUR_JWT_TOKEN"
  exit 1
fi

TOKEN=$1

echo "================================"
echo "测试支付密码API"
echo "================================"
echo ""

echo "1. 测试首次设置支付密码..."
curl -X POST http://localhost:8080/api/v1/user/paypass \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "new_pay_password": "123456"
  }' \
  -w "\nHTTP状态码: %{http_code}\n" \
  -s

echo ""
echo "================================"
echo ""

echo "2. 测试修改支付密码..."
curl -X POST http://localhost:8080/api/v1/user/paypass \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "old_pay_password": "123456",
    "new_pay_password": "654321"
  }' \
  -w "\nHTTP状态码: %{http_code}\n" \
  -s

echo ""
echo "================================"
