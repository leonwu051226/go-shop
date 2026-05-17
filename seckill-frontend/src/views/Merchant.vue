<script setup lang="ts">
import { reactive, ref } from 'vue'
import request from '../utils/request'
import { toast } from '../utils/toast'

const isSubmitting = ref(false)
const form = reactive({
  name: '',
  price: '',
  stock: 100,
  imageUrl: '',
})

const submitProduct = async () => {
  isSubmitting.value = true

  try {
    await request.post('/products', {
      name: form.name,
      price: Number(form.price),
      stock: Number(form.stock),
      imageUrl: form.imageUrl,
    })
    toast('商品新增请求已发送', 'success')
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <section class="mx-auto max-w-4xl">
    <div class="mb-8">
      <p class="font-display text-sm font-bold uppercase tracking-[0.3em] text-neon-yellow">
        Merchant Console
      </p>
      <h1 class="cyber-title mt-3 text-4xl font-black">商家控制台</h1>
    </div>

    <form class="glass-panel-strong grid gap-5 p-6 sm:p-8" @submit.prevent="submitProduct">
      <label>
        <span class="mb-2 block text-sm font-semibold text-white/70">商品名称</span>
        <input v-model="form.name" class="cyber-input" placeholder="输入商品名称" required />
      </label>

      <div class="grid gap-5 sm:grid-cols-2">
        <label>
          <span class="mb-2 block text-sm font-semibold text-white/70">秒杀价格</span>
          <input
            v-model="form.price"
            class="cyber-input"
            min="0"
            placeholder="0.00"
            required
            step="0.01"
            type="number"
          />
        </label>
        <label>
          <span class="mb-2 block text-sm font-semibold text-white/70">库存</span>
          <input v-model.number="form.stock" class="cyber-input" min="0" required type="number" />
        </label>
      </div>

      <label>
        <span class="mb-2 block text-sm font-semibold text-white/70">商品图地址</span>
        <input v-model="form.imageUrl" class="cyber-input" placeholder="https://example.com/product.png" />
      </label>

      <div class="flex justify-end">
        <button class="engine-button min-w-[9rem]" :disabled="isSubmitting" type="submit">
          {{ isSubmitting ? '传输中' : '新增商品' }}
        </button>
      </div>
    </form>
  </section>
</template>
