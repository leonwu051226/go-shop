<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from '../api/axios.js'

const router = useRouter()
const username = ref('')
const password = ref('')
const loading = ref(false)
const errorMsg = ref('')

onMounted(() => {
  username.value = 'admin'
  password.value = '123456'
})

const handleLogin = async () => {
  errorMsg.value = ''
  if (!username.value || !password.value) {
    errorMsg.value = '请输入用户名和密码'
    return
  }
  loading.value = true
  try {
    const res = await axios.post('/api/v1/login', {
      username: username.value,
      password: password.value
    })
    const token = res.data?.data?.token || res.data?.token || res.data?.data?.accessToken
    if (token) {
      localStorage.setItem('token', token)
      router.push('/seckill')
    } else {
      errorMsg.value = '登录成功但未获取到 Token'
    }
  } catch (err) {
    errorMsg.value = err.response?.data?.message || '登录失败，请检查网络或账密'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-[#F4F4F4] px-4">
    <div class="w-full max-w-md bg-white rounded-2xl shadow-lg p-8">
      <div class="text-center mb-8">
        <h1 class="text-2xl font-bold bg-gradient-to-r from-[#FF5000] to-[#FF9000] bg-clip-text text-transparent">
          淘宝百亿补贴
        </h1>
        <p class="text-gray-500 mt-2 text-sm">秒杀活动登录</p>
      </div>

      <div class="space-y-5">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">用户名</label>
          <input
            v-model="username"
            type="text"
            placeholder="请输入用户名"
            class="w-full px-4 py-3 rounded-xl border border-gray-200 focus:outline-none focus:ring-2 focus:ring-[#FF5000] focus:border-transparent transition"
            @keyup.enter="handleLogin"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">密码</label>
          <input
            v-model="password"
            type="password"
            placeholder="请输入密码"
            class="w-full px-4 py-3 rounded-xl border border-gray-200 focus:outline-none focus:ring-2 focus:ring-[#FF5000] focus:border-transparent transition"
            @keyup.enter="handleLogin"
          />
        </div>

        <p v-if="errorMsg" class="text-red-500 text-sm">{{ errorMsg }}</p>

        <button
          @click="handleLogin"
          :disabled="loading"
          class="w-full py-3 rounded-xl bg-gradient-to-r from-[#FF5000] to-[#FF9000] text-white font-semibold text-lg shadow-md hover:shadow-lg transform hover:-translate-y-0.5 transition disabled:opacity-60 disabled:cursor-not-allowed"
        >
          {{ loading ? '登录中...' : '立即登录' }}
        </button>
      </div>
    </div>
  </div>
</template>
