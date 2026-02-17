package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestBroadcastFanoutToAllClients(t *testing.T) {
	// Reset globals to a known state for the test.
	channelManager = NewChannelManager()
	channelLoggers = make(map[string]*log.Logger)

	endpoint := "/clonos/containers/"
	trimmed := strings.TrimSuffix(endpoint, "/")
	channelManager.RegisterEndpoint(endpoint)

	mux := http.NewServeMux()
	mux.HandleFunc(endpoint, handleConnections)
	mux.HandleFunc(trimmed, handleConnections)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	go handleMessages(endpoint)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + endpoint

	c1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial c1: %v", err)
	}
	t.Cleanup(func() { _ = c1.Close() })

	c2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial c2: %v", err)
	}
	t.Cleanup(func() { _ = c2.Close() })

	// Give the server a moment to register clients.
	time.Sleep(50 * time.Millisecond)

	want := []byte(`{"id":"t","cmd":"jstart"}`)
	if err := c1.WriteMessage(websocket.TextMessage, want); err != nil {
		t.Fatalf("c1 write: %v", err)
	}

	_ = c2.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, got, err := c2.ReadMessage()
	if err != nil {
		t.Fatalf("c2 read: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("unexpected message: got=%q want=%q", got, want)
	}
}

func TestConnectWithoutTrailingSlashStillWorks(t *testing.T) {
	channelManager = NewChannelManager()
	channelLoggers = make(map[string]*log.Logger)

	endpoint := "/clonos/containers/"
	trimmed := strings.TrimSuffix(endpoint, "/")
	channelManager.RegisterEndpoint(endpoint)

	mux := http.NewServeMux()
	mux.HandleFunc(endpoint, handleConnections)
	mux.HandleFunc(trimmed, handleConnections)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	go handleMessages(endpoint)

	// Dial without trailing slash.
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + trimmed
	c1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = c1.Close() })

	want := []byte(`{"hello":"world"}`)
	if err := c1.WriteMessage(websocket.TextMessage, want); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Since this server broadcasts to all clients including sender,
	// the sender should receive its own message back.
	_ = c1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, got, err := c1.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("unexpected message: got=%q want=%q", got, want)
	}
}

