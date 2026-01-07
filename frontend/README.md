# Frontend - Nuxt 3 註冊 UI

這是一個使用 Nuxt 3 建立的用戶註冊前端介面。

## 功能特色

- 🎨 美觀的漸層背景設計
- ✨ 流暢的動畫效果
- 📱 完全響應式設計 (支援手機、平板、桌面)
- ✅ 即時表單驗證
- 🔄 載入狀態顯示
- 💬 友善的錯誤提示
- 🎯 清晰的成功訊息

## 技術棧

- **Nuxt 3** - Vue 3 框架
- **Vue 3 Composition API** - 使用 `<script setup>` 語法
- **原生 CSS** - 不依賴任何 UI 框架
- **Fetch API** - 用於 API 呼叫

## 開發指令

```bash
# 安裝依賴
npm install

# 啟動開發伺服器 (http://localhost:3000)
npm run dev

# 建置生產版本
npm run build

# 預覽生產版本
npm run preview

# 生成靜態網站
npm run generate
```

## 專案結構

```
frontend/
├── pages/
│   └── index.vue        # 註冊頁面
├── app.vue              # 根組件
├── nuxt.config.ts       # Nuxt 配置
└── package.json
```

## 配置

在 `nuxt.config.ts` 中可以修改 API 端點：

```typescript
runtimeConfig: {
  public: {
    apiBase: 'http://localhost:8080'  // 修改這裡
  }
}
```

## 表單驗證規則

- **使用者名稱**: 必填，至少 3 個字元
- **電子郵件**: 必填，需符合 email 格式
- **密碼**: 必填，至少 6 個字元

## UI 設計特色

### 配色方案
- 主色調：紫色漸層 (#667eea → #764ba2)
- 背景：全螢幕漸層背景
- 表單：白色卡片，圓角陰影
- 錯誤：紅色 (#f56565)
- 成功：綠色 (#22543d)

### 互動效果
- 輸入框 focus 時的邊框動畫
- 按鈕 hover 時的上浮效果
- 按鈕 active 時的下壓回饋
- 載入中的禁用狀態

### 響應式設計
- 桌面版：最大寬度 480px，居中顯示
- 手機版：自動調整內距和字體大小
- 斷點：640px

## API 整合

前端使用 Fetch API 與後端通訊：

```javascript
const response = await fetch(`${config.public.apiBase}/api/register`, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    username: 'john_doe',
    password: 'password123',
    email: 'john@example.com'
  })
})
```

## 錯誤處理

前端處理以下錯誤情況：
1. 表單驗證錯誤（本地驗證）
2. API 回傳的錯誤（如用戶名已存在）
3. 網路連線錯誤

## 瀏覽器支援

- Chrome (最新版)
- Firefox (最新版)
- Safari (最新版)
- Edge (最新版)

## License

MIT
