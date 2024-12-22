package server_test

import (
	"encoding/json"
	"exiftool-wrapper/server"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTagsRoute(t *testing.T) {
	req, err := http.NewRequest("GET", "/tags", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleTags)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	var response struct {
		Tags []struct {
			Path        string            `json:"path"`
			Writable    bool              `json:"writable"`
			Type        string            `json:"type"`
			Group       string            `json:"group"`
			Description map[string]string `json:"description"`
		} `json:"tags"`
	}

	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Tags) == 0 {
		t.Error("Expected tags in response, got none")
	}

	// Check that each tag has required fields populated
	for i, tag := range response.Tags {
		if tag.Path == "" {
			t.Errorf("Tag %d missing path", i)
		}
		if tag.Type == "" {
			t.Errorf("Tag %d missing type", i)
		}
		if tag.Group == "" {
			t.Errorf("Tag %d missing group", i)
		}
		if len(tag.Description) == 0 {
			t.Errorf("Tag %d missing description", i)
		}
	}
}
