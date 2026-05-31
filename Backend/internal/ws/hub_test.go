package ws

import "testing"

func TestHubBroadcastsToRoomClientsOnly(t *testing.T) {
	hub := NewHub()

	roomA := NewClient("room-a", 1)
	roomB := NewClient("room-b", 1)

	unregisterA := hub.Register("room-a", roomA)
	defer unregisterA()
	defer roomA.Close()

	unregisterB := hub.Register("room-b", roomB)
	defer unregisterB()
	defer roomB.Close()

	hub.Publish("room-a", EventAuctionPriceUpdated, map[string]any{"price": 100})

	select {
	case msg := <-roomA.Messages():
		if msg.RoomID != "room-a" {
			t.Fatalf("expected room-a message, got %s", msg.RoomID)
		}
		if msg.Event != EventAuctionPriceUpdated {
			t.Fatalf("expected event %s, got %s", EventAuctionPriceUpdated, msg.Event)
		}
	default:
		t.Fatal("expected room-a client to receive broadcast")
	}

	select {
	case msg := <-roomB.Messages():
		t.Fatalf("did not expect room-b message, got %+v", msg)
	default:
	}
}
