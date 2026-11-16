package task

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	NewHandler(nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"ok"`) {
		t.Fatalf("health: status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestInvalidCreateInput(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"empty title", `{"title":"  "}`},
		{"unknown field", `{"title":"Hello","unknown":true}`},
		{"invalid JSON", `{"title":`},
		{"two objects", `{"title":"A"}{"title":"B"}`},
		{"long title", `{"title":"` + strings.Repeat("a", 201) + `"}`},
		{"long description", `{"title":"Valid","description":"` + strings.Repeat("a", 1001) + `"}`},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			NewHandler(nil).ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("got status %d, want 400: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestInvalidIDAndMethod(t *testing.T) {
	tests := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/tasks/abc", http.StatusBadRequest},
		{http.MethodGet, "/tasks/0", http.StatusBadRequest},
		{http.MethodPatch, "/tasks/1", http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		rec := httptest.NewRecorder()
		NewHandler(nil).ServeHTTP(rec, req)
		if rec.Code != tt.want {
			t.Errorf("%s %s: got %d, want %d", tt.method, tt.path, rec.Code, tt.want)
		}
	}
}
