export interface User {
  id: string
  nickname: string
  avatar: string
  role?: string
  isLoggedIn: boolean
}

export interface LiveRoom {
  id: string
  title: string
  anchorName: string
  coverImage: string
  status: 'living' | 'offline'
  onlineCount: number
  thumbnail?: string
}

export enum AuctionItemStatus {
  PENDING = '待上架',
  NOT_STARTED = '未开始',
  COMING_SOON = '即将开始',
  BIDDING = '竞拍中',
  SOLD = '已成交',
  PASSED = '已流拍',
  CANCELLED = '已取消',
  ENDED = '已结束'
}

export interface AuctionItem {
  id: string
  title: string
  description: string
  images: string[]
  depositAmount: number
  startPrice: number
  currentPrice: number
  minIncrement: number
  maxPrice: number | null
  status: AuctionItemStatus
  startTime?: string
  endTime?: string
  duration: number
  extendedSeconds: number
  extensionTriggerSeconds: number
  currentLeader?: string
}

export interface RankingItem {
  userId: string
  nickname: string
  avatar: string
  price: number
}

export interface MyBidStatus {
  myHighestPrice: number
  myRank: number
  isLeading: boolean
}

export interface BidHistory {
  id: string
  itemId: string
  itemTitle: string
  itemImage: string
  bidPrice: number
  result: 'win' | 'lose' | 'pending'
  bidTime: string
}

export interface LiveComment {
  userId: string
  nickname: string
  content: string
}

export interface AppState {
  lastVisitedRoomId: string | null
  currentTab: 'live' | 'profile'
}
