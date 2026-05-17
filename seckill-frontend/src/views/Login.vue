<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore, type UserRole } from '../store/user'
import {
  getCaptcha,
  loginUser,
  registerUser,
  type ApiResponse,
  type AuthResponse,
  type CaptchaResponse,
} from '../utils/request'
import { toast } from '../utils/toast'

type CaptchaEnvelope = ApiResponse<CaptchaResponse> | CaptchaResponse

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const selectedRole = ref<UserRole>(0)
const isRegisterMode = ref(false)
const isSubmitting = ref(false)
const isCaptchaLoading = ref(false)
const captchaId = ref('')
const captchaBase64 = ref('')
const captchaCode = ref('')
const credentials = reactive({
  account: '',
  password: '',
  confirmPassword: '',
})

const roleOptions: Array<{ label: string; value: UserRole }> = [
  { label: '普通买家', value: 0 },
  { label: '商家', value: 1 },
]

const normalizeCaptchaImage = (image: string) => {
  if (!image) {
    return ''
  }

  return image.startsWith('data:image') ? image : `data:image/png;base64,${image}`
}

const pickCaptchaPayload = (payload: CaptchaEnvelope): CaptchaResponse => {
  const envelope = payload as ApiResponse<CaptchaResponse>
  return envelope.data ? envelope.data : payload as CaptchaResponse
}

const getErrorStatus = (error: unknown) => {
  return (error as { response?: { status?: number } }).response?.status
}

const handleCaptchaFailure = async (error: unknown) => {
  if (getErrorStatus(error) === 400) {
    captchaCode.value = ''
    await refreshCaptcha()
  }
}

const refreshCaptcha = async () => {
  isCaptchaLoading.value = true

  try {
    const response = await getCaptcha()
    const payload = pickCaptchaPayload(response.data)
    const image =
      payload.captcha_image ?? payload.captchaImage ?? payload.b64s ?? payload.image ?? ''

    captchaId.value = payload.captcha_id ?? payload.captchaId ?? payload.id ?? ''
    captchaBase64.value = normalizeCaptchaImage(image)
    captchaCode.value = ''
  } catch {
    toast('验证码加载失败，请稍后重试', 'error')
  } finally {
    isCaptchaLoading.value = false
  }
}

const resetSecrets = () => {
  credentials.password = ''
  credentials.confirmPassword = ''
  captchaCode.value = ''
}

const toggleMode = () => {
  isRegisterMode.value = !isRegisterMode.value
  resetSecrets()
  refreshCaptcha()
}

const validateBaseFields = () => {
  if (!credentials.account.trim() || !credentials.password.trim()) {
    toast('请输入账号和密码', 'error')
    return false
  }

  if (!captchaId.value || !captchaCode.value.trim()) {
    toast('请输入神经校验码', 'error')
    return false
  }

  return true
}

const login = async () => {
  if (!validateBaseFields()) {
    return
  }

  isSubmitting.value = true

  try {
    const response = await loginUser({
      username: credentials.account.trim(),
      password: credentials.password.trim(),
      captcha_id: captchaId.value,
      captcha_code: captchaCode.value.trim(),
    })
    const auth: AuthResponse = response.data.data ?? {}
    const token = auth.token ?? auth.access_token
    const role = auth.role === 0 || auth.role === 1 ? auth.role : selectedRole.value
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'

    if (!token) {
      toast('登录成功但未收到 token', 'warning')
      await refreshCaptcha()
      return
    }

    userStore.setAuth(token, role)
    toast(`${role === 1 ? '商家' : '买家'}身份已接入`, 'success')
    router.push(redirect)
  } catch (error) {
    await handleCaptchaFailure(error)
  } finally {
    isSubmitting.value = false
  }
}

const register = async () => {
  if (!validateBaseFields()) {
    return
  }

  const password = credentials.password.trim()
  const confirmPassword = credentials.confirmPassword.trim()

  if (!confirmPassword) {
    toast('请再次输入密码', 'error')
    return
  }

  if (password !== confirmPassword) {
    toast('两次输入的密码不一致', 'error')
    return
  }

  isSubmitting.value = true

  try {
    await registerUser({
      username: credentials.account.trim(),
      password,
      captcha_id: captchaId.value,
      captcha_code: captchaCode.value.trim(),
    })
    toast('注册成功，请登录', 'success')
    isRegisterMode.value = false
    resetSecrets()
    await refreshCaptcha()
  } catch (error) {
    await handleCaptchaFailure(error)
  } finally {
    isSubmitting.value = false
  }
}

const submit = () => {
  if (isRegisterMode.value) {
    register()
    return
  }

  login()
}

onMounted(() => {
  refreshCaptcha()
})
</script>

<template>
  <section class="mx-auto flex min-h-[calc(100vh-8rem)] max-w-5xl items-center">
    <div class="grid w-full gap-6 lg:grid-cols-[1fr_0.9fr]">
      <div class="flex flex-col justify-center">
        <p class="mb-3 font-display text-sm font-bold uppercase tracking-[0.3em] text-neon-pink">
          Access Terminal
        </p>
        <h1 class="cyber-title max-w-2xl text-4xl font-black sm:text-6xl">
          秒杀商城身份接入
        </h1>
        <p class="mt-5 max-w-xl text-base leading-8 text-white/60">
          连接本地网关，选择买家或商家身份，输入神经校验码后进入赛博朋克秒杀商城操作台。
        </p>
      </div>

      <form class="glass-panel-strong p-6 sm:p-8" @submit.prevent="submit">
        <div class="mb-6">
          <p class="font-display text-sm font-bold uppercase tracking-[0.28em] text-neon-yellow">
            {{ isRegisterMode ? 'Register Node' : 'Login Node' }}
          </p>
          <h2 class="mt-2 font-display text-2xl font-black text-white">
            {{ isRegisterMode ? '创建神经档案' : '启动身份引擎' }}
          </h2>
        </div>

        <div class="mb-7 flex gap-3">
          <button
            v-for="role in roleOptions"
            :key="role.value"
            type="button"
            class="cyber-tab flex-1"
            :class="{ 'cyber-tab-active': selectedRole === role.value }"
            @click="selectedRole = role.value"
          >
            {{ role.label }}
          </button>
        </div>

        <label class="mb-5 block">
          <span class="mb-2 block text-sm font-semibold text-white/70">账号</span>
          <input
            v-model="credentials.account"
            class="cyber-input"
            autocomplete="username"
            placeholder="请输入账号"
          />
        </label>

        <label class="mb-5 block">
          <span class="mb-2 block text-sm font-semibold text-white/70">密码</span>
          <input
            v-model="credentials.password"
            class="cyber-input"
            type="password"
            :autocomplete="isRegisterMode ? 'new-password' : 'current-password'"
            placeholder="请输入密码"
          />
        </label>

        <Transition name="slide-glow">
          <label v-if="isRegisterMode" class="mb-5 block">
            <span class="mb-2 block text-sm font-semibold text-white/70">确认密码</span>
            <input
              v-model="credentials.confirmPassword"
              class="cyber-input"
              type="password"
              autocomplete="new-password"
              placeholder="再次输入密码"
            />
          </label>
        </Transition>

        <div class="mb-7 grid gap-3 sm:grid-cols-[1fr_150px]">
          <input
            v-model="captchaCode"
            class="cyber-input"
            autocomplete="off"
            placeholder="输入神经校验码 (Captcha)"
          />
          <button
            class="glass-panel flex h-12 items-center justify-center overflow-hidden rounded-lg transition-shadow hover:shadow-[0_0_10px_#00f3ff]"
            title="点击刷新验证码"
            type="button"
            @click="refreshCaptcha"
          >
            <img
              v-if="captchaBase64 && !isCaptchaLoading"
              :src="captchaBase64"
              alt="Captcha"
              class="h-full w-full cursor-pointer object-cover"
            />
            <span v-else class="text-xs font-semibold text-white/55">
              {{ isCaptchaLoading ? '加载中' : '刷新验证码' }}
            </span>
          </button>
        </div>

        <button class="engine-button w-full" :disabled="isSubmitting" type="submit">
          {{ isSubmitting ? '传输中' : isRegisterMode ? '创建档案 / Register' : '启动引擎 / Login' }}
        </button>

        <button
          class="mt-5 w-full cursor-pointer text-center text-sm font-semibold text-neon-cyan transition-colors hover:text-neon-pink"
          type="button"
          @click="toggleMode"
        >
          {{ isRegisterMode ? '已有档案？返回登录 (Back to Login)' : '新来的？创建神经链接 (Register)' }}
        </button>
      </form>
    </div>
  </section>
</template>

<style scoped>
.slide-glow-enter-active,
.slide-glow-leave-active {
  overflow: hidden;
  transition:
    opacity 0.25s ease,
    transform 0.25s ease,
    max-height 0.25s ease;
}

.slide-glow-enter-from,
.slide-glow-leave-to {
  max-height: 0;
  opacity: 0;
  transform: translateY(-10px);
}

.slide-glow-enter-to,
.slide-glow-leave-from {
  max-height: 96px;
  opacity: 1;
  transform: translateY(0);
}
</style>
