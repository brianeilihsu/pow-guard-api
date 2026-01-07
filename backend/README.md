# POW API - User Registration API

這是一個使用 Golang 開發的用戶註冊 API。

## 功能特性

- 用戶註冊（username、password、email）
- 密碼使用 bcrypt 加密儲存
- 輸入驗證
- RESTful API 設計

## 環境需求

- Go 1.21 或更高版本

## 安裝步驟

1. 安裝依賴套件：
```bash
go mod download
```

2. 執行服務：
```bash
go run main.go
```

服務將在 `http://localhost:8080` 啟動

## API 端點

### 1. 健康檢查
```
GET /health
```

**回應範例：**
```json
{
  "success": true,
  "message": "Server is running"
}
```

### 2. 用戶註冊
```
POST /api/register
```

**請求 Body：**
```json
{
  "username": "john_doe",
  "password": "password123",
  "email": "john@example.com"
}
```

**驗證規則：**
- `username`: 必填，至少 3 個字元
- `password`: 必填，至少 6 個字元
- `email`: 必填，必須是有效的 email 格式

**成功回應 (201 Created)：**
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "username": "john_doe",
    "email": "john@example.com"
  }
}
```

**錯誤回應範例：**

- 無效的請求格式 (400 Bad Request)：
```json
{
  "success": false,
  "message": "Invalid request body"
}
```

- 驗證失敗 (400 Bad Request)：
```json
{
  "success": false,
  "message": "Password must be at least 6 characters"
}
```

- 用戶已存在 (409 Conflict)：
```json
{
  "success": false,
  "message": "Username already exists"
}
```

## 使用 cURL 測試

```bash
# 健康檢查
curl http://localhost:8080/health

# 註冊新用戶
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "password123",
    "email": "john@example.com"
  }'
```

## 專案結構

```
pow-api/
├── main.go          # 主程式，包含所有 API 處理邏輯
├── go.mod           # Go 模組依賴管理
└── README.md        # 專案說明文件
```

## 安全性特性

- 密碼使用 bcrypt 進行雜湊加密，不會以明文儲存
- 輸入驗證防止無效數據
- Email 格式驗證

## 注意事項

目前使用記憶體儲存用戶資料（map），重啟服務後資料會消失。在生產環境中，應該使用真實的資料庫（如 PostgreSQL、MySQL、MongoDB 等）來持久化儲存用戶資料。

## 後續改進建議

1. 整合資料庫（PostgreSQL、MySQL 或 MongoDB）
2. 加入 JWT 認證
3. 加入登入 API
4. 加入更多的安全性措施（如 rate limiting）
5. 加入單元測試和整合測試
6. 使用環境變數管理配置
