package task

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestCRUDIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration test")
	}

	// Bảo vệ database phát triển khỏi bị dùng nhầm khi chạy integration test.
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Path != "/taskdb_test" {
		t.Fatal("TEST_DATABASE_URL must target the dedicated taskdb_test database")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatalf("PostgreSQL test database unavailable: %v", err)
	}
	api := NewHandler(NewPostgresStore(db))

	request := func(method, path, body string, expected int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		rec := httptest.NewRecorder()
		api.ServeHTTP(rec, req)
		if rec.Code != expected {
			t.Fatalf("%s %s: got %d, want %d: %s", method, path, rec.Code, expected, rec.Body.String())
		}
		return rec
	}

	createdResponse := request(http.MethodPost, "/tasks",
		`{"title":"Learn Go","description":"PostgreSQL CRUD","completed":false}`,
		http.StatusCreated)
	var created Task
	if err := json.Unmarshal(createdResponse.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ID <= 0 || created.Title != "Learn Go" || created.CreatedAt.IsZero() {
		t.Fatalf("unexpected created task: %+v", created)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM tasks WHERE id=$1`, created.ID)
	})

	path := "/tasks/" + strconv.FormatInt(created.ID, 10)
	if got := createdResponse.Header().Get("Location"); got != path {
		t.Fatalf("Location header = %q, want %q", got, path)
	}

	rec := request(http.MethodGet, path, "", http.StatusOK)
	var fetched Task
	if err := json.Unmarshal(rec.Body.Bytes(), &fetched); err != nil {
		t.Fatal(err)
	}
	if fetched.Title != "Learn Go" {
		t.Fatalf("GET: unexpected task %+v", fetched)
	}

	rec = request(http.MethodGet, "/tasks", "", http.StatusOK)
	var listed []Task
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range listed {
		if item.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("created task not present in GET /tasks")
	}

	rec = request(http.MethodPut, path,
		`{"title":"Learn Go deeply","description":"finished","completed":true}`,
		http.StatusOK)
	var updated Task
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}
	if updated.Title != "Learn Go deeply" || !updated.Completed {
		t.Fatalf("PUT: unexpected task %+v", updated)
	}

	request(http.MethodDelete, path, "", http.StatusNoContent)
	request(http.MethodGet, path, "", http.StatusNotFound)
	request(http.MethodDelete, path, "", http.StatusNotFound)
	fmt.Println("CRUD integration test passed")
}
