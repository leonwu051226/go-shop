<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import axios from '../api/axios.js'

const router = useRouter()

const countdown = ref('00:00:00')
let timer = null
let countdownTimer = null

const products = ref([
  {
    id: 1,
    name: 'Apple iPhone 15 Pro Max 256GB',
    image: 'https://images.unsplash.com/photo-1696446701796-da61225697cc?w=400&h=400&fit=crop',
    originalPrice: 9999,
    seckillPrice: 7999,
    progress: 85,
    soldOut: false
  },
  {
    id: 2,
    name: 'Sony WH-1000XM5 降噪耳机',
    image: 'https://images.unsplash.com/photo-1618366712010-f4ae9c647dcb?w=400&h=400&fit=crop',
    originalPrice: 2999,
    seckillPrice: 1899,
    progress: 62,
    soldOut: false
  },
  {
    id: 3,
    name: 'Nintendo Switch OLED 版',
    image: 'https://images.unsplash.com/photo-1578303512597-81e6cc155b3e?w=400&h=400&fit=crop',
    originalPrice: 2599,
    seckillPrice: 1899,
    progress: 100,
    soldOut: true
  },
  {
    id: 4,
    name: 'Dyson V15 Detect 吸尘器',
    image: 'https://images.unsplash.com/photo-1558317374-a3545eca46f2?w=400&h=400&fit=crop',
    originalPrice: 5499,
    seckillPrice: 3999,
    progress: 45,
    soldOut: false
  },
  {
    id: 5,
    name: '小米 14 Ultra 16+512GB',
    image: 'https://images.unsplash.com/photo-1511707171634-5f897ff02aa9?w=400&h=400&fit=crop',
    originalPrice: 6999,
    seckillPrice: 5999,
    progress: 78,
    soldOut: false
  },
  {
    id: 6,
    name: 'SK-II 神仙水 230ml',
    image: 'https://images.unsplash.com/photo-1620916566398-39f1143ab7be?w=400&h=400&fit=crop',
    originalPrice: 1540,
    seckillPrice: 899,
    progress: 92,
    soldOut: false
  }
])

const updateCountdown = () => {
  const now = new Date()
  const end = new Date()
  end.setHours(22, 0, 0, 0)
  if (end <= now) {
    end.setDate(end.getDate() + 1)
    end.setHours(22, 0, 0, 0)
  }
  const diff = end - now
  const hours = String(Math.floor(diff / 3600000)).padStart(2, '0')
  const minutes = String(Math.floor((diff % 3600000) / 60000)).padStart(2, '0')
  const seconds = String(Math.floor((diff % 60000) / 1000)).padStart(2, '0')
  countdown.value = `${hours}:${minutes}:${seconds}`
}

const dialog = ref({ show: false, title: '', message: '', type: 'success' })

const showDialog = (title, message, type = 'success') => {
  dialog.value = { show: true, title, message, type }
  setTimeout(() => {
    dialog.value.show = false
  }, 2500)
}

const handleBuy = async (product) => {
  if (product.soldOut) return
  try {
    const res = await axios.post('/api/v1/seckill/do', {
      seckill_product_id: product.id
    })
    showDialog('抢购成功', '抢购成功，排队发货中！', 'success')
  } catch (err) {
    const status = err.response?.status
    const msg = err.response?.data?.message || ''
    if (status === 403 || msg.includes('重复') || msg.includes('已购买')) {
      showDialog('提示', '您已购买过此商品，请勿重复下订单！', 'warning')
    } else if (msg.includes('库存') || msg.includes('不足') || status === 409) {
      product.soldOut = true
      showDialog('抱歉', '该商品已抢光', 'error')
    } else {
      showDialog('错误', msg || '抢购失败，请稍后重试', 'error')
    }
  }
}

const logout = () => {
  localStorage.removeItem('token')
  router.push('/login')
}

onMounted(() => {
  updateCountdown()
  countdownTimer = setInterval(updateCountdown, 1000)
})

onUnmounted(() => {
  clearInterval(timer)
  clearInterval(countdownTimer)
})
</script>

<template>
  <div class="min-h-screen bg-[#F4F4F4] pb-10">
    <!-- 顶部 Header -->
    <div class="bg-gradient-to-r from-[#FF5000] to-[#FF9000] text-white pb-6 pt-4 px-4 rounded-b-3xl shadow-md">
      <div class="max-w-3xl mx-auto flex items-center justify-between">
        <div>
          <h1 class="text-xl font-bold">百亿补贴秒杀会场</h1>
          <p class="text-white/90 text-xs mt-1">正品好价 限时抢购</p>
        </div>
        <button @click="logout" class="text-sm bg-white/20 hover:bg-white/30 px-3 py-1 rounded-full transition">
          退出登录
        </button>
      </div>

      <!-- 倒计时 -->
      <div class="max-w-3xl mx-auto mt-5 bg-white/20 backdrop-blur-sm rounded-xl p-4 flex items-center justify-center gap-3">
        <span class="text-sm font-medium">距离本场结束还剩</span>
        <div class="flex gap-1">
          <span
            v-for="(char, idx) in countdown"
            :key="idx"
            :class="char === ':' ? 'text-white font-bold text-lg' : 'bg-white text-[#FF5000] font-bold text-lg w-8 h-8 rounded-lg flex items-center justify-center shadow-sm'"
          >
            {{ char }}
          </span>
        </div>
      </div>
    </div>

    <!-- 商品列表 -->
    <div class="max-w-3xl mx-auto px-4 mt-6 space-y-4">
      <div
        v-for="product in products"
        :key="product.id"
        class="bg-white rounded-2xl p-4 shadow-sm hover:shadow-md transition flex gap-4"
      >
        <!-- 商品图 -->
        <div class="shrink-0">
          <img
            :src="product.image"
            :alt="product.name"
            class="w-28 h-28 rounded-xl object-cover bg-gray-100"
            loading="lazy"
          />
        </div>

        <!-- 商品信息 -->
        <div class="flex-1 flex flex-col justify-between min-w-0">
          <div>
            <h3 class="text-gray-900 font-semibold text-base leading-snug truncate">{{ product.name }}</h3>
            <div class="mt-2 flex items-baseline gap-2">
              <span class="text-gray-400 text-sm line-through">¥{{ product.originalPrice }}</span>
              <span class="text-[#FF5000] text-2xl font-bold">¥{{ product.seckillPrice }}</span>
            </div>
          </div>

          <!-- 进度条 -->
          <div class="mt-2">
            <div class="flex items-center justify-between text-xs mb-1">
              <span class="text-gray-500">已抢购 {{ product.progress }}%</span>
              <span v-if="product.soldOut" class="text-gray-400">已抢光</span>
            </div>
            <div class="h-2.5 bg-gray-100 rounded-full overflow-hidden">
              <div
                class="h-full rounded-full bg-gradient-to-r from-[#FF5000] to-[#FF9000] transition-all"
                :style="{ width: product.progress + '%', opacity: product.soldOut ? 0.4 : 1 }"
              />
            </div>
          </div>
        </div>

        <!-- 抢购按钮 -->
        <div class="shrink-0 flex items-end">
          <button
            @click="handleBuy(product)"
            :disabled="product.soldOut"
            class="px-5 py-2.5 rounded-full font-semibold text-sm shadow-sm transition transform active:scale-95 disabled:opacity-60 disabled:cursor-not-allowed"
            :class="
              product.soldOut
                ? 'bg-gray-200 text-gray-500'
                : 'bg-gradient-to-r from-[#FF5000] to-[#FF9000] text-white hover:shadow-md hover:-translate-y-0.5'
            "
          >
            {{ product.soldOut ? '已抢光' : '立即抢购' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 弹窗 -->
    <transition name="fade">
      <div v-if="dialog.show" class="fixed inset-0 z-50 flex items-center justify-center px-4">
        <div class="absolute inset-0 bg-black/40" @click="dialog.show = false" />
        <div class="relative bg-white rounded-2xl p-6 max-w-xs w-full text-center shadow-2xl transform transition-all">
          <div
            class="w-12 h-12 mx-auto rounded-full flex items-center justify-center mb-3"
            :class="
              dialog.type === 'success'
                ? 'bg-green-100 text-green-600'
                : dialog.type === 'warning'
                  ? 'bg-yellow-100 text-yellow-600'
                  : 'bg-red-100 text-red-600'
            "
          >
            <svg v-if="dialog.type === 'success'" class="w-6 h-6" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
            </svg>
            <svg v-else-if="dialog.type === 'warning'" class="w-6 h-6" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01M12 3a9 9 0 110 18 9 9 0 010-18z" />
            </svg>
            <svg v-else class="w-6 h-6" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
          <h3 class="text-lg font-bold text-gray-900">{{ dialog.title }}</h3>
          <p class="text-gray-500 text-sm mt-1">{{ dialog.message }}</p>
        </div>
      </div>
    </transition>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
