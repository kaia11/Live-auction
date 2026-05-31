import { apiClient, getAccessToken, unwrapResponse, USE_MOCK } from './client'
import { BidHistory } from '@/types'
import { mockCreateBid, mockGetMyBidHistories } from './mock'

export interface CreateBidPayload {
  roomId: string
  sessionId: string
  itemId: string
  userId: string
  bidPrice: number
  requestId: string
}

export interface BackendBidResult {
  roomId: string
  sessionId: string
  itemId: string
  userId: string
  acceptedBidPrice: number
  requestId: string
  currentPrice: number
  isLeading: boolean
  extensionApplied: boolean
  ceilingReached: boolean
  nextMinimumBid: number
  vibrateSignalHint: string
}

export interface BackendBidHistory extends BidHistory {}

export const createBid = async (payload: CreateBidPayload) => {
  if (USE_MOCK) {
    return mockCreateBid(payload)
  }

  const response = await apiClient.post('/bids', payload)
  return unwrapResponse<BackendBidResult>(response)
}

export const getMyBidHistories = async () => {
  if (USE_MOCK) {
    return mockGetMyBidHistories(getAccessToken())
  }

  const response = await apiClient.get('/users/me/bids')
  return unwrapResponse<BackendBidHistory[]>(response)
}
