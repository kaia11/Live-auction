import { MyBidStatus, RankingItem, LiveRoom, AuctionItem, AuctionItemStatus, BidHistory } from '@/types'
import { BackendAuctionItem, BackendAuctionSession, BackendRoom } from '@/api/rooms'
import { BackendBidHistory } from '@/api/bids'
import { BackendMyStatus, BackendRankingResponse } from '@/api/sessions'
import { getHistoryImage, getItemCoverImage, getRoomCoverImage, getRoomThumbnailImage } from '@/assets/localImages'

export interface AuctionRuntimeViewModel {
  items: AuctionItem[]
  currentItemId: string
  currentSessionId: string
  currentCountdown: number
  top3Ranking: RankingItem[]
  myBidStatus: MyBidStatus
}

export const mapBackendRoom = (room: BackendRoom): LiveRoom => ({
  id: room.id,
  title: normalizeText(room.title),
  anchorName: normalizeText(room.anchorName ?? '直播间主播'),
  coverImage: getRoomCoverImage(room.id),
  status: room.status === 'live' ? 'living' : 'offline',
  onlineCount: room.onlineCount ?? 0,
  thumbnail: room.currentSessionId ? getRoomThumbnailImage(room.id) : undefined,
})

export const mapAuctionRuntime = (
  items: BackendAuctionItem[],
  session: BackendAuctionSession,
  ranking: BackendRankingResponse,
  myStatus: BackendMyStatus,
): AuctionRuntimeViewModel => ({
  items: items.map((item) => mapBackendItem(item, session)),
  currentItemId: session.itemId,
  currentSessionId: session.id,
  currentCountdown: getCountdownSeconds(session.endTime),
  top3Ranking: ranking.top3.map((entry) => ({
    userId: entry.userId,
    nickname: normalizeText(entry.nickname),
    avatar: entry.avatar,
    price: entry.highestBid,
  })),
  myBidStatus: {
    myHighestPrice: myStatus.myHighestBid,
    myRank: myStatus.myRank,
    isLeading: myStatus.isLeading,
  },
})

export const mapBidHistories = (histories: BackendBidHistory[]): BidHistory[] =>
  histories.map((history) => ({
    id: history.id,
    itemId: history.itemId,
    itemTitle: normalizeText(history.itemTitle),
    itemImage: getHistoryImage(history.itemId),
    bidPrice: history.bidPrice,
    result: history.result,
    bidTime: history.bidTime,
  }))

const mapBackendItem = (item: BackendAuctionItem, session: BackendAuctionSession): AuctionItem => {
  const isCurrentItem = item.id === session.itemId
  const { description, depositAmount } = extractDepositAmount(normalizeText(item.description))
  const normalizedSessionStatus = normalizeSessionStatus(session.status)
  const normalizedQueueStatus = normalizeQueueStatus(item.queueStatus)

  return {
    id: item.id,
    title: normalizeText(item.title),
    description,
    images: [getItemCoverImage(item.id)],
    depositAmount,
    startPrice: item.startPrice,
    currentPrice: isCurrentItem ? session.currentPrice : item.startPrice,
    minIncrement: item.incrementStep,
    maxPrice: item.ceilingPrice ?? null,
    status: resolveAuctionItemStatus(normalizedSessionStatus, normalizedQueueStatus, isCurrentItem),
    endTime: isCurrentItem ? session.endTime : undefined,
    duration: item.durationSeconds,
    extendedSeconds: item.extensionSeconds,
    extensionTriggerSeconds: item.extensionTriggerSeconds,
    currentLeader: isCurrentItem ? session.leaderUserId : undefined,
  }
}

const extractDepositAmount = (description: string) => {
  const marker = /#deposit=(\d+)#/i
  const matched = description.match(marker)
  if (!matched) {
    return { description, depositAmount: 0 }
  }

  const amount = Number.parseInt(matched[1], 10)
  const cleaned = description.replace(marker, '').trim()
  return {
    description: cleaned || description,
    depositAmount: Number.isFinite(amount) ? amount : 0,
  }
}

const resolveAuctionItemStatus = (
  sessionStatus: string,
  queueStatus: string,
  isCurrentItem: boolean,
): AuctionItemStatus => {
  if (isCurrentItem) {
    switch (sessionStatus) {
      case 'bidding':
        return AuctionItemStatus.BIDDING
      case 'ended_sold':
        return AuctionItemStatus.SOLD
      case 'ended_passed':
        return AuctionItemStatus.PASSED
      case 'cancelled':
        return AuctionItemStatus.CANCELLED
      case 'pending':
        if (queueStatus === 'active') {
          return AuctionItemStatus.BIDDING
        }
        if (queueStatus === 'cancelled') {
          return AuctionItemStatus.CANCELLED
        }
        if (queueStatus === 'finished') {
          return AuctionItemStatus.ENDED
        }
        return AuctionItemStatus.NOT_STARTED
      default:
        return queueStatus === 'finished' ? AuctionItemStatus.ENDED : AuctionItemStatus.NOT_STARTED
    }
  }

  switch (queueStatus) {
    case 'queued':
    case 'upcoming':
      return AuctionItemStatus.COMING_SOON
    case 'finished':
      return AuctionItemStatus.ENDED
    default:
      return AuctionItemStatus.PENDING
  }
}

const normalizeStatus = (value: string) => value.trim().toLowerCase()

const normalizeSessionStatus = (status: string) => {
  const value = normalizeStatus(status)
  if (!value) return 'pending'
  if (value.includes('ended_sold') || value === 'sold' || value.includes('已成交')) return 'ended_sold'
  if (value.includes('ended_passed') || value === 'passed' || value.includes('流拍')) return 'ended_passed'
  if (value.includes('cancel') || value.includes('已取消')) return 'cancelled'
  if (value.includes('bidding') || value === 'active' || value.includes('竞拍中')) return 'bidding'
  if (value.includes('pending') || value.includes('upcoming') || value.includes('queued') || value.includes('未开始') || value.includes('待上架') || value.includes('即将开始')) return 'pending'
  return value
}

const normalizeQueueStatus = (status: string) => {
  const value = normalizeStatus(status)
  if (!value) return 'queued'
  if (value.includes('cancel') || value.includes('已取消')) return 'cancelled'
  if (value.includes('finish') || value.includes('结束') || value.includes('已成交') || value.includes('流拍')) return 'finished'
  if (value.includes('active') || value.includes('bidding') || value.includes('竞拍中')) return 'active'
  if (value.includes('upcoming') || value.includes('即将开始')) return 'upcoming'
  if (value.includes('queued') || value.includes('待上架')) return 'queued'
  return value
}

const getCountdownSeconds = (endTime: string) => {
  if (!endTime) {
    return 0
  }

  const diffMs = new Date(endTime).getTime() - Date.now()
  return Math.max(0, Math.ceil(diffMs / 1000))
}

const normalizeText = (value: string) => {
  if (!value) {
    return value
  }

  // Some seeded cloud records may be stored as UTF-8 bytes interpreted as latin1.
  // If obvious mojibake markers appear, try to decode back into UTF-8.
  if (!/[\u4e00-\u9fa5]/.test(value) && /[ÃÂÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖ×ØÙÚÛÜÝÞßà-ÿ]/.test(value)) {
    try {
      return decodeURIComponent(escape(value))
    } catch {
      return value
    }
  }
  return value
}
