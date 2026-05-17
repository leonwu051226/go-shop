<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { getProducts, type ApiResponse, type ProductItem, type ProductListResponse } from '../utils/request'
import { toast } from '../utils/toast'

interface ProductCard {
  id: number | string
  name: string
  description: string
  tag: string
  price: string
  stock: number
  imageUrl: string
  accent: string
}

type ProductEnvelope = ApiResponse<ProductListResponse> | ProductListResponse

const fallbackProducts: ProductCard[] = [
  {
    id: 1,
    name: 'Neon Pulse 机械键盘',
    description: '低延迟轴体与霓虹背光，适合秒杀战场的高速输入。',
    tag: '限时爆破',
    price: '299.00',
    stock: 128,
    imageUrl: '',
    accent: 'from-neon-cyan/30 to-neon-pink/25',
  },
  {
    id: 2,
    name: 'Zero Latency 游戏耳机',
    description: '沉浸式声场与轻量头梁，捕捉每一次补货信号。',
    tag: '低延迟',
    price: '199.00',
    stock: 64,
    imageUrl: '',
    accent: 'from-neon-pink/30 to-neon-yellow/20',
  },
  {
    id: 3,
    name: 'Cyber Flask 保温杯',
    description: '城市补给模块，长时保温，随时保持能量上线。',
    tag: '城市补给',
    price: '59.90',
    stock: 320,
    imageUrl: '',
    accent: 'from-neon-yellow/30 to-neon-cyan/20',
  },
  {
    id: 4,
    name: 'Glass Grid 移动电源',
    description: '透明能量核心，给所有终端提供持续续航。',
    tag: '能量核心',
    price: '129.00',
    stock: 88,
    imageUrl: '',
    accent: 'from-white/20 to-neon-cyan/30',
  },
]

const searchKeyword = ref('')
const products = ref<ProductCard[]>(fallbackProducts)
const isLoading = ref(false)
let searchTimer: number | undefined

const accentPool = [
  'from-neon-cyan/30 to-neon-pink/25',
  'from-neon-pink/30 to-neon-yellow/20',
  'from-neon-yellow/30 to-neon-cyan/20',
  'from-white/20 to-neon-cyan/30',
]

const escapeHtml = (value: string) => {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;')
}

const safeHighlightedHtml = (value: string) => {
  return escapeHtml(value)
    .replaceAll('&lt;em&gt;', '<em>')
    .replaceAll('&lt;/em&gt;', '</em>')
}

const pickProductPayload = (payload: ProductEnvelope): ProductListResponse => {
  const envelope = payload as ApiResponse<ProductListResponse>
  return envelope.data ? envelope.data : payload as ProductListResponse
}

const normalizeProductList = (payload: ProductListResponse): ProductItem[] => {
  if (Array.isArray(payload)) {
    return payload
  }

  return payload.list ?? payload.records ?? payload.items ?? []
}

const normalizeProduct = (product: ProductItem, index: number): ProductCard => ({
  id: product.id,
  name: product.name ?? product.title ?? 'Unnamed Neural Goods',
  description: product.description ?? 'No signal description',
  tag: product.tag ?? 'ES MATCH',
  price: String(product.price ?? '0.00'),
  stock: product.stock ?? 0,
  imageUrl: product.image_url ?? product.imageUrl ?? '',
  accent: accentPool[index % accentPool.length],
})

const fetchProducts = async () => {
  const keyword = searchKeyword.value.trim()
  isLoading.value = true

  try {
    const response = await getProducts(keyword ? { keyword } : {})
    const list = normalizeProductList(pickProductPayload(response.data))
    products.value = list.map(normalizeProduct)
  } catch {
    toast('商品搜索链路暂时不可用', 'error')
    products.value = keyword ? [] : fallbackProducts
  } finally {
    isLoading.value = false
  }
}

const debounceSearch = () => {
  if (searchTimer) {
    window.clearTimeout(searchTimer)
  }

  searchTimer = window.setTimeout(fetchProducts, 360)
}

onMounted(() => {
  fetchProducts()
})

onBeforeUnmount(() => {
  if (searchTimer) {
    window.clearTimeout(searchTimer)
  }
})
</script>

<template>
  <section>
    <div class="mx-auto mb-8 max-w-3xl">
      <label class="glass-panel flex items-center gap-3 px-4 py-3">
        <span class="font-display text-sm font-bold text-neon-cyan">SEARCH</span>
        <input
          v-model="searchKeyword"
          class="w-full bg-transparent text-sm text-white outline-none placeholder:text-white/35 focus:shadow-[0_0_15px_#00f3ff]"
          placeholder="输入商品关键词，回车或停顿后搜索 / ES Neural Search"
          type="search"
          @input="debounceSearch"
          @keydown.enter.prevent="fetchProducts"
        />
      </label>
    </div>

    <div class="mb-8 flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
      <div>
        <p class="font-display text-sm font-bold uppercase tracking-[0.3em] text-neon-yellow">
          Flash Sale Grid
        </p>
        <h1 class="cyber-title mt-3 text-4xl font-black sm:text-5xl">
          今日秒杀
        </h1>
      </div>
      <div class="glass-panel px-4 py-3 text-sm text-white/60">
        网关：<span class="text-neon-cyan">http://localhost:30080</span>
      </div>
    </div>

    <div v-if="isLoading" class="glass-panel py-14 text-center font-mono text-sm text-neon-cyan">
      [ SYSTEM SCANNING ] 正在检索商品神经信号...
    </div>

    <div v-else-if="products.length === 0" class="glass-panel py-14 text-center font-mono text-sm text-white/40">
      [ SYSTEM WARNING ] 未检测到相关神经信号 / No Data Found
    </div>

    <div v-else class="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
      <article
        v-for="product in products"
        :key="product.id"
        class="glass-panel group overflow-hidden p-4 transition hover:-translate-y-1 hover:border-neon-cyan/40 hover:shadow-cyan"
      >
        <div
          class="mb-4 flex aspect-[4/3] items-center justify-center overflow-hidden rounded-lg border border-white/10 bg-gradient-to-br"
          :class="product.accent"
        >
          <img
            v-if="product.imageUrl"
            :src="product.imageUrl"
            :alt="product.name"
            class="h-full w-full object-cover"
          />
          <div v-else class="h-16 w-16 rotate-45 border border-white/30 bg-black/20 shadow-cyan backdrop-blur-md" />
        </div>

        <div class="mb-3 flex items-center justify-between gap-3">
          <span class="rounded-full border border-neon-pink/40 px-2 py-1 text-xs text-neon-pink">
            {{ product.tag }}
          </span>
          <span class="text-xs text-white/40">库存 {{ product.stock }}</span>
        </div>

        <h2
          class="highlighted-html min-h-12 text-base font-semibold leading-6 text-white"
          v-html="safeHighlightedHtml(product.name)"
        />

        <p
          class="highlighted-html mt-3 line-clamp-2 min-h-10 text-sm leading-5 text-white/55"
          v-html="safeHighlightedHtml(product.description)"
        />

        <div class="mt-5 flex items-center justify-between gap-3">
          <p class="neon-price text-2xl">¥{{ product.price }}</p>
          <button class="engine-button px-4 py-2" type="button">
            抢购
          </button>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.highlighted-html :deep(em) {
  color: #ff003c;
  font-style: normal;
  font-weight: 800;
  filter: drop-shadow(0 0 5px rgba(236, 72, 153, 0.8));
}
</style>
