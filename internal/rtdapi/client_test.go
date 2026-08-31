package rtdapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTriggerBuildPath(t *testing.T) {
	var gotPath string
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Header.Get("Authorization") != "Token secret" {
			t.Errorf("auth: %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(202)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"build": map[string]any{"id": 99, "commit": "abc"},
		})
	}))
	defer s.Close()
	c := New(s.URL, "secret")
	raw, err := c.TriggerBuild("github-utils", "latest")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/projects/github-utils/versions/latest/builds/" {
		t.Fatalf("path %s", gotPath)
	}
	if ExtractInt(raw, "id") == 0 {
		var env struct {
			Build json.RawMessage `json:"build"`
		}
		_ = json.Unmarshal(raw, &env)
		if ExtractInt(env.Build, "id") != 99 {
			t.Fatalf("id missing: %s", raw)
		}
	}
}

func TestListPagination(t *testing.T) {
	n := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		off := r.URL.Query().Get("offset")
		if off == "0" || off == "" {
			_ = json.NewEncoder(w).Encode(Page{
				Count: 2,
				Next:  strPtr("next"),
				Results: []json.RawMessage{
					json.RawMessage(`{"slug":"a"}`),
				},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(Page{
			Count:   2,
			Results: []json.RawMessage{json.RawMessage(`{"slug":"b"}`)},
		})
	}))
	defer s.Close()
	c := New(s.URL, "t")
	items, total, err := c.ListProjects(nil)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("got total=%d n=%d pages=%d", total, len(items), n)
	}
}

func strPtr(s string) *string { return &s }
