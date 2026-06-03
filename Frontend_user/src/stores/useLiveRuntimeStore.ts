import { create } from 'zustand'
import { RankingItem, MyBidStatus, LiveComment } from '@/types'

export interface LiveRuntimeSnapshot {
  currentRoomId: string | null
  currentItemId: string | null
  currentSessionId: string | null
  lastEventVersion: number
  comments: LiveComment[]
  top3Ranking: RankingItem[]
  myBidStatus: MyBidStatus
  onlineCount: number
  currentCountdown: number
}

export interface LiveRuntimeState extends LiveRuntimeSnapshot {
  setRuntimeState: (patch: Partial<LiveRuntimeSnapshot>) => void
  syncRuntimeSnapshot: (snapshot: Partial<LiveRuntimeSnapshot>) => void
  setCurrentRoomId: (roomId: string | null) => void
  setCurrentItemId: (itemId: string | null) => void
}

const initialRuntimeState: LiveRuntimeSnapshot = {
  currentRoomId: null,
  currentItemId: null,
  currentSessionId: null,
  lastEventVersion: 0,
  comments: [],
  top3Ranking: [],
  myBidStatus: { myHighestPrice: 0, myRank: 0, isLeading: false },
  onlineCount: 0,
  currentCountdown: 0,
}

export const useLiveRuntimeStore = create<LiveRuntimeState>((set) => ({
  ...initialRuntimeState,

  setRuntimeState: (patch) => set((state) => ({ ...state, ...patch })),

  syncRuntimeSnapshot: (snapshot) => set((state) => ({ ...state, ...snapshot })),

  setCurrentRoomId: (roomId) => set({ currentRoomId: roomId }),

  setCurrentItemId: (itemId) => set({ currentItemId: itemId }),
}))

