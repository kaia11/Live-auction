import { create } from 'zustand'
import { LiveRoom, AuctionItem, BidHistory, LiveComment } from '@/types'
import { BackendCommentPayload, createRoomComment, getCurrentSession, getRoomDetail, getRoomEvents, getRoomItems, getRooms } from '@/api/rooms'
import { configureAutoProxy, createBid, getMyBidHistories } from '@/api/bids'
import { getMyBidStatus, getSessionRanking } from '@/api/sessions'
import { mapAuctionRuntime, mapBackendRoom, mapBidHistories } from '@/adapters/auction'
import { useUserStore } from './useUserStore'
import { useLiveRoomUIStore, LiveRoomUIStateValues } from './useLiveRoomUIStore'
import { useLiveRuntimeStore, LiveRuntimeSnapshot } from './useLiveRuntimeStore'

interface LiveRoomState extends LiveRoomUIStateValues, LiveRuntimeSnapshot {
  rooms: LiveRoom[]
  items: AuctionItem[]
  bidHistories: BidHistory[]
  autoProxyFallbackMode: boolean
  autoProxyLocalEnabled: boolean
  autoProxyLocalMaxPrice: number
  setCurrentRoomId: (roomId: string) => void
  setCurrentItemId: (itemId: string) => void
  syncRooms: (rooms: LiveRoom[]) => void
  syncRuntimeSnapshot: (snapshot: Partial<LiveRuntimeSnapshot> & { items?: AuctionItem[] }) => void
  syncBidHistories: (bidHistories: BidHistory[]) => void
  syncCommentsSnapshot: (comments: LiveComment[], lastEventVersion: number) => void
  loadRooms: () => Promise<void>
  loadRoomRuntime: (roomId: string) => Promise<void>
  loadBidHistories: () => Promise<void>
  loadRoomComments: (roomId: string) => Promise<void>
  pollRoomEvents: (roomId: string) => Promise<void>
  toggleAuctionItemDrawer: () => void
  toggleBidPanel: () => void
  openBidPanel: (mode?: 'bid' | 'detail') => void
  toggleRuleModal: () => void
  closeAllModals: () => void
  setCurrentAuctionCardClosed: (closed: boolean) => void
  submitBid: (price: number) => Promise<'leading' | 'overtaken' | false>
  configureAutoProxy: (maxPrice: number, enabled: boolean) => Promise<boolean>
  submitComment: (content: string) => Promise<boolean>
}

const initialUIState: LiveRoomUIStateValues = {
  isCurrentAuctionCardClosed: false,
  showAuctionItemDrawer: false,
  showBidPanel: false,
  bidPanelMode: 'bid',
  showRuleModal: false,
  showBidSuccessModal: false,
  showOvertakenModal: false,
  showAuctionEndPanel: false,
  showDelayBanner: false,
}

const initialRuntimeState: LiveRuntimeSnapshot = {
  currentRoomId: null,
  currentItemId: null,
  currentSessionId: null,
  lastEventVersion: 0,
  comments: [],
  top3Ranking: [],
  myBidStatus: { myHighestPrice: 0, myRank: 0, isLeading: false, autoProxyEnabled: false, autoProxyMaxPrice: 0 },
  onlineCount: 0,
  currentCountdown: 0,
  connectionState: 'idle',
}

const syncUIState = (patch: Partial<LiveRoomUIStateValues>) => {
  useLiveRoomUIStore.getState().setUIState(patch)
}

const syncRuntimeState = (patch: Partial<LiveRuntimeSnapshot>) => {
  useLiveRuntimeStore.getState().syncRuntimeSnapshot(patch)
}

export const useLiveRoomStore = create<LiveRoomState>((set) => ({
  rooms: [],
  items: [],
  bidHistories: [],
  autoProxyFallbackMode: false,
  autoProxyLocalEnabled: false,
  autoProxyLocalMaxPrice: 0,
  ...initialUIState,
  ...initialRuntimeState,

  setCurrentRoomId: (roomId) => {
    useLiveRuntimeStore.getState().setCurrentRoomId(roomId)
    set({ currentRoomId: roomId })
  },

  setCurrentItemId: (itemId) => {
    useLiveRuntimeStore.getState().setCurrentItemId(itemId)
    set({ currentItemId: itemId })
  },

  syncRooms: (rooms) => set({ rooms }),

  syncRuntimeSnapshot: (snapshot) => {
    const runtimePatch: Partial<LiveRuntimeSnapshot> = {}

    if (snapshot.currentRoomId !== undefined) runtimePatch.currentRoomId = snapshot.currentRoomId
    if (snapshot.currentItemId !== undefined) runtimePatch.currentItemId = snapshot.currentItemId
    if (snapshot.currentSessionId !== undefined) runtimePatch.currentSessionId = snapshot.currentSessionId
    if (snapshot.lastEventVersion !== undefined) runtimePatch.lastEventVersion = snapshot.lastEventVersion
    if (snapshot.comments !== undefined) runtimePatch.comments = snapshot.comments
    if (snapshot.top3Ranking !== undefined) runtimePatch.top3Ranking = snapshot.top3Ranking
    if (snapshot.myBidStatus !== undefined) runtimePatch.myBidStatus = snapshot.myBidStatus
    if (snapshot.onlineCount !== undefined) runtimePatch.onlineCount = snapshot.onlineCount
    if (snapshot.currentCountdown !== undefined) runtimePatch.currentCountdown = snapshot.currentCountdown

    syncRuntimeState(runtimePatch)

    set((state) => ({
      ...state,
      ...runtimePatch,
      items: snapshot.items ?? state.items,
    }))
  },

  syncBidHistories: (bidHistories) => set({ bidHistories }),

  syncCommentsSnapshot: (comments, lastEventVersion) => {
    syncRuntimeState({ comments, lastEventVersion })
    set({ comments, lastEventVersion })
  },

  loadRooms: async () => {
    const rooms = (await getRooms()).map(mapBackendRoom)
    set({ rooms })
  },

  loadRoomRuntime: async (roomId) => {
    const userId = useUserStore.getState().user?.id
    const [room, items, session] = await Promise.all([
      getRoomDetail(roomId),
      getRoomItems(roomId),
      getCurrentSession(roomId),
    ])
    const [ranking, myStatus] = await Promise.all([
      getSessionRanking(session.id),
      getMyBidStatus(session.id, userId ?? 'user-001'),
    ])

    const runtime = mapAuctionRuntime(items, session, ranking, myStatus)
    const runtimePatch = {
      currentRoomId: roomId,
      currentItemId: runtime.currentItemId,
      currentSessionId: runtime.currentSessionId,
      top3Ranking: runtime.top3Ranking,
      myBidStatus: runtime.myBidStatus,
      onlineCount: room?.onlineCount ?? 0,
      currentCountdown: runtime.currentCountdown,
    }

    syncRuntimeState(runtimePatch)
    set((state) => ({
      ...state,
      items: runtime.items,
      ...runtimePatch,
    }))
  },

  loadBidHistories: async () => {
    const bidHistories = mapBidHistories(await getMyBidHistories())
    set({ bidHistories })
  },

  loadRoomComments: async (roomId) => {
    const events = await getRoomEvents(roomId)
    const comments = extractComments(events)
    const lastEventVersion = events.length > 0 ? events[events.length - 1].version : 0

    syncRuntimeState({ comments, lastEventVersion })
    set({ comments, lastEventVersion })
  },

  pollRoomEvents: async (roomId) => {
    const { lastEventVersion, loadRoomRuntime, loadBidHistories } = useLiveRoomStore.getState()
    const events = await getRoomEvents(roomId, lastEventVersion)
    const latestVersion = events.length > 0 ? events[events.length - 1].version : lastEventVersion

    if (latestVersion <= lastEventVersion) {
      return
    }

    const comments = mergeComments(useLiveRoomStore.getState().comments, extractComments(events))
    syncRuntimeState({
      comments,
      lastEventVersion: latestVersion,
    })

    set({
      lastEventVersion: latestVersion,
      comments,
    })

    await Promise.all([
      loadRoomRuntime(roomId),
      loadBidHistories(),
    ])
  },

  toggleAuctionItemDrawer: () =>
    set((state) => {
      const nextValue = !state.showAuctionItemDrawer
      syncUIState({ showAuctionItemDrawer: nextValue })
      return { showAuctionItemDrawer: nextValue }
    }),

  toggleBidPanel: () =>
    set((state) => {
      const nextValue = !state.showBidPanel
      syncUIState({ showBidPanel: nextValue })
      return { showBidPanel: nextValue }
    }),

  openBidPanel: (mode = 'bid') => {
    syncUIState({ showBidPanel: true, bidPanelMode: mode })
    set({ showBidPanel: true, bidPanelMode: mode })
  },

  toggleRuleModal: () =>
    set((state) => {
      const nextValue = !state.showRuleModal
      syncUIState({ showRuleModal: nextValue })
      return { showRuleModal: nextValue }
    }),

  closeAllModals: () => {
    syncUIState({
      showAuctionItemDrawer: false,
      showBidPanel: false,
      showRuleModal: false,
      showBidSuccessModal: false,
      showOvertakenModal: false,
      showAuctionEndPanel: false,
      showDelayBanner: false,
    })
    set({
      showAuctionItemDrawer: false,
      showBidPanel: false,
      showRuleModal: false,
      showBidSuccessModal: false,
      showOvertakenModal: false,
      showAuctionEndPanel: false,
      showDelayBanner: false,
    })
  },

  setCurrentAuctionCardClosed: (closed) => {
    syncUIState({ isCurrentAuctionCardClosed: closed })
    set({ isCurrentAuctionCardClosed: closed })
  },

  submitBid: async (price) => {
    const { currentRoomId, currentSessionId, currentItemId, loadRoomRuntime, loadBidHistories } = useLiveRoomStore.getState()
    const userId = useUserStore.getState().user?.id

    if (!currentRoomId || !currentSessionId || !currentItemId || !userId) {
      return false
    }

    const requestId = `req-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`

    const result = await createBid({
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

    syncRuntimeState({ lastEventVersion: latestVersion })
    set({ lastEventVersion: latestVersion })

    if (result.isLeading) {
      syncUIState({ showBidSuccessModal: true, showOvertakenModal: false })
      set({ showBidSuccessModal: true, showOvertakenModal: false })
      return 'leading'
    }

    syncUIState({ showBidSuccessModal: false, showOvertakenModal: true })
    set({ showBidSuccessModal: false, showOvertakenModal: true })
    return 'overtaken'
  },

  configureAutoProxy: async (maxPrice, enabled) => {
    const { currentRoomId, currentSessionId, currentItemId, loadRoomRuntime } = useLiveRoomStore.getState()
    if (!currentRoomId || !currentSessionId || !currentItemId) {
      return false
    }

    try {
      await configureAutoProxy({
        sessionId: currentSessionId,
        roomId: currentRoomId,
        itemId: currentItemId,
        maxPrice,
        enabled,
      })

      await loadRoomRuntime(currentRoomId)
      set({
        autoProxyFallbackMode: false,
        autoProxyLocalEnabled: false,
        autoProxyLocalMaxPrice: 0,
      })
      if (enabled) {
        await executeAutoProxyBid()
      }
      return true
    } catch (error: unknown) {
      // Cloud backend might not have the new auto-proxy endpoint yet.
      const status = (error as { response?: { status?: number } })?.response?.status
      if (status === 404 || status === 405) {
        set({
          autoProxyFallbackMode: true,
          autoProxyLocalEnabled: enabled,
          autoProxyLocalMaxPrice: enabled ? maxPrice : 0,
        })
        if (enabled) {
          await executeAutoProxyBid()
        }
        return true
      }
      throw error
    }
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

const extractComments = (events: Awaited<ReturnType<typeof getRoomEvents>>): LiveComment[] =>
  events
    .filter((event) => event.event === 'room_comment_received')
    .map((event) => event.payload as BackendCommentPayload)

const mergeComments = (prev: LiveComment[], next: LiveComment[]) => {
  if (next.length === 0) {
    return prev
  }

  const seen = new Set(prev.map((comment) => `${comment.userId}:${comment.content}`))
  const merged = [...prev]
  for (const comment of next) {
    const key = `${comment.userId}:${comment.content}`
    if (seen.has(key)) {
      continue
    }
    seen.add(key)
    merged.push(comment)
  }
  return merged
}

const executeAutoProxyBid = async () => {
  const state = useLiveRoomStore.getState()
  const userId = useUserStore.getState().user?.id
  const roomId = state.currentRoomId
  const sessionId = state.currentSessionId
  const itemId = state.currentItemId
  if (!roomId || !sessionId || !itemId || !userId) {
    return false
  }

  await state.loadRoomRuntime(roomId)
  const latest = useLiveRoomStore.getState()
  const item = latest.items.find((entry) => entry.id === latest.currentItemId)
  if (!item || item.status !== '竞拍中') {
    return false
  }
  if (item.currentLeader === userId || latest.myBidStatus.isLeading) {
    return false
  }

  const maxPrice = latest.autoProxyFallbackMode
    ? latest.autoProxyLocalMaxPrice
    : (latest.myBidStatus.autoProxyMaxPrice ?? 0)
  if (maxPrice <= 0) {
    return false
  }

  const nextBid = item.currentPrice + (item.minIncrement || 1)
  if (nextBid > maxPrice) {
    return false
  }

  try {
    await createBid({
      roomId,
      sessionId,
      itemId,
      userId,
      bidPrice: nextBid,
      requestId: `proxy-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    })
    await latest.loadRoomRuntime(roomId)
    return true
  } catch {
    return false
  }
}
