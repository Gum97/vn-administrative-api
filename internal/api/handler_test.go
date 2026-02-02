package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"vn-admin-api/internal/logger"
)

func TestHandler_Root(t *testing.T) {
	// Setup
	log := logger.New("test.log", true)
	handler := NewHandler(nil, log) // Repo is not needed for Root handler

	// Create request
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.Root(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check response body
	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	expectedMessage := "Welcome to VN Administrative API"
	if response["message"] != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, response["message"])
	}
}
