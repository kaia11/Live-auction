import type { BackendBidHistory, BackendBidResult, CreateBidPayload } from './bids'
import type { LoginPayload, LoginResponse, AuthUser } from './auth'
import type {
  BackendAuctionItem,
  BackendAuctionSession,
  BackendCommentPayload,
  BackendRoom,
  BackendRoomEvent,
} from './rooms'
import type { BackendMyStatus, BackendRankingEntry, BackendRankingResponse } from './sessions'

type MockBidRecord = {
  id: string
  sessionId: string
  roomId: string
  itemId: string
  userId: string
  bidPrice: number
  requestId: string
  createTime: string
}

type MockState = {
  users: Record<string, AuthUser>
  rooms: BackendRoom[]
  itemsByRoom: Record<string, BackendAuctionItem[]>
  sessionsByRoom: Record<string, BackendAuctionSession>
  bids: MockBidRecord[]
  bidHistoriesByUser: Record<string, BackendBidHistory[]>
  eventsByRoom: Record<string, BackendRoomEvent[]>
  processedRequests: Record<string, BackendBidResult>
  currentUserByToken: Record<string, string>
}

const now = Date.now()
const future = (minutes: number) => new Date(now + minutes * 60 * 1000).toISOString()

const mockState: MockState = {
  users: {
    'user-001': {
      id: 'user-001',
      nickname: '阿宁',
      avatar: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=200&q=80',
      phone: '13800138000',
      role: 'viewer',
    },
    'user-002': {
      id: 'user-002',
      nickname: '小满',
      avatar: 'https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=200&q=80',
      phone: '13900139000',
      role: 'viewer',
    },
    'anchor-001': {
      id: 'anchor-001',
      nickname: '主播小玉',
      avatar: 'https://images.unsplash.com/photo-1506794778202-cad84cf45f1d?w=200&q=80',
      phone: '13700137000',
      role: 'anchor',
    },
  },
  rooms: [
    {
      id: 'room-001',
      title: '古风首饰直播竞拍间',
      coverImage: '',
      videoUrl: 'https://www.w3schools.com/html/mov_bbb.mp4',
      status: 'live',
      anchorUserId: 'anchor-001',
      anchorName: '主播小玉',
      onlineCount: 1288,
      thumbnail: '',
      currentSessionId: 'session-001',
    },
  ],
  itemsByRoom: {
    'room-001': [
      {
        id: 'item-001',
        roomId: 'room-001',
        title: '和田玉吊坠',
        coverImage: '',
        description: '直播竞拍样例拍品，使用本地 mock 联调。',
        startPrice: 0,
        incrementStep: 5,
        ceilingPrice: 999,
        durationSeconds: 120,
        extensionSeconds: 30,
        extensionTriggerSeconds: 30,
        queueStatus: 'active',
      },
      {
        id: 'item-002',
        roomId: 'room-001',
        title: '鎏金花丝耳坠',
        coverImage: '',
        description: '待上场拍品。',
        startPrice: 0,
        incrementStep: 10,
        ceilingPrice: null,
        durationSeconds: 120,
        extensionSeconds: 30,
        extensionTriggerSeconds: 30,
        queueStatus: 'queued',
      },
    ],
  },
  sessionsByRoom: {
    'room-001': {
      id: 'session-001',
      roomId: 'room-001',
      itemId: 'item-001',
      status: 'bidding',
      currentPrice: 135,
      leaderUserId: 'user-002',
      endTime: future(2),
      participantCount: 2,
      incrementStep: 5,
      extensionSeconds: 30,
      extensionTriggerSeconds: 30,
      ceilingPrice: 999,
      supportsAutoProxy: true,
    },
  },
  bids: [
    {
      id: 'bid-001',
      sessionId: 'session-001',
      roomId: 'room-001',
      itemId: 'item-001',
      userId: 'user-001',
      bidPrice: 130,
      requestId: 'req-seed-001',
      createTime: new Date(now - 60_000).toISOString(),
    },
    {
      id: 'bid-002',
      sessionId: 'session-001',
      roomId: 'room-001',
      itemId: 'item-001',
      userId: 'user-002',
      bidPrice: 135,
      requestId: 'req-seed-002',
      createTime: new Date(now - 30_000).toISOString(),
    },
  ],
  bidHistoriesByUser: {
    'user-001': [
      {
        id: 'bid-001',
        itemId: 'item-001',
        itemTitle: '和田玉吊坠',
        itemImage: '',
        bidPrice: 130,
        result: 'pending',
        bidTime: new Date(now - 60_000).toISOString(),
      },
    ],
    'user-002': [
      {
        id: 'bid-002',
        itemId: 'item-001',
        itemTitle: '和田玉吊坠',
        itemImage: '',
        bidPrice: 135,
        result: 'pending',
        bidTime: new Date(now - 30_000).toISOString(),
      },
    ],
  },
  eventsByRoom: {
    'room-001': [
      {
        roomId: 'room-001',
        event: 'room_comment_received',
        payload: { userId: 'user-002', nickname: '小满', content: '这个起拍价很合适！' },
        version: 1,
        serverTime: new Date(now - 40_000).toISOString(),
      },
      {
        roomId: 'room-001',
        event: 'room_comment_received',
        payload: { userId: 'user-001', nickname: '阿宁', content: '我准备出价了。' },
        version: 2,
        serverTime: new Date(now - 20_000).toISOString(),
      },
    ],
  },
  processedRequests: {},
  currentUserByToken: {},
}

const mockDelay = async () => {
  await new Promise((resolve) => window.setTimeout(resolve, 120))
}

const createToken = (userId: string) => `mock-token:${userId}`

const getUserIdFromToken = (token: string | null) => {
  if (!token) {
    return null
  }
  if (token.startsWith('mock-token:')) {
    return token.slice('mock-token:'.length)
  }
  return null
}

const getCurrentSessionById = (sessionId: string) =>
  Object.values(mockState.sessionsByRoom).find((session) => session.id === sessionId)

const getRankingEntries = (sessionId: string): BackendRankingEntry[] => {
  const highestByUser = new Map<string, number>()
  for (const bid of mockState.bids) {
    if (bid.sessionId !== sessionId) {
      continue
    }
    const current = highestByUser.get(bid.userId) ?? 0
    if (bid.bidPrice > current) {
      highestByUser.set(bid.userId, bid.bidPrice)
    }
  }

  const entries = Array.from(highestByUser.entries()).map(([userId, highestBid]) => {
    const user = mockState.users[userId]
    return {
      userId,
      nickname: user?.nickname ?? userId,
      avatar: user?.avatar ?? '',
      rank: 0,
      highestBid,
    }
  })

  entries.sort((a, b) => b.highestBid - a.highestBid)
  entries.forEach((entry, index) => {
    entry.rank = index + 1
  })
  return entries
}

const appendEvent = (roomId: string, event: string, payload: unknown) => {
  const roomEvents = mockState.eventsByRoom[roomId] ?? []
  const version = roomEvents.length > 0 ? roomEvents[roomEvents.length - 1].version + 1 : 1
  const nextEvent: BackendRoomEvent = {
    roomId,
    event,
    payload,
    version,
    serverTime: new Date().toISOString(),
  }
  mockState.eventsByRoom[roomId] = [...roomEvents, nextEvent]
}

export const mockLogin = async (payload: LoginPayload): Promise<LoginResponse> => {
  await mockDelay()
  const user = mockState.users['user-001']
  const token = createToken(user.id)
  mockState.currentUserByToken[token] = user.id
  return {
    token,
    user: { ...user, phone: payload.phone },
  }
}

export const mockGetCurrentUser = async (token: string | null): Promise<AuthUser> => {
  await mockDelay()
  const userId = getUserIdFromToken(token)
  if (!userId) {
    throw new Error('unauthorized')
  }
  const user = mockState.users[userId]
  if (!user) {
    throw new Error('user not found')
  }
  return user
}

export const mockGetRooms = async (): Promise<BackendRoom[]> => {
  await mockDelay()
  return mockState.rooms
}

export const mockGetRoomItems = async (roomId: string): Promise<BackendAuctionItem[]> => {
  await mockDelay()
  return mockState.itemsByRoom[roomId] ?? []
}

export const mockGetCurrentSession = async (roomId: string): Promise<BackendAuctionSession> => {
  await mockDelay()
  const session = mockState.sessionsByRoom[roomId]
  if (!session) {
    throw new Error('session not found')
  }
  return session
}

export const mockGetRoomEvents = async (
  roomId: string,
  sinceVersion = 0,
): Promise<BackendRoomEvent[]> => {
  await mockDelay()
  return (mockState.eventsByRoom[roomId] ?? []).filter((event) => event.version > sinceVersion)
}

export const mockCreateRoomComment = async (
  roomId: string,
  content: string,
  token: string | null,
): Promise<BackendCommentPayload> => {
  await mockDelay()
  const userId = getUserIdFromToken(token)
  if (!userId) {
    throw new Error('unauthorized')
  }
  const user = mockState.users[userId]
  const comment = {
    userId,
    nickname: user.nickname,
    content,
  }
  appendEvent(roomId, 'room_comment_received', comment)
  return comment
}

export const mockGetSessionRanking = async (sessionId: string): Promise<BackendRankingResponse> => {
  await mockDelay()
  const entries = getRankingEntries(sessionId)
  return {
    sessionId,
    top3: entries.slice(0, 3),
    me: entries[0] ?? { userId: '', nickname: '', avatar: '', rank: 0, highestBid: 0 },
  }
}

export const mockGetMyBidStatus = async (
  sessionId: string,
  token: string | null,
): Promise<BackendMyStatus> => {
  await mockDelay()
  const userId = getUserIdFromToken(token)
  if (!userId) {
    throw new Error('unauthorized')
  }
  const session = getCurrentSessionById(sessionId)
  if (!session) {
    throw new Error('session not found')
  }
  const ranking = getRankingEntries(sessionId)
  const me = ranking.find((entry) => entry.userId === userId)
  return {
    sessionId,
    userId,
    myHighestBid: me?.highestBid ?? 0,
    myRank: me?.rank ?? 0,
    isLeading: me?.rank === 1,
    currentPrice: session.currentPrice,
    nextMinimumBid: session.currentPrice + session.incrementStep,
    vibrateSignalHint: me?.rank === 1 ? 'none' : me ? 'overtaken' : 'none',
  }
}

export const mockCreateBid = async (
  payload: CreateBidPayload,
): Promise<BackendBidResult> => {
  await mockDelay()

  const duplicated = mockState.processedRequests[payload.requestId]
  if (duplicated) {
    return duplicated
  }

  const session = mockState.sessionsByRoom[payload.roomId]
  if (!session || session.id !== payload.sessionId) {
    throw new Error('session not found')
  }

  const item = (mockState.itemsByRoom[payload.roomId] ?? []).find((entry) => entry.id === payload.itemId)
  if (!item) {
    throw new Error('item not found')
  }

  const nextMinimumBid = session.currentPrice + session.incrementStep
  if (payload.bidPrice < nextMinimumBid) {
    throw new Error('invalid bid price')
  }

  let acceptedBidPrice = payload.bidPrice
  let ceilingReached = false
  if (typeof session.ceilingPrice === 'number' && payload.bidPrice >= session.ceilingPrice) {
    acceptedBidPrice = session.ceilingPrice
    ceilingReached = true
  }

  const bidRecord: MockBidRecord = {
    id: `bid-${mockState.bids.length + 1}`,
    sessionId: payload.sessionId,
    roomId: payload.roomId,
    itemId: payload.itemId,
    userId: payload.userId,
    bidPrice: acceptedBidPrice,
    requestId: payload.requestId,
    createTime: new Date().toISOString(),
  }
  mockState.bids.push(bidRecord)

  session.currentPrice = acceptedBidPrice
  session.leaderUserId = payload.userId
  session.participantCount = new Set(
    mockState.bids.filter((bid) => bid.sessionId === session.id).map((bid) => bid.userId),
  ).size

  const history = {
    id: bidRecord.id,
    itemId: payload.itemId,
    itemTitle: item.title,
    itemImage: '',
    bidPrice: acceptedBidPrice,
    result: ceilingReached ? 'win' : 'pending',
    bidTime: bidRecord.createTime,
  } satisfies BackendBidHistory
  mockState.bidHistoriesByUser[payload.userId] = [
    history,
    ...(mockState.bidHistoriesByUser[payload.userId] ?? []),
  ]

  appendEvent(payload.roomId, 'auction_price_updated', {
    roomId: payload.roomId,
    sessionId: payload.sessionId,
    itemId: payload.itemId,
    userId: payload.userId,
    acceptedBidPrice,
    requestId: payload.requestId,
    currentPrice: acceptedBidPrice,
    isLeading: true,
    extensionApplied: false,
    ceilingReached,
    nextMinimumBid: acceptedBidPrice + session.incrementStep,
    vibrateSignalHint: 'overtake',
  })

  const result: BackendBidResult = {
    roomId: payload.roomId,
    sessionId: payload.sessionId,
    itemId: payload.itemId,
    userId: payload.userId,
    acceptedBidPrice,
    requestId: payload.requestId,
    currentPrice: acceptedBidPrice,
    isLeading: true,
    extensionApplied: false,
    ceilingReached,
    nextMinimumBid: acceptedBidPrice + session.incrementStep,
    vibrateSignalHint: 'overtake',
  }

  mockState.processedRequests[payload.requestId] = result
  return result
}

export const mockGetMyBidHistories = async (token: string | null): Promise<BackendBidHistory[]> => {
  await mockDelay()
  const userId = getUserIdFromToken(token)
  if (!userId) {
    throw new Error('unauthorized')
  }
  return mockState.bidHistoriesByUser[userId] ?? []
}
