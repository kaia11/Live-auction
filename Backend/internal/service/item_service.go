package service

import "auction-live/backend/internal/model"

type ItemService struct {
	store *memoryStore
}

func NewItemService() *ItemService {
	return &ItemService{store: sharedStore}
}

func (s *ItemService) ListRoomItems(roomID string) []model.AuctionItem {
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
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	item := s.store.items[itemID]
	if item.RoomID != roomID {
		return model.AuctionItem{}
	}

	return item
}
