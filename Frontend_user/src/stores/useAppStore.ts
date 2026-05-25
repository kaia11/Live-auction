import { create } from 'zustand'
import { AppState } from '@/types'

interface AppStore extends AppState {
  setLastVisitedRoomId: (roomId: string) => void
  setCurrentTab: (tab: 'live' | 'profile') => void
}

export const useAppStore = create<AppStore>((set) => ({
  lastVisitedRoomId: null,
  currentTab: 'live',

  setLastVisitedRoomId: (roomId) => set({ lastVisitedRoomId: roomId }),
  setCurrentTab: (tab) => set({ currentTab: tab }),
}))
