import { defineStore } from 'pinia'

export type UserRole = 0 | 1

interface UserState {
  token: string
  role: UserRole | null
}

const TOKEN_KEY = 'seckill_token'
const ROLE_KEY = 'seckill_role'

const readRole = (): UserRole | null => {
  const value = localStorage.getItem(ROLE_KEY)
  return value === '0' || value === '1' ? Number(value) as UserRole : null
}

export const useUserStore = defineStore('user', {
  state: (): UserState => ({
    token: localStorage.getItem(TOKEN_KEY) ?? '',
    role: readRole(),
  }),
  actions: {
    setAuth(token: string, role: UserRole) {
      this.token = token
      this.role = role
      localStorage.setItem(TOKEN_KEY, token)
      localStorage.setItem(ROLE_KEY, String(role))
    },
    clearAuth() {
      this.token = ''
      this.role = null
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(ROLE_KEY)
    },
  },
})
