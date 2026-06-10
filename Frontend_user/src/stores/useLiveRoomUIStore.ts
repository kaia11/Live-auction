import { create } from 'zustand'
import type { AuctionItem } from '@/types'

export interface LiveRoomUIStateValues {
  isCurrentAuctionCardClosed: boolean
  showAuctionItemDrawer: boolean
  showBidPanel: boolean
  bidPanelMode: 'bid' | 'detail'
  showRuleModal: boolean
  showBidSuccessModal: boolean
  showOvertakenModal: boolean
  showAuctionEndPanel: boolean
  showDelayBanner: boolean
  endedAuctionItem: AuctionItem | null
}

interface LiveRoomUIState extends LiveRoomUIStateValues {
  setUIState: (patch: Partial<LiveRoomUIStateValues>) => void
  toggleAuctionItemDrawer: () => void
  toggleBidPanel: () => void
  openBidPanel: (mode?: 'bid' | 'detail') => void
  toggleRuleModal: () => void
  closeAllModals: () => void
  setCurrentAuctionCardClosed: (closed: boolean) => void
}

const initialState: LiveRoomUIStateValues = {
  isCurrentAuctionCardClosed: false,
  showAuctionItemDrawer: false,
  showBidPanel: false,
  bidPanelMode: 'bid',
  showRuleModal: false,
  showBidSuccessModal: false,
  showOvertakenModal: false,
  showAuctionEndPanel: false,
  showDelayBanner: false,
  endedAuctionItem: null,
}

export const useLiveRoomUIStore = create<LiveRoomUIState>((set) => ({
  ...initialState,

  setUIState: (patch) => set((state) => ({ ...state, ...patch })),

  toggleAuctionItemDrawer: () =>
    set((state) => ({ showAuctionItemDrawer: !state.showAuctionItemDrawer })),

  toggleBidPanel: () => set((state) => ({ showBidPanel: !state.showBidPanel })),

  openBidPanel: (mode = 'bid') => set({ showBidPanel: true, bidPanelMode: mode }),

  toggleRuleModal: () => set((state) => ({ showRuleModal: !state.showRuleModal })),

  closeAllModals: () =>
    set({
      showAuctionItemDrawer: false,
      showBidPanel: false,
      showRuleModal: false,
      showBidSuccessModal: false,
      showOvertakenModal: false,
      showAuctionEndPanel: false,
      showDelayBanner: false,
      endedAuctionItem: null,
    }),

  setCurrentAuctionCardClosed: (closed) => set({ isCurrentAuctionCardClosed: closed }),
}))

