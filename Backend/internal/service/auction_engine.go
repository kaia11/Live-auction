package service

import (
	"fmt"
	"time"

	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/model"
	"auction-live/backend/internal/realtime"
	"auction-live/backend/internal/repository"
	"auction-live/backend/internal/statemachine"
)

type AuctionEngine struct {
	store       *memoryStore
	runtime     *realtime.Runtime
	roomRepo    repository.RoomRepository
	itemRepo    repository.ItemRepository
	sessionRepo repository.SessionRepository
	resultRepo  repository.ResultRepository
	orderRepo   repository.OrderRepository
}

type SessionSettlement struct {
	RoomID        string              `json:"roomId"`
	SessionID     string              `json:"sessionId"`
	ItemID        string              `json:"itemId"`
	Status        string              `json:"status"`
	QueueStatus   string              `json:"queueStatus"`
	CurrentPrice  int64               `json:"currentPrice"`
	WinnerUserID  string              `json:"winnerUserId"`
	NextSessionID string              `json:"nextSessionId,omitempty"`
	NextItemID    string              `json:"nextItemId,omitempty"`
	Order         *model.AuctionOrder `json:"order,omitempty"`
}

func NewAuctionEngine(store *memoryStore, runtime *realtime.Runtime, roomRepo repository.RoomRepository, itemRepo repository.ItemRepository, sessionRepo repository.SessionRepository, resultRepo repository.ResultRepository, orderRepo repository.OrderRepository) *AuctionEngine {
	return &AuctionEngine{
		store:       store,
		runtime:     runtime,
		roomRepo:    roomRepo,
		itemRepo:    itemRepo,
		sessionRepo: sessionRepo,
		resultRepo:  resultRepo,
		orderRepo:   orderRepo,
	}
}

func (e *AuctionEngine) StartSessionLocked(sessionID string) (map[string]any, error) {
	session, ok := e.store.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}

	item, ok := e.store.items[session.ItemID]
	if !ok {
		return nil, ErrItemNotFound
	}

	nextSessionStatus, err := statemachine.NextSessionState(session.Status, statemachine.SessionEventStart)
	if err != nil {
		return nil, fmt.Errorf("%w: cannot start session %s from status=%s", ErrInvalidSessionState, sessionID, session.Status)
	}

	nextQueueStatus, err := statemachine.NextQueueState(item.QueueStatus, statemachine.QueueEventActivate)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: cannot activate item %s for session %s from queue_status=%s",
			ErrInvalidQueueOrder,
			item.ID,
			sessionID,
			item.QueueStatus,
		)
	}

	session.Status = nextSessionStatus
	session.CurrentPrice = item.StartPrice
	session.EndTime = time.Now().Add(time.Duration(item.DurationSeconds) * time.Second).Format(time.RFC3339)
	session.IncrementStep = item.IncrementStep
	session.ExtensionSeconds = item.ExtensionSeconds
	session.ExtensionTrigger = item.ExtensionTriggerSeconds
	session.CeilingPrice = item.CeilingPrice
	e.store.sessions[sessionID] = session

	item.QueueStatus = nextQueueStatus
	e.store.items[item.ID] = item

	room := e.store.rooms[session.RoomID]
	room.CurrentSessionID = sessionID
	e.store.rooms[session.RoomID] = room
	if err := e.persistRoomItemSession(room, item, session); err != nil {
		return nil, err
	}

	if err := saveSessionState(e.runtime, session, item); err != nil {
		return nil, err
	}
	if e.runtime != nil {
		if err := e.runtime.SetRoomCurrentSession(session.RoomID, sessionID); err != nil {
			return nil, err
		}
	}

	return map[string]any{
		"roomId":      session.RoomID,
		"sessionId":   sessionID,
		"status":      session.Status,
		"endTime":     session.EndTime,
		"queueStatus": item.QueueStatus,
	}, nil
}

func (e *AuctionEngine) CancelSessionLocked(sessionID string) (map[string]any, error) {
	session, ok := e.store.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}

	if session.Status == domain.SessionStateEndedSold || session.Status == domain.SessionStateEndedPassed {
		return nil, ErrInvalidSessionState
	}

	item, ok := e.store.items[session.ItemID]
	if !ok {
		return nil, ErrItemNotFound
	}

	nextSessionStatus, err := statemachine.NextSessionState(session.Status, statemachine.SessionEventCancel)
	if err != nil {
		return nil, fmt.Errorf("%w: cannot cancel session %s from status=%s", ErrInvalidSessionState, sessionID, session.Status)
	}

	nextQueueStatus, err := statemachine.NextQueueState(item.QueueStatus, statemachine.QueueEventCancel)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: cannot cancel item %s for session %s from queue_status=%s",
			ErrInvalidQueueOrder,
			item.ID,
			sessionID,
			item.QueueStatus,
		)
	}

	session.Status = nextSessionStatus
	e.store.sessions[sessionID] = session

	item.QueueStatus = nextQueueStatus
	e.store.items[item.ID] = item
	if err := e.persistItemSession(item, session); err != nil {
		return nil, err
	}

	if err := saveSessionState(e.runtime, session, item); err != nil {
		return nil, err
	}

	return map[string]any{
		"roomId":      session.RoomID,
		"sessionId":   sessionID,
		"status":      session.Status,
		"queueStatus": item.QueueStatus,
	}, nil
}

func (e *AuctionEngine) ActivateNextItemLocked(roomID string) (map[string]any, error) {
	room, ok := e.store.rooms[roomID]
	if !ok {
		return nil, ErrRoomNotFound
	}

	if room.CurrentSessionID != "" {
		currentSession, ok := e.store.sessions[room.CurrentSessionID]
		if !ok {
			return nil, ErrSessionNotFound
		}

		if currentSession.Status == domain.SessionStatePending || currentSession.Status == domain.SessionStateBidding {
			if _, err := e.CancelSessionLocked(currentSession.ID); err != nil {
				return nil, err
			}
		}
	}

	nextSessionID, nextItemID, err := e.prepareNextSessionLocked(roomID)
	if err != nil {
		return nil, err
	}

	nextSession := e.store.sessions[nextSessionID]
	nextItem := e.store.items[nextItemID]

	return map[string]any{
		"roomId":        roomID,
		"nextItemId":    nextItemID,
		"nextSessionId": nextSessionID,
		"status":        nextSession.Status,
		"queueStatus":   nextItem.QueueStatus,
	}, nil
}

func (e *AuctionEngine) SettleSessionLocked(sessionID string) (SessionSettlement, error) {
	session, ok := e.store.sessions[sessionID]
	if !ok {
		return SessionSettlement{}, ErrSessionNotFound
	}

	if session.Status != domain.SessionStateBidding {
		return SessionSettlement{}, ErrInvalidSessionState
	}

	event := statemachine.SessionEventTimeoutNoBid
	if session.LeaderUserID != "" {
		event = statemachine.SessionEventTimeoutWithWinner
	}

	return e.settleSessionWithEventLocked(session, event)
}

func (e *AuctionEngine) ReachCeilingLocked(sessionID string) (SessionSettlement, error) {
	session, ok := e.store.sessions[sessionID]
	if !ok {
		return SessionSettlement{}, ErrSessionNotFound
	}

	if session.Status != domain.SessionStateBidding {
		return SessionSettlement{}, ErrInvalidSessionState
	}

	return e.settleSessionWithEventLocked(session, statemachine.SessionEventReachCeiling)
}

func (e *AuctionEngine) settleSessionWithEventLocked(session model.AuctionSession, event statemachine.SessionEvent) (SessionSettlement, error) {
	item, ok := e.store.items[session.ItemID]
	if !ok {
		return SessionSettlement{}, ErrItemNotFound
	}

	nextSessionStatus, err := statemachine.NextSessionState(session.Status, event)
	if err != nil {
		return SessionSettlement{}, ErrInvalidSessionState
	}

	nextQueueStatus, err := statemachine.NextQueueState(item.QueueStatus, statemachine.QueueEventFinish)
	if err != nil {
		if item.QueueStatus == domain.QueueStateFinished || item.QueueStatus == domain.QueueStateCancelled {
			return e.reconcileTerminalSessionLocked(session, item)
		}
		return SessionSettlement{}, fmt.Errorf(
			"%w: cannot finish item %s for session %s from queue_status=%s",
			ErrInvalidQueueOrder,
			item.ID,
			session.ID,
			item.QueueStatus,
		)
	}

	session.Status = nextSessionStatus
	session.EndTime = time.Now().Format(time.RFC3339)
	e.store.sessions[session.ID] = session

	item.QueueStatus = nextQueueStatus
	e.store.items[item.ID] = item
	if err := e.persistItemSession(item, session); err != nil {
		return SessionSettlement{}, err
	}

	if err := saveSessionState(e.runtime, session, item); err != nil {
		return SessionSettlement{}, err
	}

	outcome := SessionSettlement{
		RoomID:       session.RoomID,
		SessionID:    session.ID,
		ItemID:       session.ItemID,
		Status:       session.Status,
		QueueStatus:  item.QueueStatus,
		CurrentPrice: session.CurrentPrice,
		WinnerUserID: session.LeaderUserID,
	}

	if session.Status == domain.SessionStateEndedSold && session.LeaderUserID != "" {
		order := e.ensureOrderLocked(session)
		outcome.Order = &order
	}

	if e.resultRepo != nil {
		resultStatus := "unsold"
		if session.Status == domain.SessionStateEndedSold {
			resultStatus = "sold"
		} else if session.Status == domain.SessionStateCancelled {
			resultStatus = "cancelled"
		}
		_ = e.resultRepo.CreateResult(model.AuctionResult{
			SessionID:        session.ID,
			ItemID:           session.ItemID,
			ResultStatus:     resultStatus,
			WinnerUserID:     session.LeaderUserID,
			FinalPrice:       session.CurrentPrice,
			ParticipantCount: session.ParticipantCount,
		})
	}

	nextSessionID, nextItemID, err := e.prepareNextSessionLocked(session.RoomID)
	if err == nil {
		outcome.NextSessionID = nextSessionID
		outcome.NextItemID = nextItemID
	} else if err != ErrQueueExhausted {
		return SessionSettlement{}, err
	}

	return outcome, nil
}

func (e *AuctionEngine) reconcileTerminalSessionLocked(session model.AuctionSession, item model.AuctionItem) (SessionSettlement, error) {
	if item.QueueStatus == domain.QueueStateCancelled {
		session.Status = domain.SessionStateCancelled
	} else if session.LeaderUserID != "" {
		session.Status = domain.SessionStateEndedSold
	} else {
		session.Status = domain.SessionStateEndedPassed
	}
	session.EndTime = time.Now().Format(time.RFC3339)
	e.store.sessions[session.ID] = session

	if err := e.persistItemSession(item, session); err != nil {
		return SessionSettlement{}, err
	}
	if err := saveSessionState(e.runtime, session, item); err != nil {
		return SessionSettlement{}, err
	}

	outcome := SessionSettlement{
		RoomID:       session.RoomID,
		SessionID:    session.ID,
		ItemID:       session.ItemID,
		Status:       session.Status,
		QueueStatus:  item.QueueStatus,
		CurrentPrice: session.CurrentPrice,
		WinnerUserID: session.LeaderUserID,
	}

	if session.Status == domain.SessionStateEndedSold && session.LeaderUserID != "" {
		order := e.ensureOrderLocked(session)
		outcome.Order = &order
	}

	if e.resultRepo != nil {
		resultStatus := "unsold"
		if session.Status == domain.SessionStateEndedSold {
			resultStatus = "sold"
		} else if session.Status == domain.SessionStateCancelled {
			resultStatus = "cancelled"
		}
		_ = e.resultRepo.CreateResult(model.AuctionResult{
			SessionID:        session.ID,
			ItemID:           session.ItemID,
			ResultStatus:     resultStatus,
			WinnerUserID:     session.LeaderUserID,
			FinalPrice:       session.CurrentPrice,
			ParticipantCount: session.ParticipantCount,
		})
	}

	nextSessionID, nextItemID, err := e.prepareNextSessionLocked(session.RoomID)
	if err == nil {
		outcome.NextSessionID = nextSessionID
		outcome.NextItemID = nextItemID
	} else if err != ErrQueueExhausted {
		return SessionSettlement{}, err
	}

	return outcome, nil
}

func (e *AuctionEngine) prepareNextSessionLocked(roomID string) (string, string, error) {
	nextSessionID, nextItemID := e.findNextQueuedLocked(roomID)
	if nextSessionID == "" || nextItemID == "" {
		return "", "", ErrQueueExhausted
	}

	nextItem := e.store.items[nextItemID]
	nextQueueStatus, err := statemachine.NextQueueState(nextItem.QueueStatus, statemachine.QueueEventMarkUpcoming)
	if err != nil {
		return "", "", fmt.Errorf(
			"%w: cannot mark upcoming item %s in room %s from queue_status=%s",
			ErrInvalidQueueOrder,
			nextItemID,
			roomID,
			nextItem.QueueStatus,
		)
	}
	nextItem.QueueStatus = nextQueueStatus
	e.store.items[nextItemID] = nextItem

	nextSession := e.store.sessions[nextSessionID]
	nextSessionStatus, err := statemachine.NextSessionState(nextSession.Status, statemachine.SessionEventResetToPending)
	if err != nil {
		return "", "", fmt.Errorf(
			"%w: cannot reset session %s in room %s from status=%s",
			ErrInvalidSessionState,
			nextSessionID,
			roomID,
			nextSession.Status,
		)
	}
	nextSession.Status = nextSessionStatus
	nextSession.CurrentPrice = nextItem.StartPrice
	nextSession.LeaderUserID = ""
	nextSession.ParticipantCount = 0
	nextSession.EndTime = ""
	nextSession.IncrementStep = nextItem.IncrementStep
	nextSession.ExtensionSeconds = nextItem.ExtensionSeconds
	nextSession.ExtensionTrigger = nextItem.ExtensionTriggerSeconds
	nextSession.CeilingPrice = nextItem.CeilingPrice
	e.store.sessions[nextSessionID] = nextSession

	room := e.store.rooms[roomID]
	room.CurrentSessionID = nextSessionID
	e.store.rooms[roomID] = room
	if err := e.persistRoomItemSession(room, nextItem, nextSession); err != nil {
		return "", "", err
	}

	if err := saveSessionState(e.runtime, nextSession, nextItem); err != nil {
		return "", "", err
	}
	if e.runtime != nil {
		if err := e.runtime.SetRoomCurrentSession(roomID, nextSessionID); err != nil {
			return "", "", err
		}
	}

	return nextSessionID, nextItemID, nil
}

func (e *AuctionEngine) findNextQueuedLocked(roomID string) (string, string) {
	for _, itemID := range e.store.roomItems[roomID] {
		item := e.store.items[itemID]
		if item.QueueStatus == domain.QueueStateQueued || item.QueueStatus == domain.QueueStateUpcoming {
			for sessionID, session := range e.store.sessions {
				if session.RoomID == roomID && session.ItemID == itemID {
					return sessionID, itemID
				}
			}
		}
	}

	return "", ""
}

func (e *AuctionEngine) ensureOrderLocked(session model.AuctionSession) model.AuctionOrder {
	if order, ok := e.store.ordersBySession[session.ID]; ok {
		return order
	}

	order := model.AuctionOrder{
		ID:          fmt.Sprintf("order-%s", session.ID),
		SessionID:   session.ID,
		RoomID:      session.RoomID,
		ItemID:      session.ItemID,
		BuyerUserID: session.LeaderUserID,
		Amount:      session.CurrentPrice,
		Status:      domain.OrderStatusPendingPayment,
		CreateTime:  time.Now().Format(time.RFC3339),
	}

	e.store.ordersBySession[session.ID] = order
	e.store.ordersByID[order.ID] = order
	e.store.orders = append(e.store.orders, order)
	if e.orderRepo != nil {
		_ = e.orderRepo.CreateOrder(order)
	}
	return order
}

func (e *AuctionEngine) persistRoomItemSession(room model.LiveRoom, item model.AuctionItem, session model.AuctionSession) error {
	if e.roomRepo != nil {
		if err := e.roomRepo.SaveRoom(room); err != nil {
			return err
		}
	}
	if e.itemRepo != nil {
		if err := e.itemRepo.SaveItem(item); err != nil {
			return err
		}
	}
	if e.sessionRepo != nil {
		if err := e.sessionRepo.SaveSession(session); err != nil {
			return err
		}
	}
	return nil
}

func (e *AuctionEngine) persistItemSession(item model.AuctionItem, session model.AuctionSession) error {
	if e.itemRepo != nil {
		if err := e.itemRepo.SaveItem(item); err != nil {
			return err
		}
	}
	if e.sessionRepo != nil {
		if err := e.sessionRepo.SaveSession(session); err != nil {
			return err
		}
	}
	return nil
}
