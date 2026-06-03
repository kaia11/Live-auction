import { useQuery } from '@tanstack/react-query'
import { BackendCommentPayload, getRoomEvents } from '@/api/rooms'
import { LiveComment } from '@/types'

export interface RoomCommentsSnapshot {
  comments: LiveComment[]
  lastEventVersion: number
}

export const roomCommentsQueryKey = (roomId: string | undefined) => ['room-comments', roomId]

export const useRoomCommentsQuery = (roomId: string | undefined) =>
  useQuery({
    queryKey: roomCommentsQueryKey(roomId),
    enabled: Boolean(roomId),
    queryFn: async (): Promise<RoomCommentsSnapshot> => {
      const events = await getRoomEvents(roomId as string)
      return {
        comments: events
          .filter((event) => event.event === 'room_comment_received')
          .map((event) => event.payload as BackendCommentPayload),
        lastEventVersion: events.length > 0 ? events[events.length - 1].version : 0,
      }
    },
  })

