import { apiClient, clearAccessToken, getAccessToken, setAccessToken, unwrapResponse, USE_MOCK } from './client'
import { mockGetCurrentUser, mockLogin, mockRegister } from './mock'

export interface LoginPayload {
  username: string
  password: string
}

export interface RegisterPayload {
  username: string
  password: string
}

export interface AuthUser {
  id: string
  username: string
  nickname: string
  avatar: string
  role?: string
}

export interface LoginResponse {
  token: string
  user: AuthUser
}

export const login = async (payload: LoginPayload) => {
  if (USE_MOCK) {
    const result = await mockLogin(payload)
    setAccessToken(result.token)
    return result
  }

  const response = await apiClient.post('/auth/login', {
    ...payload,
    clientType: 'viewer',
  })
  const result = unwrapResponse<LoginResponse>(response)
  setAccessToken(result.token)
  return result
}

export const register = async (payload: RegisterPayload) => {
  if (USE_MOCK) {
    const result = await mockRegister(payload)
    setAccessToken(result.token)
    return result
  }

  const response = await apiClient.post('/auth/register', {
    ...payload,
    clientType: 'viewer',
  })
  const result = unwrapResponse<LoginResponse>(response)
  setAccessToken(result.token)
  return result
}

export const getCurrentUser = async () => {
  if (USE_MOCK) {
    return mockGetCurrentUser(getAccessToken())
  }

  const response = await apiClient.get('/users/me')
  return unwrapResponse<AuthUser>(response)
}

export const clearSession = () => {
  clearAccessToken()
}
