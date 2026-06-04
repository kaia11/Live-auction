package service

import (
	"auction-live/backend/internal/model"
	"auction-live/backend/internal/repository"
)

type RoomService struct {
	store *memoryStore
	repo  repository.RoomRepository
}

func NewRoomService(repo repository.RoomRepository) *RoomService {
	return &RoomService{store: sharedStore, repo: repo}
}

func (s *RoomService) ListRooms() []model.LiveRoom {
	if s.repo != nil {
		rooms, err := s.repo.ListRooms()
		if err == nil && len(rooms) > 0 {
			return rooms
		}
	}

	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	rooms := make([]model.LiveRoom, 0, len(s.store.rooms))
	for _, room := range s.store.rooms {
		rooms = append(rooms, room)
	}

	return rooms
}

func (s *RoomService) ListRoomsByAnchorUserID(anchorUserID string) []model.LiveRoom {
	if s.repo != nil {
		rooms, err := s.repo.ListRoomsByAnchorUserID(anchorUserID)
		if err == nil {
			return rooms
		}
	}

	return s.store.ListRoomsByAnchorUserID(anchorUserID)
}

func (s *RoomService) GetRoomDetail(roomID string) model.LiveRoom {
	if s.repo != nil {
		room, err := s.repo.GetRoomDetail(roomID)
		if err == nil && room != nil && room.ID != "" {
			return *room
		}
	}

	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	return s.store.rooms[roomID]
}
