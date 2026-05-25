package service

import "auction-live/backend/internal/model"

type RoomService struct {
	store *memoryStore
}

func NewRoomService() *RoomService {
	return &RoomService{store: sharedStore}
}

func (s *RoomService) ListRooms() []model.LiveRoom {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	rooms := make([]model.LiveRoom, 0, len(s.store.rooms))
	for _, room := range s.store.rooms {
		rooms = append(rooms, room)
	}

	return rooms
}

func (s *RoomService) GetRoomDetail(roomID string) model.LiveRoom {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	return s.store.rooms[roomID]
}
