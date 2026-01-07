#!/bin/bash

# POW API 測試腳本
# 展示完整的 PoW 註冊流程

API_BASE="http://localhost:8080"

echo "=================================="
echo "   POW API 註冊測試"
echo "=================================="
echo ""

# Step 1: 獲取挑戰
echo "步驟 1: 獲取 PoW 挑戰..."
CHALLENGE_RESPONSE=$(curl -s "$API_BASE/api/challenge")
echo "回應: $CHALLENGE_RESPONSE"
echo ""

# 解析 challenge 和 difficulty
CHALLENGE=$(echo $CHALLENGE_RESPONSE | jq -r '.data.challenge')
DIFFICULTY=$(echo $CHALLENGE_RESPONSE | jq -r '.data.difficulty')

echo "Challenge: $CHALLENGE"
echo "Difficulty: $DIFFICULTY (需要 $DIFFICULTY 個前導零)"
echo ""

# Step 2: 計算 PoW
echo "步驟 2: 計算 Proof of Work..."
echo "使用 Python 計算..."

# 使用 Python 計算 nonce
RESULT=$(python3 << EOF
import hashlib
import time

challenge = "$CHALLENGE"
difficulty = $DIFFICULTY
target = "0" * difficulty
nonce = 0
start_time = time.time()

while True:
    data = challenge + str(nonce)
    hash_result = hashlib.sha256(data.encode()).hexdigest()

    if hash_result.startswith(target):
        elapsed = time.time() - start_time
        print(f"{nonce}|{hash_result}|{elapsed:.2f}")
        break

    nonce += 1

    # 安全限制：最多嘗試 10,000,000 次
    if nonce > 10000000:
        print("error|timeout|0")
        break
EOF
)

NONCE=$(echo $RESULT | cut -d'|' -f1)
HASH=$(echo $RESULT | cut -d'|' -f2)
TIME=$(echo $RESULT | cut -d'|' -f3)

if [ "$NONCE" = "error" ]; then
    echo "❌ PoW 計算失敗（超時）"
    exit 1
fi

echo "✓ 找到有效的 nonce: $NONCE"
echo "  Hash: $HASH"
echo "  嘗試次數: $NONCE"
echo "  耗時: ${TIME}秒"
echo ""

# Step 3: 提交註冊
echo "步驟 3: 提交註冊請求..."

TIMESTAMP=$(date +%s)
USERNAME="test_user_$TIMESTAMP"
EMAIL="test_$TIMESTAMP@example.com"

REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE/api/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"$USERNAME\",
    \"password\": \"password123\",
    \"email\": \"$EMAIL\",
    \"challenge\": \"$CHALLENGE\",
    \"nonce\": \"$NONCE\"
  }")

echo "回應: $REGISTER_RESPONSE"
echo ""

# 檢查結果
SUCCESS=$(echo $REGISTER_RESPONSE | jq -r '.success')

if [ "$SUCCESS" = "true" ]; then
    echo "=================================="
    echo "✅ 註冊成功！"
    echo "=================================="
    echo ""
    echo "用戶資訊:"
    echo $REGISTER_RESPONSE | jq '.data'
    echo ""
    echo "統計數據:"
    echo "- PoW 嘗試次數: $NONCE"
    echo "- 計算耗時: ${TIME}秒"
    echo "- 平均速度: $(echo "scale=0; $NONCE / $TIME" | bc) 次/秒"
else
    echo "=================================="
    echo "❌ 註冊失敗"
    echo "=================================="
    MESSAGE=$(echo $REGISTER_RESPONSE | jq -r '.message')
    echo "錯誤訊息: $MESSAGE"
fi

echo ""
echo "測試完成！"
