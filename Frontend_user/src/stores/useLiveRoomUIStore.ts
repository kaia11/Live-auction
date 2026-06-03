import { create } from 'zustand'

export interface LiveRoomUIStateValues {
  isCurrentAuctionCardClosed: boolean
  showAuctionItemDrawer: boolean
  showBidPanel: boolean
  showRuleModal: boolean
  showBidSuccessModal: boolean
  showOvertakenModal: boolean
  showAuctionEndPanel: boolean
  showDelayBanner: boolean
}

interface LiveRoomUIState extends LiveRoomUIStateValues {
  setUIState: (patch: Partial<LiveRoomUIStateValues>) => void
  toggleAuctionItemDrawer: () => void
  toggleBidPanel: () => void
  toggleRuleModal: () => void
  closeAllModals: () => void
  setCurrentAuctionCardClosed: (closed: boolean) => void
}

const initialState: LiveRoomUIStateValues = {
  isCurrentAuctionCardClosed: false,
  showAuctionItemDrawer: false,
  showBidPanel: false,
  showRuleModal: false,
  showBidSuccessModal: false,
  showOvertakenModal: false,
  showAuctionEndPanel: false,
  showDelayBanner: false,
}

export const useLiveRoomUIStore = create<LiveRoomUIState>((set) => ({
  ...initialState,

  setUIState: (patch) => set((state) => ({ ...state, ...patch })),

  toggleAuctionItemDrawer: () =>
    set((state) => ({ showAuctionItemDrawer: !state.showAuctionItemDrawer })),

  toggleBidPanel: () => set((state) => ({ showBidPanel: !state.showBidPanel })),

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
    }),

  setCurrentAuctionCardClosed: (closed) => set({ isCurrentAuctionCardClosed: closed }),
}))

