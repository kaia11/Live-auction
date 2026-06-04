import { useEffect, useRef } from 'react'
import { BackendCommentPayload, BackendRoomEvent } from '@/api/rooms'
import { USE_MOCK, apiClient, getAccessToken } from '@/api/client'
import { useLiveRoomStore } from '@/stores/useLiveRoomStore'
import { useLiveRuntimeStore } from '@/stores/useLiveRuntimeStore'
import { useLiveRoomUIStore } from '@/stores/useLiveRoomUIStore'

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

export const useLiveRoomRealtime = (roomId: string | undefined) => {
  const reconnectTimerRef = useRef<number | null>(null)

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

      if (
        runtimeStateBefore.myBidStatus.isLeading &&
        !runtimeStateAfter.myBidStatus.isLeading
      ) {
        uiStore.setUIState({ showOvertakenModal: true })
      }

      if (runtimeStateAfter.currentCountdown > runtimeStateBefore.currentCountdown) {
        uiStore.setUIState({ showDelayBanner: true })
        window.setTimeout(() => {
          useLiveRoomUIStore.getState().setUIState({ showDelayBanner: false })
        }, 2500)
      }

      if (event.event === 'auction_session_ended') {
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
          void syncAfterEvent(event)
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
