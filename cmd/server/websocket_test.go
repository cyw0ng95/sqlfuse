package main

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func TestWebSocketLogStreaming(t *testing.T) {
	// Initialize server config for testing
	serverConfig = ServerConfig{
		Port:                "8080",
		ExecutorsConfigPath: "/tmp/test-executors.json",
		ServerName:          "test server",
		ServerVersion:       "0.1",
		Job: JobConfig{
			MaxOutputBytes: 65536,
			PersistPath:    "/tmp/test-jobs.json",
		},
	}

	// Create a test job
	job := &Job{
		ID:             "test-1",
		Cmd:            "test command",
		Status:         JobRunning,
		logSubscribers: make(map[*websocket.Conn]bool),
	}

	// Store the job
	jobMu.Lock()
	jobStore[job.ID] = job
	jobMu.Unlock()

	// Create Echo instance
	e := echo.New()
	RegisterJobRoutes(e)

	// Start test server
	server := httptest.NewServer(e)
	defer server.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/job/logs/stream?id=test-1"

	// Connect to WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer ws.Close()

	// Test broadcasting a message
	go func() {
		time.Sleep(100 * time.Millisecond)
		job.broadcastLogMessage("stdout", "test message\n")
	}()

	// Read messages with timeout
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	
	foundTestMessage := false
	for i := 0; i < 3; i++ {
		_, message, err := ws.ReadMessage()
		if err != nil {
			if i == 0 {
				t.Fatalf("Failed to read first message: %v", err)
			}
			break
		}
		
		// Check if this is the test message
		if bytes.Contains(message, []byte("test message")) {
			foundTestMessage = true
			break
		}
	}

	if !foundTestMessage {
		t.Error("Did not receive the test message")
	}

	// Clean up
	jobMu.Lock()
	delete(jobStore, job.ID)
	jobMu.Unlock()
}

func TestStreamReader(t *testing.T) {
	job := &Job{
		ID:             "test-2",
		Cmd:            "test command",
		Status:         JobRunning,
		logSubscribers: make(map[*websocket.Conn]bool),
	}

	// Create a test reader
	testData := "line 1\nline 2\nline 3\n"
	reader := strings.NewReader(testData)
	var buffer bytes.Buffer

	// Run streamReader
	streamReader(job, "stdout", reader, &buffer)

	// Verify buffer contains all lines
	if buffer.String() != testData {
		t.Errorf("Expected buffer to contain %q, got %q", testData, buffer.String())
	}
}
