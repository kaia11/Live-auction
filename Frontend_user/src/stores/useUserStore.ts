import { create } from 'zustand'
import { User } from '@/types'

interface UserState {
  user: User | null
  setUser: (user: User) => void
  login: (phone: string, password: string) => Promise<boolean>
  logout: () => void
}

export const useUserStore = create<UserState>((set) => ({
  user: null,

  setUser: (user) => set({ user }),

  login: async (phone, password) => {
    await new Promise(resolve => setTimeout(resolve, 1000))
    const mockUser: User = {
      id: 'user-001',
      nickname: '珠宝收藏家',
      avatar: 'https://images.unsplash.com/photo-1662893404641-cf606970ae40?w=200',
      isLoggedIn: true,
    }
    set({ user: mockUser })
    return true
  },

  logout: () => {
    set({ user: null })
  },
}))
