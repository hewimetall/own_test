package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestRoutesServeHomePage(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "Please sign in") {
		t.Fatalf("expected home page HTML, got %q", recorder.Body.String())
	}
}

func TestRoutesServeGraphiQL(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/graphiql", nil)
	recorder := httptest.NewRecorder()

	routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !strings.Contains(strings.ToLower(recorder.Body.String()), "graphiql") {
		t.Fatalf("expected GraphiQL response, got %q", recorder.Body.String())
	}
}

func TestWebsocketServerEchoesMessages(t *testing.T) {
	server := httptest.NewServer(routes())
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/echo"
	conn, response, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		status := 0
		if response != nil {
			status = response.StatusCode
		}
		t.Fatalf("websocket dial failed with status %d: %v", status, err)
	}
	defer conn.Close()

	if err := conn.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
		t.Fatalf("write websocket message: %v", err)
	}
	msgType, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read websocket message: %v", err)
	}
	if msgType != websocket.TextMessage || string(msg) != "hello" {
		t.Fatalf("expected echoed text message, got type %d body %q", msgType, string(msg))
	}
}

func TestWebsocketServerRejectsPlainHTTP(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/echo", nil)
	recorder := httptest.NewRecorder()

	websocketServer(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
}
