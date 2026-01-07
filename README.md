# POW API - 完整的註冊系統

這是一個使用 Golang 後端 + Nuxt 3 前端的完整用戶註冊系統，具備 **Proof of Work (工作量證明)** 保護機制防止 API 濫用。

## 專案結構

```
pow-api/
├── backend/          # Golang API 後端
│   ├── main.go
│   ├── go.mod
│   └── go.sum
├── frontend/         # Nuxt 3 前端
│   ├── pages/
│   │   └── index.vue
│   ├── app.vue
│   ├── nuxt.config.ts
│   └── package.json
└── README.md
```

## 功能特性

### 後端 (Golang)
- ✅ PostgreSQL 資料庫整合
- ✅ RESTful API 設計
- ✅ bcrypt 密碼加密
- ✅ 輸入驗證
- ✅ CORS 支援
- ✅ 用戶重複檢查
- 🔐 **Proof of Work (PoW) 保護** - 防止 API 濫用和 DDoS 攻擊
- ⏰ Challenge 過期機制 (5 分鐘 TTL)
- 🔄 自動清理過期 challenges

### 前端 (Nuxt 3)
- ✅ 美觀的註冊表單 UI
- ✅ 即時表單驗證
- ✅ 錯誤處理和顯示
- ✅ 響應式設計 (RWD)
- ✅ 成功/失敗訊息提示
- ✅ 載入狀態顯示
- 🔐 **自動 PoW 計算** - SHA-256 工作量證明
- 📊 即時顯示 PoW 計算進度
- 🎨 PoW 保護徽章顯示

## 環境需求

- Go 1.21+
- Node.js 18+
- PostgreSQL (透過 Docker/OrbStack)

## 安裝和執行

### 1. 啟動 PostgreSQL 資料庫

```bash
docker run -d \
  --name pow-postgres \
  -e POSTGRES_USER=powuser \
  -e POSTGRES_PASSWORD=powpass \
  -e POSTGRES_DB=powdb \
  -p 5432:5432 \
  postgres:16-alpine
```

建立資料表：
```bash
docker exec pow-postgres psql -U powuser -d powdb -c "
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);"
```

### 2. 啟動後端 API

```bash
cd backend
go mod download
go run main.go
```

後端會在 `http://localhost:8080` 啟動

### 3. 啟動前端

```bash
cd frontend
npm install
npm run dev
```

前端會在 `http://localhost:3000` 啟動

## 使用說明

1. 打開瀏覽器訪問 `http://localhost:3000`
2. 填寫註冊表單：
   - **使用者名稱**：至少 3 個字元
   - **電子郵件**：有效的 email 格式
   - **密碼**：至少 6 個字元
3. 點擊「註冊」按鈕
4. **自動執行 PoW**: 系統會自動：
   - 向後端請求挑戰 (challenge)
   - 在瀏覽器中計算 Proof of Work
   - 顯示計算進度（已嘗試次數）
   - 完成後自動提交註冊
5. 成功後會顯示成功訊息（包含 PoW 嘗試次數）

> 💡 **PoW 計算時間**: 難度 4 通常需要 0.5-3 秒（平均 ~65,536 次嘗試）

## 🔐 Proof of Work 保護

本系統使用 PoW 機制防止以下攻擊：
- ✅ 暴力註冊攻擊
- ✅ DDoS 攻擊
- ✅ 自動化腳本濫用

**工作原理**: 客戶端必須找到一個 nonce，使得 `SHA256(challenge + nonce)` 產生指定數量的前導零。

詳細說明請參考：[POW_PROTECTION.md](POW_PROTECTION.md)

## API 端點

### 取得 PoW 挑戰
```
GET http://localhost:8080/api/challenge
```

**回應:**
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

### 註冊 API (需要 PoW)
```
POST http://localhost:8080/api/register
Content-Type: application/json

{
  "username": "john_doe",
  "password": "password123",
  "email": "john@example.com",
  "challenge": "eb809f73d9f5f32ca2f050099ee971cb...",
  "nonce": "12847"
}
```

> ⚠️ **注意**: `challenge` 和 `nonce` 為必填欄位，必須先從 `/api/challenge` 取得 challenge，並計算出有效的 nonce。

**成功回應 (201):**
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "id": 1,
    "username": "john_doe",
    "email": "john@example.com",
    "created_at": "2026-01-06T10:20:33.955381Z"
  }
}
```

**錯誤回應 (400/409):**
```json
{
  "success": false,
  "message": "Username already exists"
}
```

### 健康檢查
```
GET http://localhost:8080/health
```

## 查看資料庫資料

```bash
# 查看所有用戶
docker exec pow-postgres psql -U powuser -d powdb -c "SELECT id, username, email, created_at FROM users;"

# 查看用戶數量
docker exec pow-postgres psql -U powuser -d powdb -c "SELECT COUNT(*) FROM users;"
```

## 停止服務

```bash
# 停止 PostgreSQL
docker stop pow-postgres

# 可選：移除容器
docker rm pow-postgres
```

## 技術棧

**後端:**
- Go 1.21
- gorilla/mux (路由)
- lib/pq (PostgreSQL 驅動)
- bcrypt (密碼加密)

**前端:**
- Nuxt 3
- Vue 3 Composition API
- 原生 CSS (漸層背景、動畫效果)

**資料庫:**
- PostgreSQL 16

## 安全性特性

- ✅ 密碼使用 bcrypt 加密儲存
- ✅ SQL 使用參數化查詢防止 SQL Injection
- ✅ CORS 配置
- ✅ 前後端雙重驗證
- ✅ 用戶名稱唯一性檢查

## 後續改進建議

1. 加入登入功能
2. 加入 JWT 認證
3. Email 驗證功能
4. 密碼重設功能
5. 用戶個人資料頁面
6. Rate limiting
7. 單元測試和整合測試
8. Docker Compose 一鍵部署
9. 環境變數配置管理
10. 日誌系統

## 授權

MIT License
