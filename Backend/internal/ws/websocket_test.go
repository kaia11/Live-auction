package ws

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpgradeRejectsNonWebSocketRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ws?roomId=room-001", nil)
	recorder := httptest.NewRecorder()

	_, err := Upgrade(recorder, req)
	if err == nil {
		t.Fatal("expected upgrade to fail for non-websocket request")
	}
}
