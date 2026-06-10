export type RoomStatus = 'live' | 'offline'
export type QueueStatus = 'queued' | 'upcoming' | 'active' | 'finished' | 'cancelled'
export type SessionStatus = 'pending' | 'bidding' | 'ended_sold' | 'ended_passed' | 'cancelled'
export type OrderStatus = 'pending_payment' | 'paid' | 'shipped' | 'completed' | 'cancelled'
export type OrderAction = 'ship' | 'complete' | 'cancel'

export interface LiveRoom {
  id: string
  title: string
  coverImage: string
  videoUrl: string
  status: RoomStatus
  anchorUserId: string
  anchorName: string
  onlineCount: number
  thumbnail?: string
  currentSessionId: string
}

export interface AuctionItem {
  id: string
  roomId: string
  title: string
  coverImage: string
  description: string
  startPrice: number
  incrementStep: number
  ceilingPrice?: number | null
  durationSeconds: number
  extensionSeconds: number
  extensionTriggerSeconds: number
  queueStatus: QueueStatus
}

export interface AdminSession {
  roomId: string
  sessionId: string
  itemId: string
  status: SessionStatus
  currentPrice: number
  queueStatus: QueueStatus
  startTime?: string
  endTime: string
  participantCount?: number
  viewerCount?: number
  durationSeconds?: number
}

export interface GoodsRow {
  itemId: string
  sessionId: string
  roomId: string
  title: string
  coverImage: string
  description: string
  startPrice: number
  incrementStep: number
  ceilingPrice?: number | null
  durationSeconds: number
  extensionSeconds: number
  extensionTriggerSeconds: number
  currentPrice: number
  queueStatus: QueueStatus
  sessionStatus: SessionStatus
  endTime: string
  displayStatus: string
}

export interface CreateItemPayload {
  title: string
  coverImage: string
  description: string
  startPrice: number
  incrementStep: number
  ceilingPrice?: number | null
  durationSeconds: number
  extensionSeconds: number
  extensionTriggerSeconds: number
}

export interface UpdateItemPayload {
  title?: string
  coverImage?: string
  description?: string
  startPrice?: number
  incrementStep?: number
  ceilingPrice?: number | null
  durationSeconds?: number
  extensionSeconds?: number
  extensionTriggerSeconds?: number
}

export interface AdminOrder {
  orderId: string
  sessionId: string
  roomId: string
  itemId: string
  buyerUserId: string
  amount: number
  status: OrderStatus
  createTime: string
}

export interface DashboardOverview {
  totalRooms: number
  totalSessions: number
  soldSessions: number
  cancelledSessions: number
}

export interface DashboardTimelineEvent {
  time: string
  event: string
  sessionId: string
  itemId: string
  price: number
  userId: string
}

export interface SessionActionResponse {
  roomId: string
  sessionId?: string
  nextSessionId?: string
  itemId?: string
  nextItemId?: string
  status: SessionStatus
  queueStatus: QueueStatus
  endTime?: string
}

export interface SessionSettlementResponse {
  roomId: string
  sessionId: string
  itemId: string
  status: SessionStatus
  queueStatus: QueueStatus
  currentPrice: number
  winnerUserId: string
  nextSessionId?: string
  nextItemId?: string
}
