import { apiClient, unwrapResponse } from './client'
import type {
  AdminOrder,
  AdminSession,
  AuctionItem,
  CreateItemPayload,
  DashboardOverview,
  DashboardTimelineEvent,
  OrderAction,
  SessionActionResponse,
  SessionSettlementResponse,
  UpdateItemPayload,
} from '@/types'

export const createItem = async (roomId: string, payload: CreateItemPayload) => {
  const response = await apiClient.post(`/admin/rooms/${roomId}/items`, payload)
  return unwrapResponse<AuctionItem>(response)
}

export const updateItem = async (itemId: string, payload: UpdateItemPayload) => {
  const response = await apiClient.patch(`/admin/items/${itemId}`, payload)
  return unwrapResponse<AuctionItem>(response)
}

export const getRoomItems = async (roomId: string) => {
  const response = await apiClient.get(`/rooms/${roomId}/items`)
  return unwrapResponse<AuctionItem[]>(response)
}

export const getRoomSessions = async (roomId: string) => {
  const response = await apiClient.get(`/admin/rooms/${roomId}/sessions`)
  return unwrapResponse<AdminSession[]>(response)
}

export const startSession = async (sessionId: string) => {
  const response = await apiClient.post(`/admin/sessions/${sessionId}/start`)
  return unwrapResponse<SessionActionResponse>(response)
}

export const cancelSession = async (sessionId: string) => {
  const response = await apiClient.post(`/admin/sessions/${sessionId}/cancel`)
  return unwrapResponse<SessionActionResponse>(response)
}

export const settleSession = async (sessionId: string) => {
  const response = await apiClient.post(`/admin/sessions/${sessionId}/settle`)
  return unwrapResponse<SessionSettlementResponse>(response)
}

export const activateNextItem = async (roomId: string) => {
  const response = await apiClient.post(`/admin/rooms/${roomId}/queue/next`)
  return unwrapResponse<SessionActionResponse>(response)
}

export const startRoomLive = async (roomId: string) => {
  const response = await apiClient.post(`/admin/rooms/${roomId}/start`)
  return unwrapResponse<{ roomId: string; status: 'live' | 'offline' }>(response)
}

export const stopRoomLive = async (roomId: string) => {
  const response = await apiClient.post(`/admin/rooms/${roomId}/stop`)
  return unwrapResponse<{ roomId: string; status: 'live' | 'offline' }>(response)
}

export const getAdminOrders = async () => {
  const response = await apiClient.get('/admin/orders')
  return unwrapResponse<AdminOrder[]>(response)
}

export const updateOrderStatus = async (orderId: string, action: OrderAction) => {
  const response = await apiClient.post(`/admin/orders/${orderId}/status`, { action })
  return unwrapResponse<AdminOrder>(response)
}

export const getOverview = async () => {
  const response = await apiClient.get('/admin/stats/overview')
  return unwrapResponse<DashboardOverview>(response)
}

export const getTimeline = async () => {
  const response = await apiClient.get('/admin/stats/timeline')
  return unwrapResponse<DashboardTimelineEvent[]>(response)
}
