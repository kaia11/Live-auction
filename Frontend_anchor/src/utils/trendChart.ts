import type { AdminSession, AuctionItem, DashboardTimelineEvent } from '@/types'

export interface TrendChartSeries {
  key: string
  label: string
  color: string
}

export interface TrafficBarPoint {
  itemId: string
  label: string
  viewerCount: number
}

const CHART_COLORS = ['#2563eb', '#db2777', '#ea580c', '#059669', '#7c3aed', '#0891b2']

const shortenLabel = (title: string) => {
  if (title.length <= 12) {
    return title
  }
  return title.slice(-12)
}

const parseTime = (time: string) => {
  const value = new Date(time).getTime()
  return Number.isFinite(value) ? value : null
}

export const formatRelativeSeconds = (seconds: number) => {
  if (seconds < 60) {
    return `${seconds}s`
  }
  const minutes = Math.floor(seconds / 60)
  const remain = seconds % 60
  return remain > 0 ? `${minutes}m${remain}s` : `${minutes}m`
}

const resolveSessionStartMs = (
  itemId: string,
  session: AdminSession | undefined,
  events: DashboardTimelineEvent[],
) => {
  const startFromSession = session?.startTime ? parseTime(session.startTime) : null
  if (startFromSession !== null) {
    return startFromSession
  }

  const firstBid = events.find((event) => event.itemId === itemId)
  if (firstBid) {
    return parseTime(firstBid.time)
  }

  if (session?.endTime && session.durationSeconds) {
    const endMs = parseTime(session.endTime)
    if (endMs !== null) {
      return endMs - session.durationSeconds * 1000
    }
  }

  return null
}

export const buildTrafficBarData = (
  items: AuctionItem[],
  sessions: AdminSession[],
): TrafficBarPoint[] => {
  const lastItems = items.slice(-6)

  return lastItems.map((item) => {
    const session = sessions.find((entry) => entry.itemId === item.id)
    return {
      itemId: item.id,
      label: shortenLabel(item.title),
      viewerCount: session?.viewerCount ?? 0,
    }
  })
}

export const buildPriceLineModel = (
  events: DashboardTimelineEvent[],
  items: AuctionItem[],
  sessions: AdminSession[],
) => {
  const lastItems = items.slice(-6)
  const targetItemIds = lastItems.map((item) => item.id)

  if (targetItemIds.length === 0) {
    return { data: [], series: [] as TrendChartSeries[] }
  }

  const targetItemIdSet = new Set(targetItemIds)
  const series: TrendChartSeries[] = lastItems.map((item, index) => ({
    key: item.id,
    label: shortenLabel(item.title),
    color: CHART_COLORS[index % CHART_COLORS.length],
  }))

  const filteredEvents = events
    .filter((event) => targetItemIdSet.has(event.itemId))
    .sort((left, right) => new Date(left.time).getTime() - new Date(right.time).getTime())

  if (filteredEvents.length === 0) {
    return { data: [], series: [] as TrendChartSeries[] }
  }

  const startMsByItem = Object.fromEntries(
    lastItems.map((item) => {
      const session = sessions.find((entry) => entry.itemId === item.id)
      return [item.id, resolveSessionStartMs(item.id, session, filteredEvents)]
    }),
  ) as Record<string, number | null>

  const pointsByItem = Object.fromEntries(
    series.map((entry) => [entry.key, [] as Array<{ relativeSeconds: number; price: number }>]),
  ) as Record<string, Array<{ relativeSeconds: number; price: number }>>

  for (const event of filteredEvents) {
    const startMs = startMsByItem[event.itemId]
    const eventMs = parseTime(event.time)
    if (startMs === null || eventMs === null) {
      continue
    }

    const relativeSeconds = Math.max(0, Math.round((eventMs - startMs) / 1000))
    pointsByItem[event.itemId].push({ relativeSeconds, price: event.price })
  }

  for (const itemId of Object.keys(pointsByItem)) {
    pointsByItem[itemId].sort((left, right) => left.relativeSeconds - right.relativeSeconds)
  }

  const relativeSecondSet = new Set<number>()
  for (const points of Object.values(pointsByItem)) {
    for (const point of points) {
      relativeSecondSet.add(point.relativeSeconds)
    }
  }

  const allRelativeSeconds = [...relativeSecondSet].sort((left, right) => left - right)

  const data = allRelativeSeconds.map((relativeSeconds) => {
    const row: Record<string, string | number> = {
      relativeSeconds,
      relativeLabel: formatRelativeSeconds(relativeSeconds),
    }

    for (const entry of series) {
      const points = pointsByItem[entry.key]
      let latestPrice: number | undefined
      for (const point of points) {
        if (point.relativeSeconds <= relativeSeconds) {
          latestPrice = point.price
        } else {
          break
        }
      }
      if (latestPrice !== undefined) {
        row[entry.key] = latestPrice
      }
    }

    return row
  })

  const activeSeries = series.filter((entry) => pointsByItem[entry.key].length > 0)

  return { data, series: activeSeries }
}
