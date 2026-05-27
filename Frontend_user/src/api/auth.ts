import { apiClient, clearAccessToken, setAccessToken, unwrapResponse } from './client'

export interface LoginPayload {
  phone: string
  password: string
}

export interface AuthUser {
  id: string
  nickname: string
  avatar: string
  phone?: string
  role?: string
}

export interface LoginResponse {
  token: string
  user: AuthUser
}

export const login = async (payload: LoginPayload) => {
  const response = await apiClient.post('/auth/login', payload)
  const result = unwrapResponse<LoginResponse>(response)
  setAccessToken(result.token)
  return result
}

export const getCurrentUser = async () => {
  const response = await apiClient.get('/users/me')
  return unwrapResponse<AuthUser>(response)
}

export const clearSession = () => {
  clearAccessToken()
}
