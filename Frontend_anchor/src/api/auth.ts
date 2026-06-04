import { apiClient, clearAccessToken, setAccessToken, unwrapResponse } from './client'

export interface AuthUser {
  id: string
  username: string
  nickname: string
  avatar: string
  role?: string
}

export interface LoginPayload {
  username: string
  password: string
}

export interface RegisterPayload {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: AuthUser
}

export const login = async (payload: LoginPayload) => {
  const response = await apiClient.post('/auth/login', {
    ...payload,
    clientType: 'anchor',
  })
  const result = unwrapResponse<LoginResponse>(response)
  setAccessToken(result.token)
  return result
}

export const register = async (payload: RegisterPayload) => {
  const response = await apiClient.post('/auth/register', {
    ...payload,
    clientType: 'anchor',
  })
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
