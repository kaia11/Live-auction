import { create } from 'zustand'
import { User } from '@/types'
import { clearSession, getCurrentUser, login as loginRequest } from '@/api/auth'

interface UserState {
  user: User | null
  isHydrated: boolean
  setUser: (user: User) => void
  login: (phone: string, password: string) => Promise<boolean>
  hydrateUser: () => Promise<void>
  logout: () => void
}

export const useUserStore = create<UserState>((set) => ({
  user: null,
  isHydrated: false,

  setUser: (user) => set({ user }),

  login: async (phone, password) => {
    try {
      const result = await loginRequest({ phone, password })
      set({
        user: {
          id: result.user.id,
          nickname: result.user.nickname,
          avatar: result.user.avatar,
          isLoggedIn: true,
        },
        isHydrated: true,
      })
      return true
    } catch {
      clearSession()
      set({ user: null, isHydrated: true })
      return false
    }
  },

  hydrateUser: async () => {
    try {
      const currentUser = await getCurrentUser()
      set({
        user: {
          id: currentUser.id,
          nickname: currentUser.nickname,
          avatar: currentUser.avatar,
          isLoggedIn: true,
        },
        isHydrated: true,
      })
    } catch {
      clearSession()
      set({ user: null, isHydrated: true })
    }
  },

  logout: () => {
    clearSession()
    set({ user: null, isHydrated: true })
  },
}))
