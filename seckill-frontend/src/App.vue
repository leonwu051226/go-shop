<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { useUserStore } from './store/user'

type ToastKind = 'info' | 'success' | 'warning' | 'error'

interface ToastPayload {
  message: string
  type?: ToastKind
}

const router = useRouter()
const userStore = useUserStore()
const toast = ref<ToastPayload | null>(null)
let toastTimer: number | undefined

const isMerchant = computed(() => userStore.role === 1)
const isLoggedIn = computed(() => Boolean(userStore.token))
const toastClasses = computed(() => {
  if (toast.value?.type === 'error') {
    return 'border-neon-pink/60 bg-neon-pink/20 text-white'
  }

  if (toast.value?.type === 'success') {
    return 'border-neon-cyan/60 bg-neon-cyan/20 text-white'
  }

  if (toast.value?.type === 'warning') {
    return 'border-neon-yellow/60 bg-neon-yellow/20 text-white'
  }

  return 'border-white/10 bg-white/10 text-white'
})

const showToast = (event: Event) => {
  const payload = (event as CustomEvent<ToastPayload>).detail
  toast.value = {
    message: payload?.message ?? '系统提示',
    type: payload?.type ?? 'info',
  }

  if (toastTimer) {
    window.clearTimeout(toastTimer)
  }

  toastTimer = window.setTimeout(() => {
    toast.value = null
  }, 2800)
}

const logout = () => {
  userStore.clearAuth()
  window.dispatchEvent(
    new CustomEvent<ToastPayload>('app-toast', {
      detail: { message: '身份凭证已清除', type: 'success' },
    }),
  )
  router.push({ name: 'login' })
}

onMounted(() => {
  window.addEventListener('app-toast', showToast)
})

onBeforeUnmount(() => {
  window.removeEventListener('app-toast', showToast)
})
</script>

<template>
  <div class="min-h-screen text-white">
    <header class="sticky top-0 z-40 border-b border-white/10 bg-abyss/70 backdrop-blur-xl">
      <nav class="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        <RouterLink to="/" class="cyber-title font-display text-xl font-bold">
          SECKILL://MALL
        </RouterLink>

        <div class="flex items-center gap-2">
          <RouterLink
            to="/"
            class="rounded-lg px-3 py-2 text-sm font-semibold text-white/70 transition hover:text-neon-cyan"
            active-class="text-neon-cyan"
          >
            首页
          </RouterLink>
          <RouterLink
            to="/profile"
            class="rounded-lg px-3 py-2 text-sm font-semibold text-white/70 transition hover:text-neon-cyan"
            active-class="text-neon-cyan"
          >
            我的
          </RouterLink>
          <RouterLink
            v-if="isMerchant"
            to="/merchant"
            class="rounded-lg px-3 py-2 text-sm font-semibold text-white/70 transition hover:text-neon-pink"
            active-class="text-neon-pink"
          >
            商家
          </RouterLink>
        </div>

        <button v-if="isLoggedIn" class="ghost-button hidden sm:inline-flex" type="button" @click="logout">
          退出
        </button>
        <RouterLink v-else to="/login" class="ghost-button hidden sm:inline-flex">登录</RouterLink>
      </nav>
    </header>

    <main class="mx-auto min-h-[calc(100vh-4rem)] max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <RouterView />
    </main>

    <Transition name="fade">
      <div v-if="toast" class="toast-surface" :class="toastClasses">
        {{ toast.message }}
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
