import axios from 'axios'
import { API_BASE_URL, USE_MOCK } from './config'

export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
  serverTime: string
}

const TOKEN_STORAGE_KEY = 'live-auction-token'
let serverTimeOffsetMs = 0

export { USE_MOCK }

export const getAccessToken = () => {
  if (typeof window === 'undefined') {
    return null
  }

  return window.localStorage.getItem(TOKEN_STORAGE_KEY)
}

export const setAccessToken = (token: string) => {
  if (typeof window === 'undefined') {
    return
  }

  window.localStorage.setItem(TOKEN_STORAGE_KEY, token)
}

export const clearAccessToken = () => {
  if (typeof window === 'undefined') {
    return
  }

  window.localStorage.removeItem(TOKEN_STORAGE_KEY)
}

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
})

apiClient.interceptors.request.use((config) => {
  const token = getAccessToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }

  return config
})

const syncServerTimeOffset = (serverTime: string) => {
  if (!serverTime) return
  const ts = new Date(serverTime).getTime()
  if (!Number.isFinite(ts)) return
  serverTimeOffsetMs = ts - Date.now()
}

export const getServerNowMs = () => Date.now() + serverTimeOffsetMs

export const unwrapResponse = <T>(response: { data: ApiEnvelope<T> }) => {
  syncServerTimeOffset(response.data.serverTime)
  return response.data.data
}
