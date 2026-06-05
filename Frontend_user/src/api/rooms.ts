import { apiClient, getAccessToken, unwrapResponse, USE_MOCK } from './client'
import {
  mockCreateRoomComment,
  mockGetCurrentSession,
  mockGetRoomEvents,
  mockGetRoomItems,
  mockGetRooms,
} from './mock'

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
  if (USE_MOCK) {
    return mockGetRooms()
  }

  const response = await apiClient.get('/rooms')
  return unwrapResponse<BackendRoom[]>(response)
}

export const getRoomDetail = async (roomId: string) => {
  if (USE_MOCK) {
    const rooms = await mockGetRooms()
    return rooms.find((room) => room.id === roomId) ?? null
  }

  const response = await apiClient.get(`/rooms/${roomId}`)
  return unwrapResponse<BackendRoom>(response)
}

export const getRoomItems = async (roomId: string) => {
  if (USE_MOCK) {
    return mockGetRoomItems(roomId)
  }

  const response = await apiClient.get(`/rooms/${roomId}/items`)
  return unwrapResponse<BackendAuctionItem[]>(response)
}

export const getCurrentSession = async (roomId: string) => {
  if (USE_MOCK) {
    return mockGetCurrentSession(roomId)
  }

  const response = await apiClient.get(`/rooms/${roomId}/current-session`)
  return unwrapResponse<BackendAuctionSession>(response)
}

export const getRoomEvents = async (roomId: string, sinceVersion = 0) => {
  if (USE_MOCK) {
    return mockGetRoomEvents(roomId, sinceVersion)
  }

  const response = await apiClient.get(`/rooms/${roomId}/events`)
  const result = unwrapResponse<{ roomId: string; sinceVersion: number; latestVersion: number; events: BackendRoomEvent[] }>(response)
  return result.events
}

export const createRoomComment = async (roomId: string, content: string) => {
  if (USE_MOCK) {
    return mockCreateRoomComment(roomId, content, getAccessToken())
  }

  const response = await apiClient.post(`/rooms/${roomId}/comments`, { content })
  return unwrapResponse<BackendCommentPayload>(response)
}
