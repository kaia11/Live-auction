import type { BidHistory } from '@/types'

export interface ProfileStats {
  participatedSessions: number
  successfulBids: number
  totalBids: number
}

const pickLatestHistoryPerItem = (histories: BidHistory[]) => {
  const byItem = new Map<string, BidHistory>()
  for (const history of histories) {
    const existing = byItem.get(history.itemId)
    if (!existing) {
      byItem.set(history.itemId, history)
      continue
    }
    if (history.bidPrice > existing.bidPrice) {
      byItem.set(history.itemId, history)
      continue
    }
    if (history.bidPrice === existing.bidPrice) {
      const nextTime = new Date(history.bidTime).getTime()
      const oldTime = new Date(existing.bidTime).getTime()
      if (Number.isFinite(nextTime) && Number.isFinite(oldTime) && nextTime > oldTime) {
        byItem.set(history.itemId, history)
      }
    }
  }
  return Array.from(byItem.values())
}

export const computeProfileStats = (histories: BidHistory[]): ProfileStats => {
  const latestByItem = pickLatestHistoryPerItem(histories)
  return {
    participatedSessions: latestByItem.length,
    successfulBids: latestByItem.filter((history) => history.result === 'win').length,
    totalBids: histories.length,
  }
}

export const getProfileBadge = (stats: ProfileStats) => {
  if (stats.successfulBids >= 3) {
    return 'VIP 收藏用户'
  }
  if (stats.successfulBids >= 1) {
    return '成交用户'
  }
  if (stats.totalBids > 0) {
    return '活跃竞拍用户'
  }
  return '新用户'
}
