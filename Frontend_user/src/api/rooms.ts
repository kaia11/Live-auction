import { apiClient, unwrapResponse } from './client'

export interface BackendRoom {
  id: string
  title: string
  coverImage: string
  videoUrl: string
  status: string
  anchorUserId: string
  anchorName?: string
  onlineCount?: number
  thumbnail?: string
  currentSessionId: string
}

export interface BackendAuctionItem {
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
  queueStatus: string
}

export interface BackendAuctionSession {
  id: string
  roomId: string
  itemId: string
  status: string
  currentPrice: number
  leaderUserId: string
  endTime: string
  participantCount: number
  incrementStep: number
  extensionSeconds: number
  extensionTriggerSeconds: number
  ceilingPrice?: number | null
  supportsAutoProxy: boolean
}

export interface BackendRoomEvent {
  roomId: string
  event: string
  payload: unknown
  version: number
  serverTime: string
}

export interface BackendCommentPayload {
  userId: string
  nickname: string
  content: string
}

export const getRooms = async () => {
  const response = await apiClient.get('/rooms')
  return unwrapResponse<BackendRoom[]>(response)
}

export const getRoomItems = async (roomId: string) => {
  const response = await apiClient.get(`/rooms/${roomId}/items`)
  return unwrapResponse<BackendAuctionItem[]>(response)
}

export const getCurrentSession = async (roomId: string) => {
  const response = await apiClient.get(`/rooms/${roomId}/current-session`)
  return unwrapResponse<BackendAuctionSession>(response)
}

export const getRoomEvents = async (roomId: string) => {
  const response = await apiClient.get(`/rooms/${roomId}/events`)
  return unwrapResponse<BackendRoomEvent[]>(response)
}

export const createRoomComment = async (roomId: string, content: string) => {
  const response = await apiClient.post(`/rooms/${roomId}/comments`, { content })
  return unwrapResponse<BackendCommentPayload>(response)
}
