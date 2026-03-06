package ws

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

type nodeStatusUpdate struct {
	serverID int64
	status   string
	lastSeen string
	version  string
}

type recordingNodeStatusStore struct {
	mu      sync.Mutex
	updates []nodeStatusUpdate
}

func (s *recordingNodeStatusStore) UpdateNodeStatus(_ context.Context, serverID int64, status, lastSeen, version string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updates = append(s.updates, nodeStatusUpdate{serverID: serverID, status: status, lastSeen: lastSeen, version: version})
	return nil
}

func (s *recordingNodeStatusStore) latest() nodeStatusUpdate {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.updates) == 0 {
		return nodeStatusUpdate{}
	}
	return s.updates[len(s.updates)-1]
}

func TestHandlerHeartbeatPersistsOnlineAndAcks(t *testing.T) {
	serverConn, clientConn := newTestWebsocketPair(t)
	conn := NewConn(serverConn, 42)
	store := &recordingNodeStatusStore{}
	handler := NewHandler(NewHub(), nil, nil, store, slog.New(slog.NewTextHandler(io.Discard, nil)))

	payload, err := json.Marshal(Heartbeat{Version: "1.2.3", CPUPercent: 10, MemUsedMB: 64, MemTotalMB: 128})
	if err != nil {
		t.Fatalf("marshal heartbeat: %v", err)
	}
	handler.handleHeartbeat(context.Background(), conn, Envelope{Type: TypeHeartbeat, Payload: payload})

	_, message, err := clientConn.Read(context.Background())
	if err != nil {
		t.Fatalf("read ack: %v", err)
	}
	var env Envelope
	if err := json.Unmarshal(message, &env); err != nil {
		t.Fatalf("unmarshal ack envelope: %v", err)
	}
	if env.Type != TypeHeartbeatAck {
		t.Fatalf("ack type = %q, want %q", env.Type, TypeHeartbeatAck)
	}
	if got := store.latest(); got.status != string(AgentStatusOnline) || got.version != "1.2.3" {
		t.Fatalf("latest update = %+v, want online with version", got)
	}
}

func TestHandleConnectionMarksNodeUnhealthyOnDisconnect(t *testing.T) {
	serverConn, clientConn := newTestWebsocketPair(t)
	hub := NewHub()
	conn := NewConn(serverConn, 9)
	hub.Register(conn)
	store := &recordingNodeStatusStore{}
	handler := NewHandler(hub, nil, nil, store, slog.New(slog.NewTextHandler(io.Discard, nil)))

	done := make(chan struct{})
	go func() {
		handler.HandleConnection(context.Background(), conn)
		close(done)
	}()

	if err := clientConn.Close(websocket.StatusNormalClosure, "bye"); err != nil {
		t.Fatalf("close client: %v", err)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not exit after disconnect")
	}

	if _, ok := hub.Get(9); ok {
		t.Fatal("expected connection to be unregistered")
	}
	if got := store.latest(); got.status != string(AgentStatusUnhealthy) {
		t.Fatalf("latest update = %+v, want unhealthy", got)
	}
}

func newTestWebsocketPair(t *testing.T) (*websocket.Conn, *websocket.Conn) {
	t.Helper()
	accepted := make(chan *websocket.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Errorf("accept websocket: %v", err)
			return
		}
		accepted <- conn
	}))
	t.Cleanup(srv.Close)

	clientConn, _, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	t.Cleanup(func() { _ = clientConn.Close(websocket.StatusNormalClosure, "") })

	select {
	case serverConn := <-accepted:
		t.Cleanup(func() { _ = serverConn.Close(websocket.StatusNormalClosure, "") })
		return serverConn, clientConn
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for server websocket")
		return nil, nil
	}
}
