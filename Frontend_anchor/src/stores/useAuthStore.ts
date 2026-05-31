import { create } from 'zustand'
import { clearSession, getCurrentUser, login } from '@/api/auth'
import { getAccessToken } from '@/api/client'
import type { AuthUser, LoginPayload } from '@/api/auth'

const isMerchantRole = (role?: string) => role === 'anchor' || role === 'admin'

interface AuthState {
  user: AuthUser | null
  token: string | null
  loading: boolean
  hydrated: boolean
  loginWithPassword: (payload: LoginPayload) => Promise<void>
  restoreSession: () => Promise<void>
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: getAccessToken(),
  loading: false,
  hydrated: false,
  loginWithPassword: async (payload) => {
    set({ loading: true })
    try {
      const result = await login(payload)
      if (!isMerchantRole(result.user.role)) {
        clearSession()
        set({ user: null, token: null, loading: false, hydrated: true })
        throw new Error('仅主播或管理员账号可登录商家端')
      }
      set({ user: result.user, token: result.token, loading: false, hydrated: true })
    } catch (error) {
      set({ loading: false, hydrated: true })
      throw error
    }
  },
  restoreSession: async () => {
    const token = getAccessToken()
    if (!token) {
      set({ user: null, token: null, hydrated: true })
      return
    }

    set({ loading: true })
    try {
      const user = await getCurrentUser()
      if (!isMerchantRole(user.role)) {
        clearSession()
        set({ user: null, token: null, loading: false, hydrated: true })
        return
      }
      set({ user, token, loading: false, hydrated: true })
    } catch {
      clearSession()
      set({ user: null, token: null, loading: false, hydrated: true })
    }
  },
  logout: () => {
    clearSession()
    set({ user: null, token: null, hydrated: true, loading: false })
  },
}))
