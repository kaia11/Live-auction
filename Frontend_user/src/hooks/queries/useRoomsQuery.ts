import { useQuery } from '@tanstack/react-query'
import { getRooms } from '@/api/rooms'
import { mapBackendRoom } from '@/adapters/auction'

export const roomsQueryKey = ['rooms']

export const useRoomsQuery = () =>
  useQuery({
    queryKey: roomsQueryKey,
    queryFn: async () => {
      const rooms = await getRooms()
      return rooms.map(mapBackendRoom)
    },
  })

