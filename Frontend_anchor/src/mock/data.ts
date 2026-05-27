import type { AuctionGoods, DashboardStats, OrderItem } from '@/types'

export const dashboardStats: DashboardStats = {
  inProgress: 3,
  sold: 19,
  unsold: 4,
  totalAmount: 286_500,
  onlineCount: 3_420,
  bidsToday: 1_286,
}

export const mockGoods: AuctionGoods[] = [
  {
    id: 'JWL-1001',
    name: '天然冰种翡翠圆牌吊坠',
    cover:
      'https://images.unsplash.com/photo-1611652022419-a9419f74343d?auto=format&fit=crop&w=600&q=80',
    intro: '18K 金镶嵌，附国检证书，适合日常佩戴与收藏。',
    startPrice: 0,
    increment: 20,
    ceilingPrice: 1680,
    durationSec: 900,
    currentPrice: 1240,
    bidCount: 43,
    status: '竞拍中',
    traffic: [102, 120, 160, 188, 201, 220, 245],
    priceTrack: [0, 200, 420, 660, 920, 1100, 1240],
  },
  {
    id: 'JWL-1002',
    name: '和田白玉平安无事牌',
    cover:
      'https://images.unsplash.com/photo-1515562141207-7a88fb7ce338?auto=format&fit=crop&w=600&q=80',
    intro: '细腻油润，配手工编绳，男女通戴。',
    startPrice: 100,
    increment: 10,
    ceilingPrice: 1000,
    durationSec: 1000,
    currentPrice: 960,
    bidCount: 30,
    status: '已成交',
    traffic: [90, 118, 130, 176, 188, 199, 210],
    priceTrack: [100, 210, 330, 500, 640, 820, 960],
  },
  {
    id: 'JWL-1003',
    name: '南红玛瑙福豆挂件',
    cover:
      'https://images.unsplash.com/photo-1605100804763-247f67b3557e?auto=format&fit=crop&w=600&q=80',
    intro: '颜色饱满，寓意连中三元，适合送礼。',
    startPrice: 0,
    increment: 15,
    durationSec: 720,
    currentPrice: 0,
    bidCount: 0,
    status: '待上架',
    traffic: [0, 0, 0, 0, 0, 0, 0],
    priceTrack: [0, 0, 0, 0, 0, 0, 0],
  },
]

export const mockOrders: OrderItem[] = [
  {
    id: 'ODR-20260527-001',
    goodsId: 'JWL-1002',
    goodsName: '和田白玉平安无事牌',
    buyerName: '玉友_林先生',
    amount: 960,
    status: '已发货',
    logisticsNo: 'SF239948888CN',
    createdAt: '2026-05-27 20:18:12',
  },
  {
    id: 'ODR-20260527-002',
    goodsId: 'JWL-1004',
    goodsName: '碧玉猫眼戒面',
    buyerName: '翠友_安安',
    amount: 1280,
    status: '待发货',
    createdAt: '2026-05-27 21:03:26',
  },
]
