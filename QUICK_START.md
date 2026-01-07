# 快速開始指南

## 🚀 5 分鐘快速啟動

### 1. 啟動資料庫
```bash
docker start pow-postgres
```

如果容器不存在，建立新的：
```bash
docker run -d \
  --name pow-postgres \
  -e POSTGRES_USER=powuser \
  -e POSTGRES_PASSWORD=powpass \
  -e POSTGRES_DB=powdb \
  -p 5432:5432 \
  postgres:16-alpine
```

### 2. 啟動後端
```bash
cd backend
go run main.go
```

看到這個訊息表示成功：
```
Successfully connected to database
Server starting on :8080
PoW difficulty: 4 leading zeros
```

### 3. 啟動前端
```bash
cd frontend
npm run dev
```

### 4. 開始使用
打開瀏覽器：http://localhost:3000

## ⚡ 測試 PoW 功能

### 使用測試腳本
```bash
./test_pow.sh
```

### 手動測試
```bash
# 取得挑戰
curl http://localhost:8080/api/challenge

# 然後在瀏覽器中註冊
# 觀察 PoW 計算過程
```

## 🛑 停止所有服務

```bash
# 停止資料庫
docker stop pow-postgres

# 停止後端和前端（Ctrl+C）
```

## 📚 更多資訊

- [完整文檔](README.md)
- [PoW 保護機制](POW_PROTECTION.md)
- [實作總結](IMPLEMENTATION_SUMMARY.md)

## 🔧 調整 PoW 難度

編輯 `backend/main.go` 第 35 行：

```go
powDifficulty = 4  // 改成 3 會更快，5 會更慢
```

## 🐛 常見問題

### 後端啟動失敗
- 檢查 PostgreSQL 是否運行：`docker ps | grep pow-postgres`
- 檢查 port 8080 是否被佔用：`lsof -i:8080`

### 前端無法連接後端
- 確認後端在 port 8080 運行
- 檢查 CORS 設定

### PoW 計算太慢
- 降低難度（改成 3 或 2）
- 檢查 CPU 使用率

### Challenge 過期
- 預設 TTL 是 5 分鐘
- 如需更長時間，修改 `challengeTTL`
