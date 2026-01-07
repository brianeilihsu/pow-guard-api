<template>
  <div class="container">
    <div class="register-card">
      <h1 class="title">建立帳號</h1>
      <p class="subtitle">歡迎註冊 POW API 服務</p>

      <form @submit.prevent="handleSubmit" class="form">
        <!-- Username Input -->
        <div class="form-group">
          <label for="username" class="label">使用者名稱</label>
          <input
            id="username"
            v-model="formData.username"
            type="text"
            class="input"
            :class="{ 'input-error': errors.username }"
            placeholder="請輸入使用者名稱（至少 3 個字元）"
            @input="clearError('username')"
            :disabled="isLoading || isSolvingPoW"
          />
          <span v-if="errors.username" class="error-message">{{ errors.username }}</span>
        </div>

        <!-- Email Input -->
        <div class="form-group">
          <label for="email" class="label">電子郵件</label>
          <input
            id="email"
            v-model="formData.email"
            type="email"
            class="input"
            :class="{ 'input-error': errors.email }"
            placeholder="請輸入電子郵件"
            @input="clearError('email')"
            :disabled="isLoading || isSolvingPoW"
          />
          <span v-if="errors.email" class="error-message">{{ errors.email }}</span>
        </div>

        <!-- Password Input -->
        <div class="form-group">
          <label for="password" class="label">密碼</label>
          <input
            id="password"
            v-model="formData.password"
            type="password"
            class="input"
            :class="{ 'input-error': errors.password }"
            placeholder="請輸入密碼（至少 6 個字元）"
            @input="clearError('password')"
            :disabled="isLoading || isSolvingPoW"
          />
          <span v-if="errors.password" class="error-message">{{ errors.password }}</span>
        </div>

        <!-- PoW Progress -->
        <div v-if="isSolvingPoW" class="pow-progress">
          <div class="pow-spinner"></div>
          <div class="pow-info">
            <p class="pow-text">🔐 正在計算工作量證明...</p>
            <p class="pow-stats">已嘗試: {{ powAttempts }} 次</p>
          </div>
        </div>

        <!-- Success Message -->
        <div v-if="successMessage" class="success-message">
          {{ successMessage }}
        </div>

        <!-- Error Message -->
        <div v-if="generalError" class="general-error">
          {{ generalError }}
        </div>

        <!-- Submit Button -->
        <button type="submit" class="submit-button" :disabled="isLoading || isSolvingPoW">
          <span v-if="!isLoading && !isSolvingPoW">註冊</span>
          <span v-else-if="isSolvingPoW">計算中...</span>
          <span v-else>註冊中...</span>
        </button>
      </form>

      <div class="footer">
        已經有帳號了？<a href="#" class="link">登入</a>
      </div>

      <!-- PoW Info -->
      <div class="pow-badge">
        <span class="badge-icon">🛡️</span>
        <span class="badge-text">受 Proof of Work 保護</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'

const config = useRuntimeConfig()

const formData = reactive({
  username: '',
  email: '',
  password: ''
})

const errors = reactive({
  username: '',
  email: '',
  password: ''
})

const isLoading = ref(false)
const isSolvingPoW = ref(false)
const powAttempts = ref(0)
const successMessage = ref('')
const generalError = ref('')

const clearError = (field) => {
  errors[field] = ''
  generalError.value = ''
}

const validateForm = () => {
  let isValid = true

  // Reset errors
  errors.username = ''
  errors.email = ''
  errors.password = ''
  generalError.value = ''

  // Validate username
  if (!formData.username) {
    errors.username = '請輸入使用者名稱'
    isValid = false
  } else if (formData.username.length < 3) {
    errors.username = '使用者名稱至少需要 3 個字元'
    isValid = false
  }

  // Validate email
  if (!formData.email) {
    errors.email = '請輸入電子郵件'
    isValid = false
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
    errors.email = '請輸入有效的電子郵件格式'
    isValid = false
  }

  // Validate password
  if (!formData.password) {
    errors.password = '請輸入密碼'
    isValid = false
  } else if (formData.password.length < 6) {
    errors.password = '密碼至少需要 6 個字元'
    isValid = false
  }

  return isValid
}

// SHA-256 hash function
async function sha256(message) {
  const msgBuffer = new TextEncoder().encode(message)
  const hashBuffer = await crypto.subtle.digest('SHA-256', msgBuffer)
  const hashArray = Array.from(new Uint8Array(hashBuffer))
  const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('')
  return hashHex
}

// Solve Proof of Work
async function solvePoW(challenge, difficulty) {
  isSolvingPoW.value = true
  powAttempts.value = 0

  const requiredPrefix = '0'.repeat(difficulty)
  let nonce = 0

  while (true) {
    const nonceStr = nonce.toString()
    const hash = await sha256(challenge + nonceStr)

    powAttempts.value = nonce + 1

    if (hash.startsWith(requiredPrefix)) {
      isSolvingPoW.value = false
      return nonceStr
    }

    nonce++

    // Yield to browser every 1000 attempts to prevent freezing
    if (nonce % 1000 === 0) {
      await new Promise(resolve => setTimeout(resolve, 0))
    }
  }
}

const handleSubmit = async () => {
  if (!validateForm()) {
    return
  }

  successMessage.value = ''
  generalError.value = ''

  try {
    // Step 1: Get challenge from server
    const challengeResponse = await fetch(`${config.public.apiBase}/api/challenge`, {
      method: 'GET'
    })

    if (!challengeResponse.ok) {
      throw new Error('Failed to get challenge')
    }

    const challengeData = await challengeResponse.json()

    if (!challengeData.success) {
      throw new Error(challengeData.message || 'Failed to get challenge')
    }

    const { challenge, difficulty } = challengeData.data

    console.log(`Solving PoW with difficulty ${difficulty}...`)

    // Step 2: Solve Proof of Work
    const nonce = await solvePoW(challenge, difficulty)

    console.log(`PoW solved! Nonce: ${nonce}, Attempts: ${powAttempts.value}`)

    // Step 3: Submit registration with PoW
    isLoading.value = true

    const response = await fetch(`${config.public.apiBase}/api/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        ...formData,
        challenge,
        nonce
      })
    })

    const data = await response.json()

    if (response.ok && data.success) {
      successMessage.value = `註冊成功！歡迎加入 POW API (PoW 嘗試次數: ${powAttempts.value})`
      // Reset form
      formData.username = ''
      formData.email = ''
      formData.password = ''
      powAttempts.value = 0
    } else {
      generalError.value = data.message || '註冊失敗，請稍後再試'
    }
  } catch (error) {
    generalError.value = '無法連接到伺服器，請稍後再試'
    console.error('Registration error:', error)
  } finally {
    isLoading.value = false
    isSolvingPoW.value = false
  }
}
</script>

<style scoped>
* {
  box-sizing: border-box;
}

.container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', sans-serif;
}

.register-card {
  background: white;
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  padding: 48px;
  width: 100%;
  max-width: 480px;
  position: relative;
}

.title {
  font-size: 32px;
  font-weight: 700;
  color: #1a202c;
  margin: 0 0 8px 0;
  text-align: center;
}

.subtitle {
  font-size: 16px;
  color: #718096;
  margin: 0 0 32px 0;
  text-align: center;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.label {
  font-size: 14px;
  font-weight: 600;
  color: #2d3748;
}

.input {
  padding: 12px 16px;
  font-size: 16px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  transition: all 0.2s;
  outline: none;
}

.input:focus {
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.input:disabled {
  background-color: #f7fafc;
  cursor: not-allowed;
  opacity: 0.6;
}

.input-error {
  border-color: #f56565;
}

.input-error:focus {
  border-color: #f56565;
  box-shadow: 0 0 0 3px rgba(245, 101, 101, 0.1);
}

.error-message {
  font-size: 14px;
  color: #f56565;
}

.pow-progress {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: linear-gradient(135deg, #f0f4ff 0%, #f5f0ff 100%);
  border-radius: 8px;
  border: 2px solid #667eea;
}

.pow-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid #e2e8f0;
  border-top-color: #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.pow-info {
  flex: 1;
}

.pow-text {
  margin: 0 0 4px 0;
  font-size: 14px;
  font-weight: 600;
  color: #4c51bf;
}

.pow-stats {
  margin: 0;
  font-size: 12px;
  color: #718096;
}

.success-message {
  padding: 12px 16px;
  background-color: #c6f6d5;
  color: #22543d;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
}

.general-error {
  padding: 12px 16px;
  background-color: #fed7d7;
  color: #742a2a;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
}

.submit-button {
  padding: 14px 24px;
  font-size: 16px;
  font-weight: 600;
  color: white;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  margin-top: 8px;
}

.submit-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 10px 20px rgba(102, 126, 234, 0.3);
}

.submit-button:active:not(:disabled) {
  transform: translateY(0);
}

.submit-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.footer {
  margin-top: 24px;
  text-align: center;
  font-size: 14px;
  color: #718096;
}

.link {
  color: #667eea;
  text-decoration: none;
  font-weight: 600;
  margin-left: 4px;
}

.link:hover {
  text-decoration: underline;
}

.pow-badge {
  margin-top: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 8px 16px;
  background: linear-gradient(135deg, #f0f4ff 0%, #f5f0ff 100%);
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.badge-icon {
  font-size: 16px;
}

.badge-text {
  font-size: 12px;
  font-weight: 600;
  color: #4c51bf;
}

@media (max-width: 640px) {
  .register-card {
    padding: 32px 24px;
  }

  .title {
    font-size: 28px;
  }
}
</style>
