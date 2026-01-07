# Proof of Work (PoW) 保護機制

這個專案使用 **Proof of Work (工作量證明)** 來防止註冊 API 被濫用或 DDoS 攻擊。

## 🔐 工作原理

### 1. 挑戰-回應機制

註冊流程分為三個步驟：

```
客戶端                                服務端
  │                                    │
  │  1. 請求挑戰 (GET /api/challenge) │
  │ ──────────────────────────────────>│
  │                                    │ 生成隨機挑戰
  │ <───────────────────────────────── │
  │  { challenge, difficulty }         │
  │                                    │
  │  2. 計算 PoW (客戶端)              │
  │     尋找 nonce 使得：              │
  │     SHA256(challenge + nonce)      │
  │     開頭有 N 個零                  │
  │                                    │
  │  3. 提交註冊 + PoW 證明            │
  │ ──────────────────────────────────>│
  │                                    │ 驗證 PoW
  │                                    │ ✓ 通過 → 註冊
  │                                    │ ✗ 失敗 → 拒絕
  │ <───────────────────────────────── │
  │  註冊結果                           │
```

### 2. PoW 演算法

**目標**: 找到一個 `nonce`，使得 `SHA256(challenge + nonce)` 的結果以 N 個零開頭

**範例** (難度 = 4):
```javascript
challenge = "abc123..."
nonce = 0, 1, 2, ... 直到找到符合條件的 nonce

SHA256("abc123..." + "0") = "a7f3b2..." ✗
SHA256("abc123..." + "1") = "9c2d8e..." ✗
SHA256("abc123..." + "2") = "f1e4a9..." ✗
...
SHA256("abc123..." + "12847") = "0000a3..." ✓ (符合！)
```

### 3. 難度設定

在 [backend/main.go:35](backend/main.go#L35)：

```go
powDifficulty = 4  // 需要 4 個前導零
challengeTTL = 5 * time.Minute  // 挑戰有效期 5 分鐘
```

**難度對應的計算複雜度**：
- 難度 1: 平均 ~16 次嘗試
- 難度 2: 平均 ~256 次嘗試
- 難度 3: 平均 ~4,096 次嘗試
- 難度 4: 平均 ~65,536 次嘗試 ⭐ (目前設定)
- 難度 5: 平均 ~1,048,576 次嘗試
- 難度 6: 平均 ~16,777,216 次嘗試

## 🛡️ 防護效果

### 防止攻擊類型

1. **暴力註冊攻擊**
   - 攻擊者需要為每次註冊請求計算 PoW
   - 難度 4 平均需要 ~65,536 次 SHA256 運算
   - 大幅增加批量註冊的成本

2. **DDoS 攻擊**
   - 無效的請求會被 PoW 驗證拒絕
   - 攻擊者無法輕易偽造有效的 PoW

3. **自動化腳本**
   - 簡單的 HTTP 請求無法通過 PoW 驗證
   - 需要實際計算才能註冊

### 合法用戶影響

- **現代瀏覽器**: 通常在 0.5-3 秒內完成 (難度 4)
- **用戶體驗**: 顯示進度條和嘗試次數
- **透明處理**: 自動在背景計算，無需用戶干預

## 📊 API 端點

### GET /api/challenge

獲取新的 PoW 挑戰

**回應**:
```json
{
  "success": true,
  "message": "Challenge generated",
  "data": {
    "challenge": "eb809f73d9f5f32ca2f050099ee971cb...",
    "difficulty": 4
  }
}
```

### POST /api/register

提交註冊請求 (需要 PoW 證明)

**請求**:
```json
{
  "username": "john_doe",
  "password": "password123",
  "email": "john@example.com",
  "challenge": "eb809f73d9f5f32ca2f050099ee971cb...",
  "nonce": "12847"
}
```

**驗證失敗回應**:
```json
{
  "success": false,
  "message": "Invalid proof of work"
}
```

## 🧪 測試 PoW

### 手動測試

```bash
# 1. 獲取挑戰
CHALLENGE=$(curl -s http://localhost:8080/api/challenge | jq -r '.data.challenge')
echo "Challenge: $CHALLENGE"

# 2. 手動計算 PoW (使用 Python 範例)
python3 << EOF
import hashlib

challenge = "$CHALLENGE"
nonce = 0
difficulty = 4
target = "0" * difficulty

while True:
    data = challenge + str(nonce)
    hash_result = hashlib.sha256(data.encode()).hexdigest()
    if hash_result.startswith(target):
        print(f"Found nonce: {nonce}")
        print(f"Hash: {hash_result}")
        break
    nonce += 1
EOF
```

### 壓力測試

嘗試在沒有 PoW 的情況下註冊（應該失敗）：

```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test",
    "password": "password",
    "email": "test@example.com",
    "challenge": "invalid",
    "nonce": "0"
  }'

# 預期回應: {"success":false,"message":"Invalid or expired challenge"}
```

## ⚙️ 調整難度

根據你的需求調整難度：

### 降低難度 (更快，但保護較弱)

在 [backend/main.go:35](backend/main.go#L35) 修改：
```go
powDifficulty = 3  // 平均 ~4,096 次嘗試
```

### 提高難度 (更強保護，但較慢)

```go
powDifficulty = 5  // 平均 ~1,048,576 次嘗試
```

## 📈 效能考量

### 客戶端

- **CPU 使用**: PoW 計算會短暫使用 100% CPU
- **優化**: 每 1000 次嘗試會 yield 給瀏覽器，防止凍結
- **Web Workers**: 可考慮移到 Worker 執行緒（進階）

### 服務端

- **驗證成本**: O(1) - 只需一次 SHA256 運算
- **記憶體**: Challenge 儲存在記憶體中，定期清理過期項目
- **並發**: 使用 RWMutex 保護 challenge map

## 🔍 監控

後端會記錄 PoW 事件：

```
2026/01/07 10:49:28 Generated challenge: eb809f73d9f5f32c... (expires in 5m0s)
2026/01/07 10:49:31 PoW verified for challenge: eb809f73d9f5f32c...
```

監控這些日誌可以：
- 追蹤 PoW 使用情況
- 識別可疑的活動模式
- 調整難度設定

## 🚀 生產環境建議

1. **動態難度調整**: 根據負載自動調整難度
2. **Rate Limiting**: PoW 之外再加上 IP 限流
3. **分散式儲存**: 使用 Redis 儲存 challenges（支援多實例）
4. **監控告警**: 設置異常 PoW 請求的告警
5. **降級機制**: 高負載時可暫時提高難度

## 📚 相關資源

- [Hashcash - Proof of Work](https://en.wikipedia.org/wiki/Hashcash)
- [Bitcoin PoW](https://en.bitcoin.it/wiki/Proof_of_work)
- [Web Crypto API](https://developer.mozilla.org/en-US/docs/Web/API/Web_Crypto_API)

## 🎯 總結

**優點**:
- ✅ 有效防止自動化攻擊
- ✅ 對合法用戶影響小
- ✅ 無需 CAPTCHA
- ✅ 可調整難度

**缺點**:
- ❌ 增加客戶端運算負擔
- ❌ 低端設備可能較慢
- ❌ 無法完全阻止分散式攻擊

**最佳實踐**: PoW + Rate Limiting + IP 黑名單 = 多層防護
