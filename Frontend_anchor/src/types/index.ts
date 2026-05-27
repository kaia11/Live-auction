export type GoodsStatus = '待上架' | '即将开始' | '竞拍中' | '已成交' | '已流拍' | '已取消'

export interface AuctionGoods {
  id: string
  name: string
  cover: string
  intro: string
  startPrice: number
  increment: number
  ceilingPrice?: number
  durationSec: number
  currentPrice: number
  bidCount: number
  status: GoodsStatus
  traffic: number[]
  priceTrack: number[]
}

export interface OrderItem {
  id: string
  goodsId: string
  goodsName: string
  buyerName: string
  amount: number
  status: '待发货' | '已发货' | '已完成'
  logisticsNo?: string
  createdAt: string
}

export interface DashboardStats {
  inProgress: number
  sold: number
  unsold: number
  totalAmount: number
  onlineCount: number
  bidsToday: number
}
