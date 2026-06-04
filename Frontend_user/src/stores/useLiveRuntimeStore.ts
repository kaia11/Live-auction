import { create } from 'zustand'
import { RankingItem, MyBidStatus, LiveComment } from '@/types'

export type LiveConnectionState = 'idle' | 'connecting' | 'connected' | 'reconnecting' | 'disconnected' | 'error'

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
  connectionState: LiveConnectionState
}

export interface LiveRuntimeState extends LiveRuntimeSnapshot {
  setRuntimeState: (patch: Partial<LiveRuntimeSnapshot>) => void
  syncRuntimeSnapshot: (snapshot: Partial<LiveRuntimeSnapshot>) => void
  setCurrentRoomId: (roomId: string | null) => void
  setCurrentItemId: (itemId: string | null) => void
  setConnectionState: (connectionState: LiveConnectionState) => void
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
  connectionState: 'idle',
}

export const useLiveRuntimeStore = create<LiveRuntimeState>((set) => ({
  ...initialRuntimeState,

  setRuntimeState: (patch) => set((state) => ({ ...state, ...patch })),

  syncRuntimeSnapshot: (snapshot) => set((state) => ({ ...state, ...snapshot })),

  setCurrentRoomId: (roomId) => set({ currentRoomId: roomId }),

  setCurrentItemId: (itemId) => set({ currentItemId: itemId }),

  setConnectionState: (connectionState) => set({ connectionState }),
}))
