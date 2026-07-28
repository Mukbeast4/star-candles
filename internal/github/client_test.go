package github

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestStarCount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo" {
			t.Errorf("path = %q, want /repos/owner/repo", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("Authorization = %q, want Bearer tok", got)
		}
		_, _ = fmt.Fprint(w, `{"stargazers_count": 1284}`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	count, err := client.StarCount(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("StarCount() error = %v", err)
	}
	if count != 1284 {
		t.Errorf("StarCount() = %d, want 1284", count)
	}
}

func TestStarCountNoTokenOmitsAuthorization(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("Authorization = %q, want empty", got)
		}
		_, _ = fmt.Fprint(w, `{"stargazers_count": 3}`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	if _, err := client.StarCount(context.Background(), "owner/repo"); err != nil {
		t.Fatalf("StarCount() error = %v", err)
	}
}

func TestStarCountAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"message": "Not Found"}`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	_, err := client.StarCount(context.Background(), "owner/repo")
	if err == nil {
		t.Fatal("StarCount() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error = %v, want status 404 mentioned", err)
	}
}

func TestStarEventsPaginatesAndSorts(t *testing.T) {
	page1 := make([]string, 100)
	base := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	for i := range page1 {
		page1[i] = fmt.Sprintf(`{"starred_at": %q}`, base.Add(time.Duration(99-i)*time.Hour).Format(time.RFC3339))
	}
	page2 := fmt.Sprintf(`[{"starred_at": %q}]`, base.Add(200*time.Hour).Format(time.RFC3339))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "application/vnd.github.star+json" {
			t.Errorf("Accept = %q, want star+json", got)
		}
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = fmt.Fprintf(w, "[%s]", strings.Join(page1, ","))
		case "2":
			_, _ = fmt.Fprint(w, page2)
		default:
			t.Errorf("unexpected page %q", r.URL.Query().Get("page"))
			_, _ = fmt.Fprint(w, "[]")
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	events, err := client.StarEvents(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("StarEvents() error = %v", err)
	}
	if len(events) != 101 {
		t.Fatalf("len(events) = %d, want 101", len(events))
	}
	for i := 1; i < len(events); i++ {
		if events[i].Before(events[i-1]) {
			t.Fatalf("events not sorted at index %d", i)
		}
	}
	if !events[100].Equal(base.Add(200 * time.Hour)) {
		t.Errorf("last event = %v, want %v", events[100], base.Add(200*time.Hour))
	}
}

func TestStarEventsSinglePage(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = fmt.Fprint(w, `[{"starred_at": "2026-07-01T10:00:00Z"}]`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	events, err := client.StarEvents(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("StarEvents() error = %v", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
	if len(events) != 1 {
		t.Errorf("len(events) = %d, want 1", len(events))
	}
}
