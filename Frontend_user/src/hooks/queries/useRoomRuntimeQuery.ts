import { useQuery } from '@tanstack/react-query'
import { getCurrentSession, getRoomDetail, getRoomItems } from '@/api/rooms'
import { getMyBidStatus, getSessionRanking } from '@/api/sessions'
import { mapAuctionRuntime } from '@/adapters/auction'
import { useUserStore } from '@/stores/useUserStore'

export const roomRuntimeQueryKey = (roomId: string | undefined) => ['room-runtime', roomId]

export const useRoomRuntimeQuery = (roomId: string | undefined) => {
  const userId = useUserStore((state) => state.user?.id)

  return useQuery({
    queryKey: roomRuntimeQueryKey(roomId),
    enabled: Boolean(roomId),
    queryFn: async () => {
      const safeRoomId = roomId as string
      const [room, items, session] = await Promise.all([
        getRoomDetail(safeRoomId),
        getRoomItems(safeRoomId),
        getCurrentSession(safeRoomId),
      ])
      const [ranking, myStatus] = await Promise.all([
        getSessionRanking(session.id),
        getMyBidStatus(session.id, userId ?? 'user-001'),
      ])
      const runtime = mapAuctionRuntime(items, session, ranking, myStatus)

      return {
        ...runtime,
        onlineCount: room?.onlineCount ?? 0,
      }
    },
  })
}
