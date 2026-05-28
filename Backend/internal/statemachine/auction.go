package statemachine

import (
	"errors"

	"auction-live/backend/internal/domain"
)

var ErrInvalidTransition = errors.New("invalid state transition")

type SessionEvent string

const (
	SessionEventStart             SessionEvent = "start"
	SessionEventReachCeiling      SessionEvent = "reach_ceiling"
	SessionEventTimeoutWithWinner SessionEvent = "timeout_with_winner"
	SessionEventTimeoutNoBid      SessionEvent = "timeout_no_bid"
	SessionEventCancel            SessionEvent = "cancel"
	SessionEventResetToPending    SessionEvent = "reset_to_pending"
)

type QueueEvent string

const (
	QueueEventMarkUpcoming QueueEvent = "mark_upcoming"
	QueueEventActivate     QueueEvent = "activate"
	QueueEventFinish       QueueEvent = "finish"
	QueueEventCancel       QueueEvent = "cancel"
	QueueEventResetQueued  QueueEvent = "reset_queued"
)

func NextSessionState(current string, event SessionEvent) (string, error) {
	switch current {
	case domain.SessionStatePending:
		switch event {
		case SessionEventStart:
			return domain.SessionStateBidding, nil
		case SessionEventCancel:
			return domain.SessionStateCancelled, nil
		case SessionEventResetToPending:
			return domain.SessionStatePending, nil
		}
	case domain.SessionStateBidding:
		switch event {
		case SessionEventReachCeiling, SessionEventTimeoutWithWinner:
			return domain.SessionStateEndedSold, nil
		case SessionEventTimeoutNoBid:
			return domain.SessionStateEndedPassed, nil
		case SessionEventCancel:
			return domain.SessionStateCancelled, nil
		}
	case domain.SessionStateEndedSold, domain.SessionStateEndedPassed, domain.SessionStateCancelled:
		if event == SessionEventResetToPending {
			return domain.SessionStatePending, nil
		}
	}

	return "", ErrInvalidTransition
}

func NextQueueState(current string, event QueueEvent) (string, error) {
	switch current {
	case domain.QueueStateQueued:
		switch event {
		case QueueEventMarkUpcoming:
			return domain.QueueStateUpcoming, nil
		case QueueEventCancel:
			return domain.QueueStateCancelled, nil
		case QueueEventResetQueued:
			return domain.QueueStateQueued, nil
		}
	case domain.QueueStateUpcoming:
		switch event {
		case QueueEventActivate:
			return domain.QueueStateActive, nil
		case QueueEventCancel:
			return domain.QueueStateCancelled, nil
		case QueueEventResetQueued:
			return domain.QueueStateQueued, nil
		}
	case domain.QueueStateActive:
		switch event {
		case QueueEventFinish:
			return domain.QueueStateFinished, nil
		case QueueEventCancel:
			return domain.QueueStateCancelled, nil
		}
	case domain.QueueStateFinished, domain.QueueStateCancelled:
		if event == QueueEventResetQueued {
			return domain.QueueStateQueued, nil
		}
	}

	return "", ErrInvalidTransition
}
