import { apiClient, unwrapResponse } from './client'
import type { LiveRoom } from '@/types'

export const getRooms = async () => {
  const response = await apiClient.get('/rooms')
  return unwrapResponse<LiveRoom[]>(response)
}

