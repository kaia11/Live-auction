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

  return {
    id: item.id,
    title: normalizeText(item.title),
    description: normalizeText(item.description),
    images: [getItemCoverImage(item.id)],
    startPrice: item.startPrice,
    currentPrice: isCurrentItem ? session.currentPrice : item.startPrice,
    minIncrement: item.incrementStep,
    maxPrice: item.ceilingPrice ?? null,
    status: resolveAuctionItemStatus(item, session, isCurrentItem),
    endTime: isCurrentItem ? session.endTime : undefined,
    duration: item.durationSeconds,
    extendedSeconds: item.extensionSeconds,
    extensionTriggerSeconds: item.extensionTriggerSeconds,
    currentLeader: isCurrentItem ? session.leaderUserId : undefined,
  }
}

const resolveAuctionItemStatus = (
  item: BackendAuctionItem,
  session: BackendAuctionSession,
  isCurrentItem: boolean,
): AuctionItemStatus => {
  if (isCurrentItem) {
    switch (session.status) {
      case 'bidding':
        return AuctionItemStatus.BIDDING
      case 'ended_sold':
        return AuctionItemStatus.SOLD
      case 'ended_passed':
        return AuctionItemStatus.PASSED
      case 'cancelled':
        return AuctionItemStatus.CANCELLED
      default:
        return AuctionItemStatus.NOT_STARTED
    }
  }

  switch (item.queueStatus) {
    case 'queued':
    case 'upcoming':
      return AuctionItemStatus.COMING_SOON
    case 'finished':
      return AuctionItemStatus.ENDED
    default:
      return AuctionItemStatus.PENDING
  }
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
