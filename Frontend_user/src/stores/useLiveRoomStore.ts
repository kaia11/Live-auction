import { create } from 'zustand'
import { LiveRoom, AuctionItem, RankingItem, MyBidStatus, BidHistory, LiveComment } from '@/types'
import { BackendCommentPayload, getCurrentSession, getRoomEvents, getRoomItems, getRooms, createRoomComment } from '@/api/rooms'
import { createBid, getMyBidHistories } from '@/api/bids'
import { getMyBidStatus, getSessionRanking } from '@/api/sessions'
import { mapAuctionRuntime, mapBackendRoom, mapBidHistories } from '@/adapters/auction'
import { useUserStore } from './useUserStore'

interface LiveRoomState {
  rooms: LiveRoom[]
  currentRoomId: string | null
  items: AuctionItem[]
  currentItemId: string | null
  currentSessionId: string | null
  lastEventVersion: number
  comments: LiveComment[]
  top3Ranking: RankingItem[]
  myBidStatus: MyBidStatus
  bidHistories: BidHistory[]
  onlineCount: number
  currentCountdown: number
  isCurrentAuctionCardClosed: boolean
  showAuctionItemDrawer: boolean
  showBidPanel: boolean
  showRuleModal: boolean
  showBidSuccessModal: boolean
  showOvertakenModal: boolean
  showAuctionEndPanel: boolean
  showDelayBanner: boolean
  setCurrentRoomId: (roomId: string) => void
  setCurrentItemId: (itemId: string) => void
  loadRooms: () => Promise<void>
  loadRoomRuntime: (roomId: string) => Promise<void>
  loadBidHistories: () => Promise<void>
  loadRoomComments: (roomId: string) => Promise<void>
  pollRoomEvents: (roomId: string) => Promise<void>
  toggleAuctionItemDrawer: () => void
  toggleBidPanel: () => void
  toggleRuleModal: () => void
  closeAllModals: () => void
  setCurrentAuctionCardClosed: (closed: boolean) => void
  submitBid: (price: number) => Promise<boolean>
  submitComment: (content: string) => Promise<boolean>
}

export const useLiveRoomStore = create<LiveRoomState>((set) => ({
  rooms: [],
  currentRoomId: null,
  items: [],
  currentItemId: null,
  currentSessionId: null,
  lastEventVersion: 0,
  comments: [],
  top3Ranking: [],
  myBidStatus: { myHighestPrice: 0, myRank: 0, isLeading: false },
  bidHistories: [],
  onlineCount: 0,
  currentCountdown: 0,
  isCurrentAuctionCardClosed: false,
  showAuctionItemDrawer: false,
  showBidPanel: false,
  showRuleModal: false,
  showBidSuccessModal: false,
  showOvertakenModal: false,
  showAuctionEndPanel: false,
  showDelayBanner: false,

  setCurrentRoomId: (roomId) => set({ currentRoomId: roomId }),

  setCurrentItemId: (itemId) => set({ currentItemId: itemId }),

  loadRooms: async () => {
    const rooms = await getRooms()
    set({
      rooms: rooms.map(mapBackendRoom),
    })
  },

  loadRoomRuntime: async (roomId) => {
    const userId = useUserStore.getState().user?.id
    const [items, session] = await Promise.all([
      getRoomItems(roomId),
      getCurrentSession(roomId),
    ])
    const [ranking, myStatus] = await Promise.all([
      getSessionRanking(session.id),
      getMyBidStatus(session.id, userId ?? 'user-001'),
    ])

    const currentRoom = getMappedRoomById(roomId, useLiveRoomStore.getState().rooms)
    const runtime = mapAuctionRuntime(items, session, ranking, myStatus)

    set({
      currentRoomId: roomId,
      items: runtime.items,
      currentItemId: runtime.currentItemId,
      currentSessionId: runtime.currentSessionId,
      lastEventVersion: useLiveRoomStore.getState().lastEventVersion,
      top3Ranking: runtime.top3Ranking,
      myBidStatus: runtime.myBidStatus,
      onlineCount: currentRoom?.onlineCount ?? session.participantCount,
      currentCountdown: runtime.currentCountdown,
    })
  },

  loadBidHistories: async () => {
    const histories = await getMyBidHistories()
    set({
      bidHistories: mapBidHistories(histories),
    })
  },

  loadRoomComments: async (roomId) => {
    const events = await getRoomEvents(roomId)
    set({
      comments: extractComments(events),
      lastEventVersion: events.length > 0 ? events[events.length - 1].version : 0,
    })
  },

  pollRoomEvents: async (roomId) => {
    const events = await getRoomEvents(roomId)
    const latestVersion = events.length > 0 ? events[events.length - 1].version : 0
    const { lastEventVersion, loadRoomRuntime, loadBidHistories } = useLiveRoomStore.getState()

    if (latestVersion <= lastEventVersion) {
      return
    }

    set({
      lastEventVersion: latestVersion,
      comments: extractComments(events),
    })

    await Promise.all([
      loadRoomRuntime(roomId),
      loadBidHistories(),
    ])
  },

  toggleAuctionItemDrawer: () => set(state => ({ showAuctionItemDrawer: !state.showAuctionItemDrawer })),

  toggleBidPanel: () => set(state => ({ showBidPanel: !state.showBidPanel })),

  toggleRuleModal: () => set(state => ({ showRuleModal: !state.showRuleModal })),

  closeAllModals: () => set({
    showAuctionItemDrawer: false,
    showBidPanel: false,
    showRuleModal: false,
    showBidSuccessModal: false,
    showOvertakenModal: false,
    showAuctionEndPanel: false,
    showDelayBanner: false,
  }),

  setCurrentAuctionCardClosed: (closed) => set({ isCurrentAuctionCardClosed: closed }),

  submitBid: async (price) => {
    const { currentRoomId, currentSessionId, currentItemId, loadRoomRuntime, loadBidHistories } = useLiveRoomStore.getState()
    const userId = useUserStore.getState().user?.id

    if (!currentRoomId || !currentSessionId || !currentItemId || !userId) {
      return false
    }

    const requestId = `req-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`

    await createBid({
      roomId: currentRoomId,
      sessionId: currentSessionId,
      itemId: currentItemId,
      userId,
      bidPrice: price,
      requestId,
    })

    await Promise.all([
      loadRoomRuntime(currentRoomId),
      loadBidHistories(),
    ])

    const events = await getRoomEvents(currentRoomId)
    const latestVersion = events.length > 0 ? events[events.length - 1].version : 0

    set(() => ({
      lastEventVersion: latestVersion,
      showBidSuccessModal: true,
    }))
    return true
  },

  submitComment: async (content) => {
    const roomId = useLiveRoomStore.getState().currentRoomId
    if (!roomId) {
      return false
    }

    await createRoomComment(roomId, content)
    await useLiveRoomStore.getState().loadRoomComments(roomId)
    return true
  },
}))

const getMappedRoomById = (roomId: string, rooms: LiveRoom[]) => rooms.find((room) => room.id === roomId)

const extractComments = (events: Awaited<ReturnType<typeof getRoomEvents>>): LiveComment[] =>
  events
    .filter((event) => event.event === 'room_comment_received')
    .map((event) => event.payload as BackendCommentPayload)
