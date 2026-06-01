package service

import (
	"auction-live/backend/internal/model"
	"auction-live/backend/internal/repository"
)

type ItemService struct {
	store *memoryStore
	repo  repository.ItemRepository
}

func NewItemService(repo repository.ItemRepository) *ItemService {
	return &ItemService{store: sharedStore, repo: repo}
}

func (s *ItemService) ListRoomItems(roomID string) []model.AuctionItem {
	if s.repo != nil {
		items, err := s.repo.ListRoomItems(roomID)
		if err == nil && len(items) > 0 {
			return items
		}
	}

	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	itemIDs := s.store.roomItems[roomID]
	items := make([]model.AuctionItem, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		items = append(items, s.store.items[itemID])
	}

	return items
}

func (s *ItemService) GetItemDetail(roomID string, itemID string) model.AuctionItem {
	if s.repo != nil {
		item, err := s.repo.GetItemDetail(roomID, itemID)
		if err == nil && item != nil && item.ID != "" {
			return *item
		}
	}

	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	item := s.store.items[itemID]
	if item.RoomID != roomID {
		return model.AuctionItem{}
	}

	return item
}
