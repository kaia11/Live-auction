import { apiClient, unwrapResponse } from './client'

export interface BackendRankingEntry {
  userId: string
  nickname: string
  avatar: string
  rank: number
  highestBid: number
}

export interface BackendRankingResponse {
  sessionId: string
  top3: BackendRankingEntry[]
  me: BackendRankingEntry
}

export interface BackendMyStatus {
  sessionId: string
  userId: string
  myHighestBid: number
  myRank: number
  isLeading: boolean
  currentPrice: number
  nextMinimumBid: number
  vibrateSignalHint: string
}

export const getSessionRanking = async (sessionId: string) => {
  const response = await apiClient.get(`/sessions/${sessionId}/ranking`)
  return unwrapResponse<BackendRankingResponse>(response)
}

export const getMyBidStatus = async (sessionId: string, userId: string) => {
  const response = await apiClient.get(`/sessions/${sessionId}/my-status`, {
    params: { userId },
  })
  return unwrapResponse<BackendMyStatus>(response)
}
