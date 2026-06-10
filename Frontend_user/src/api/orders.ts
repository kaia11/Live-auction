import { apiClient, getAccessToken, unwrapResponse, USE_MOCK } from './client'
import type { UserOrder } from '@/types'

export const getMyOrders = async () => {
  if (USE_MOCK || !getAccessToken()) {
    return [] as UserOrder[]
  }

  const response = await apiClient.get('/users/me/orders')
  const orders = unwrapResponse<
    Array<{
      id?: string
      orderId?: string
      sessionId: string
      roomId: string
      itemId: string
      buyerUserId: string
      amount: number
      status: UserOrder['status']
      createTime: string
    }>
  >(response)

  return orders.map((order) => ({
    id: order.id ?? order.orderId ?? '',
    sessionId: order.sessionId,
    roomId: order.roomId,
    itemId: order.itemId,
    buyerUserId: order.buyerUserId,
    amount: order.amount,
    status: order.status,
    createTime: order.createTime,
  }))
}

export const payOrder = async (orderId: string) => {
  const response = await apiClient.post(`/users/me/orders/${orderId}/pay`)
  const order = unwrapResponse<{
    id: string
    sessionId: string
    roomId: string
    itemId: string
    buyerUserId: string
    amount: number
    status: UserOrder['status']
    createTime: string
  }>(response)

  return {
    id: order.id,
    sessionId: order.sessionId,
    roomId: order.roomId,
    itemId: order.itemId,
    buyerUserId: order.buyerUserId,
    amount: order.amount,
    status: order.status,
    createTime: order.createTime,
  } satisfies UserOrder
}
