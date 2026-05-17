<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useUserStore } from '../store/user'

interface Order {
  id: string
  name: string
  amount: string
  status: '未支付' | '已锁单' | '已完成' | '已取消'
  created_at: string
}

interface OrderView extends Order {
  remainingLabel: string
  isUnpaid: boolean
  isExpired: boolean
}

const PAYMENT_WINDOW_MS = 15 * 60 * 1000
const userStore = useUserStore()
const now = ref(Date.now())
let ticker: number | undefined

const roleLabel = computed(() => (userStore.role === 1 ? '商家' : '普通买家'))
const orders: Order[] = [
  {
    id: 'SO-20260517-001',
    name: 'Neon Pulse 机械键盘',
    amount: '299.00',
    status: '未支付',
    created_at: new Date(Date.now() - 2 * 60 * 1000).toISOString(),
  },
  {
    id: 'SO-20260517-002',
    name: 'Cyber Flask 保温杯',
    amount: '59.90',
    status: '已取消',
    created_at: new Date(Date.now() - 20 * 60 * 1000).toISOString(),
  },
  {
    id: 'SO-20260517-003',
    name: 'Glass Grid 移动电源',
    amount: '129.00',
    status: '已完成',
    created_at: new Date(Date.now() - 42 * 60 * 1000).toISOString(),
  },
]

const formatRemaining = (createdAt: string) => {
  const elapsed = now.value - new Date(createdAt).getTime()
  const remaining = Math.max(PAYMENT_WINDOW_MS - elapsed, 0)
  const minutes = Math.floor(remaining / 60_000)
  const seconds = Math.floor((remaining % 60_000) / 1000)

  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
}

const orderViews = computed<OrderView[]>(() =>
  orders.map((order) => ({
    ...order,
    remainingLabel: formatRemaining(order.created_at),
    isUnpaid: order.status === '未支付',
    isExpired: order.status === '已取消',
  })),
)

onMounted(() => {
  ticker = window.setInterval(() => {
    now.value = Date.now()
  }, 1000)
})

onBeforeUnmount(() => {
  if (ticker) {
    window.clearInterval(ticker)
  }
})
</script>

<template>
  <section class="grid gap-6 lg:grid-cols-[0.75fr_1.25fr]">
    <aside class="glass-panel-strong p-6">
      <p class="font-display text-sm font-bold uppercase tracking-[0.3em] text-neon-pink">
        User Core
      </p>
      <h1 class="cyber-title mt-3 text-3xl font-black">个人中心</h1>

      <div class="mt-8 space-y-4">
        <div class="rounded-lg border border-white/10 bg-black/20 p-4">
          <p class="text-xs text-white/40">当前身份</p>
          <p class="mt-2 text-xl font-bold text-neon-cyan">{{ roleLabel }}</p>
        </div>
        <div class="rounded-lg border border-white/10 bg-black/20 p-4">
          <p class="text-xs text-white/40">Token</p>
          <p class="mt-2 break-all text-sm text-white/70">{{ userStore.token }}</p>
        </div>
      </div>
    </aside>

    <div class="glass-panel p-6">
      <div class="mb-5 flex items-center justify-between gap-3">
        <h2 class="font-display text-2xl font-bold text-white">我的订单</h2>
        <span class="rounded-full border border-neon-yellow/40 px-3 py-1 text-sm text-neon-yellow">
          {{ orderViews.length }} 单
        </span>
      </div>

      <div v-if="orderViews.length === 0" class="rounded-lg border border-white/10 bg-black/20 py-14 text-center">
        <p class="font-mono text-sm text-white/40">
          [ SYSTEM WARNING ] 未检测到订单神经信号 / No Data Found
        </p>
      </div>

      <div v-else class="space-y-3">
        <article
          v-for="order in orderViews"
          :key="order.id"
          class="relative grid gap-3 overflow-hidden rounded-lg border border-white/10 bg-black/20 p-4 transition sm:grid-cols-[1fr_auto_auto] sm:items-center"
          :class="{ 'opacity-45 grayscale': order.isExpired }"
        >
          <div
            v-if="order.isExpired"
            class="pointer-events-none absolute inset-0 z-10 flex items-center justify-center"
          >
            <span class="-rotate-12 rounded border border-white/10 bg-black/50 px-5 py-2 font-mono text-lg font-black tracking-[0.25em] text-white/30">
              已失效 / EXPIRED
            </span>
          </div>

          <div>
            <p class="text-sm text-white/40">{{ order.id }}</p>
            <h3 class="mt-1 font-semibold text-white">{{ order.name }}</h3>
            <p class="mt-1 text-xs text-white/35">创建时间 {{ new Date(order.created_at).toLocaleString() }}</p>
          </div>

          <div class="text-left sm:text-right">
            <p class="neon-price text-xl">¥{{ order.amount }}</p>
            <p
              v-if="order.isUnpaid"
              class="mt-2 animate-pulse font-mono text-sm font-bold text-red-500 drop-shadow-[0_0_8px_#ef4444]"
            >
              PAY IN {{ order.remainingLabel }}
            </p>
          </div>

          <span
            class="rounded-full border px-3 py-1 text-center text-sm"
            :class="order.isExpired ? 'border-white/20 text-white/35' : 'border-neon-cyan/40 text-neon-cyan'"
          >
            {{ order.status }}
          </span>
        </article>
      </div>
    </div>
  </section>
</template>
