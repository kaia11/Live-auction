import { create } from 'zustand'
import { User } from '@/types'
import { clearSession, getCurrentUser, login as loginRequest, register as registerRequest } from '@/api/auth'

const isViewerRole = (role?: string) => role === 'viewer'

interface UserState {
  user: User | null
  isHydrated: boolean
  setUser: (user: User) => void
  login: (username: string, password: string) => Promise<boolean>
  register: (username: string, password: string) => Promise<boolean>
  hydrateUser: () => Promise<void>
  logout: () => void
}

export const useUserStore = create<UserState>((set) => ({
  user: null,
  isHydrated: false,

  setUser: (user) => set({ user }),

  login: async (username, password) => {
    try {
      const result = await loginRequest({ username, password })
      if (!isViewerRole(result.user.role)) {
        clearSession()
        set({ user: null, isHydrated: true })
        return false
      }

      set({
        user: {
          id: result.user.id,
          nickname: result.user.nickname,
          avatar: result.user.avatar,
          role: result.user.role,
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

  register: async (username, password) => {
    try {
      const result = await registerRequest({ username, password })
      if (!isViewerRole(result.user.role)) {
        clearSession()
        set({ user: null, isHydrated: true })
        return false
      }

      set({
        user: {
          id: result.user.id,
          nickname: result.user.nickname,
          avatar: result.user.avatar,
          role: result.user.role,
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
      if (!isViewerRole(currentUser.role)) {
        clearSession()
        set({ user: null, isHydrated: true })
        return
      }

      set({
        user: {
          id: currentUser.id,
          nickname: currentUser.nickname,
          avatar: currentUser.avatar,
          role: currentUser.role,
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
