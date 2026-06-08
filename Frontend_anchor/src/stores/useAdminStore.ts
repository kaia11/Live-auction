import { create } from 'zustand'
import type { LiveRoom } from '@/types'

const ROOM_STORAGE_KEY = 'live-auction-anchor-room-id'

const getStoredRoomId = () => {
  if (typeof window === 'undefined') {
    return ''
  }

  return window.localStorage.getItem(ROOM_STORAGE_KEY) ?? ''
}

const setStoredRoomId = (roomId: string) => {
  if (typeof window === 'undefined') {
    return
  }

  window.localStorage.setItem(ROOM_STORAGE_KEY, roomId)
}

interface AdminState {
  rooms: LiveRoom[]
  currentRoomId: string
  setRooms: (rooms: LiveRoom[]) => void
  setCurrentRoomId: (roomId: string) => void
  updateRoomStatus: (roomId: string, status: LiveRoom['status']) => void
}

export const useAdminStore = create<AdminState>((set) => ({
  rooms: [],
  currentRoomId: getStoredRoomId(),
  setRooms: (rooms) =>
    set((state) => {
      const nextRoomId =
        state.currentRoomId && rooms.some((room) => room.id === state.currentRoomId)
          ? state.currentRoomId
          : rooms[0]?.id ?? ''

      if (nextRoomId) {
        setStoredRoomId(nextRoomId)
      }

      return {
        rooms,
        currentRoomId: nextRoomId,
      }
    }),
  setCurrentRoomId: (roomId) => {
    setStoredRoomId(roomId)
    set({ currentRoomId: roomId })
  },
  updateRoomStatus: (roomId, status) =>
    set((state) => ({
      rooms: state.rooms.map((room) => (room.id === roomId ? { ...room, status } : room)),
    })),
}))
