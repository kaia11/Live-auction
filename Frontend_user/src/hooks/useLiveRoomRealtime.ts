import { useEffect, useRef } from 'react'
import { BackendCommentPayload, BackendRoomEvent } from '@/api/rooms'
import { BackendBidResult } from '@/api/bids'
import { USE_MOCK, apiClient, getAccessToken } from '@/api/client'
import { useLiveRoomStore } from '@/stores/useLiveRoomStore'
import { useLiveRuntimeStore } from '@/stores/useLiveRuntimeStore'
import { useLiveRoomUIStore } from '@/stores/useLiveRoomUIStore'
import { clearDepositPaid } from '@/utils/deposit'
import { useUserStore } from '@/stores/useUserStore'

const getWebSocketUrl = (roomId: string) => {
  const baseUrl = new URL(apiClient.defaults.baseURL ?? window.location.origin)
  const protocol = baseUrl.protocol === 'https:' ? 'wss:' : 'ws:'
  const token = getAccessToken()
  const url = new URL(`${protocol}//${baseUrl.host}/ws`)
  url.searchParams.set('roomId', roomId)
  if (token) {
    url.searchParams.set('token', token)
  }
  return url.toString()
}

const isReconnectState = (state: ReturnType<typeof useLiveRuntimeStore.getState>['connectionState']) =>
  state === 'connected' || state === 'reconnecting'

const SESSION_TRANSITION_EVENTS = new Set([
  'auction_session_ended',
  'auction_session_upcoming',
  'auction_session_activated',
  'room_item_queue_updated',
])

export const useLiveRoomRealtime = (roomId: string | undefined) => {
  const reconnectTimerRef = useRef<number | null>(null)
  const syncQueueRef = useRef(Promise.resolve())

  useEffect(() => {
    if (!roomId || USE_MOCK || typeof window === 'undefined') {
      useLiveRuntimeStore.getState().setConnectionState(USE_MOCK ? 'idle' : 'disconnected')
      return
    }

    let isUnmounted = false
    let socket: WebSocket | null = null

    const clearReconnectTimer = () => {
      if (reconnectTimerRef.current !== null) {
        window.clearTimeout(reconnectTimerRef.current)
        reconnectTimerRef.current = null
      }
    }

    const syncAfterEvent = async (event: BackendRoomEvent) => {
      const liveRoomStore = useLiveRoomStore.getState()
      const runtimeStateBefore = useLiveRuntimeStore.getState()

      if (event.event === 'room_comment_received') {
        const payload = event.payload as BackendCommentPayload
        const nextComments = [
          ...runtimeStateBefore.comments,
          payload,
        ]
        liveRoomStore.syncCommentsSnapshot(nextComments, event.version)
        return
      }

      await Promise.all([
        liveRoomStore.loadRoomRuntime(roomId),
        liveRoomStore.loadBidHistories(),
      ])

      liveRoomStore.syncRuntimeSnapshot({
        lastEventVersion: event.version,
      })

      const runtimeStateAfter = useLiveRuntimeStore.getState()
      const uiStore = useLiveRoomUIStore.getState()
      const itemsAfter = useLiveRoomStore.getState().items
      const currentItemAfter = itemsAfter.find((item) => item.id === runtimeStateAfter.currentItemId)

      const auctionContextChanged =
        runtimeStateBefore.currentSessionId !== runtimeStateAfter.currentSessionId ||
        runtimeStateBefore.currentItemId !== runtimeStateAfter.currentItemId

      const isSameAuctionContext =
        !auctionContextChanged &&
        runtimeStateBefore.currentSessionId !== null &&
        runtimeStateBefore.currentItemId !== null

      const isActiveBidding = currentItemAfter?.status === '竞拍中'

      const shouldShowOvertakenModal =
        event.event === 'auction_price_updated' &&
        isSameAuctionContext &&
        isActiveBidding &&
        runtimeStateBefore.myBidStatus.isLeading &&
        !runtimeStateAfter.myBidStatus.isLeading &&
        runtimeStateBefore.myBidStatus.myHighestPrice > 0 &&
        runtimeStateAfter.myBidStatus.myHighestPrice > 0

      if (auctionContextChanged || SESSION_TRANSITION_EVENTS.has(event.event)) {
        uiStore.setUIState({ showOvertakenModal: false, showDelayBanner: false })
      }

      if (shouldShowOvertakenModal) {
        uiStore.setUIState({ showOvertakenModal: true })
      }

      const bidPayload =
        event.event === 'auction_price_updated'
          ? (event.payload as Partial<BackendBidResult>)
          : null
      const countdownDelta = runtimeStateAfter.currentCountdown - runtimeStateBefore.currentCountdown
      const extensionSeconds = currentItemAfter?.extendedSeconds ?? 0
      const shouldShowDelayBanner =
        event.event === 'auction_price_updated' &&
        isSameAuctionContext &&
        isActiveBidding &&
        (bidPayload?.extensionApplied === true ||
          (bidPayload?.extensionApplied === undefined &&
            countdownDelta > 0 &&
            extensionSeconds > 0 &&
            countdownDelta <= extensionSeconds + 2))

      if (shouldShowDelayBanner) {
        uiStore.setUIState({ showDelayBanner: true })
        window.setTimeout(() => {
          useLiveRoomUIStore.getState().setUIState({ showDelayBanner: false })
        }, 2500)
      }

      if (event.event === 'auction_session_ended') {
        const userId = useUserStore.getState().user?.id
        const endedItemId = runtimeStateBefore.currentItemId
        if (userId && endedItemId) {
          clearDepositPaid(userId, endedItemId)
        }
        uiStore.setUIState({ showAuctionEndPanel: true })
      }
    }

    const connect = () => {
      clearReconnectTimer()
      const currentConnectionState = useLiveRuntimeStore.getState().connectionState
      useLiveRuntimeStore
        .getState()
        .setConnectionState(isReconnectState(currentConnectionState) ? 'reconnecting' : 'connecting')

      socket = new WebSocket(getWebSocketUrl(roomId))

      socket.onopen = () => {
        useLiveRuntimeStore.getState().setConnectionState('connected')
      }

      socket.onmessage = (messageEvent) => {
        try {
          const event = JSON.parse(messageEvent.data) as BackendRoomEvent
          syncQueueRef.current = syncQueueRef.current
            .then(() => syncAfterEvent(event))
            .catch(() => {
              useLiveRuntimeStore.getState().setConnectionState('error')
            })
        } catch {
          useLiveRuntimeStore.getState().setConnectionState('error')
        }
      }

      socket.onerror = () => {
        useLiveRuntimeStore.getState().setConnectionState('error')
      }

      socket.onclose = () => {
        if (isUnmounted) {
          useLiveRuntimeStore.getState().setConnectionState('disconnected')
          return
        }

        useLiveRuntimeStore.getState().setConnectionState('reconnecting')
        reconnectTimerRef.current = window.setTimeout(() => {
          connect()
        }, 2500)
      }
    }

    connect()

    return () => {
      isUnmounted = true
      clearReconnectTimer()
      useLiveRuntimeStore.getState().setConnectionState('disconnected')
      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.close()
      }
    }
  }, [roomId])
}
