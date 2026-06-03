import { create } from 'zustand'

interface BidDraftState {
  draftPrice: number | null
  lastBidPrice: number | null
  quickBidOptions: number[]
  autoBidEnabled: boolean
  setDraftPrice: (price: number | null) => void
  setLastBidPrice: (price: number | null) => void
  setQuickBidOptions: (options: number[]) => void
  setAutoBidEnabled: (enabled: boolean) => void
  resetDraft: () => void
}

export const useBidDraftStore = create<BidDraftState>((set) => ({
  draftPrice: null,
  lastBidPrice: null,
  quickBidOptions: [],
  autoBidEnabled: false,

  setDraftPrice: (price) => set({ draftPrice: price }),
  setLastBidPrice: (price) => set({ lastBidPrice: price }),
  setQuickBidOptions: (options) => set({ quickBidOptions: options }),
  setAutoBidEnabled: (enabled) => set({ autoBidEnabled: enabled }),
  resetDraft: () =>
    set({
      draftPrice: null,
      lastBidPrice: null,
      quickBidOptions: [],
      autoBidEnabled: false,
    }),
}))

