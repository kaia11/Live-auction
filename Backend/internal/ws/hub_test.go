package ws

import "testing"

func TestHubBroadcastsToRoomClientsOnly(t *testing.T) {
	hub := NewHub(nil)

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

func TestHubRoomClientCountTracksConnections(t *testing.T) {
	hub := NewHub(nil)

	roomA1 := NewClient("room-a", 1)
	roomA2 := NewClient("room-a", 1)

	unregisterA1 := hub.Register("room-a", roomA1)
	defer roomA1.Close()
	if got := hub.RoomClientCount("room-a"); got != 1 {
		t.Fatalf("expected 1 client, got %d", got)
	}

	unregisterA2 := hub.Register("room-a", roomA2)
	defer roomA2.Close()
	if got := hub.RoomClientCount("room-a"); got != 2 {
		t.Fatalf("expected 2 clients, got %d", got)
	}

	unregisterA2()
	if got := hub.RoomClientCount("room-a"); got != 1 {
		t.Fatalf("expected 1 client after unregister, got %d", got)
	}

	unregisterA1()
	if got := hub.RoomClientCount("room-a"); got != 0 {
		t.Fatalf("expected 0 clients after unregister, got %d", got)
	}
}
