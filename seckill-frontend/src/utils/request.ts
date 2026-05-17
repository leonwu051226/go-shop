import axios, { type AxiosError, type AxiosResponse } from 'axios'
import router from '../router'
import { useUserStore } from '../store/user'
import { toast } from './toast'

export interface ApiResponse<T = unknown> {
  code?: number
  message?: string
  data?: T
}

export interface LoginPayload {
  username: string
  password: string
  captcha_id: string
  captcha_code: string
}

export interface RegisterPayload {
  username: string
  password: string
  captcha_id: string
  captcha_code: string
}

export interface AuthResponse {
  token?: string
  access_token?: string
  role?: 0 | 1
}

export interface CaptchaResponse {
  captcha_id?: string
  captchaId?: string
  id?: string
  captcha_image?: string
  captchaImage?: string
  b64s?: string
  image?: string
}

export interface ProductQuery {
  keyword?: string
}

export interface ProductItem {
  id: number | string
  name?: string
  title?: string
  description?: string
  price?: number | string
  stock?: number
  image_url?: string
  imageUrl?: string
  tag?: string
}

export type ProductListResponse =
  | ProductItem[]
  | {
      list?: ProductItem[]
      records?: ProductItem[]
      items?: ProductItem[]
      total?: number
    }

const request = axios.create({
  baseURL: 'http://localhost:30080',
  timeout: 10000,
})

request.interceptors.request.use((config) => {
  const userStore = useUserStore()

  if (userStore.token) {
    config.headers.Authorization = `Bearer ${userStore.token}`
  }

  return config
})

request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => response,
  (error: AxiosError<ApiResponse>) => {
    const status = error.response?.status
    const message = error.response?.data?.message ?? error.message ?? '网络请求异常'

    if (status === 401) {
      const userStore = useUserStore()
      userStore.clearAuth()
      toast('登录状态已过期，请重新登录', 'warning')
      router.push({ name: 'login', query: { redirect: router.currentRoute.value.fullPath } })
    } else if (status === 403) {
      toast('当前身份无权访问该资源', 'error')
    } else {
      toast(message, 'error')
    }

    return Promise.reject(error)
  },
)

export const getCaptcha = () => {
  return request.get<ApiResponse<CaptchaResponse> | CaptchaResponse>('/api/v1/captcha')
}

export const loginUser = (payload: LoginPayload) => {
  return request.post<ApiResponse<AuthResponse>>('/api/v1/user/login', payload)
}

export const registerUser = (payload: RegisterPayload) => {
  return request.post<ApiResponse>('/api/v1/user/register', payload)
}

export const getProducts = (params: ProductQuery = {}) => {
  return request.get<ApiResponse<ProductListResponse> | ProductListResponse>('/api/v1/products', {
    params,
  })
}

export default request
