package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"mangasite-backend/internal/api"
	"mangasite-backend/internal/library"
)

func newTestServer(t *testing.T, upstream string) *httptest.Server {
	t.Helper()
	s, err := library.New(filepath.Join(t.TempDir(), "lib.json"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	srv := api.NewServer(s, upstream)
	ts := httptest.NewServer(srv.Routes())
	t.Cleanup(ts.Close)
	return ts
}

func TestHealth(t *testing.T) {
	ts := newTestServer(t, "http://example.com")
	res, err := http.Get(ts.URL + "/api/health")
	if err != nil {
		t.Fatalf("GET health: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("status = %d", res.StatusCode)
	}
}

func TestLibraryCRUD(t *testing.T) {
	ts := newTestServer(t, "http://example.com")
	client := ts.Client()

	res, err := http.Get(ts.URL + "/api/library")
	if err != nil {
		t.Fatalf("GET library: %v", err)
	}
	var list []library.Entry
	if err := json.NewDecoder(res.Body).Decode(&list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	res.Body.Close()
	if len(list) != 0 {
		t.Fatalf("list = %d, want 0", len(list))
	}

	body, _ := json.Marshal(library.Entry{MangaID: "m1", Title: "Berserk", Status: "reading"})
	post, err := client.Post(ts.URL+"/api/library", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	post.Body.Close()
	if post.StatusCode != 201 {
		t.Fatalf("POST = %d, want 201", post.StatusCode)
	}

	bad, _ := json.Marshal(library.Entry{MangaID: "m2", Title: "X", Status: "readding"})
	badRes, err := client.Post(ts.URL+"/api/library", "application/json", bytes.NewReader(bad))
	if err != nil {
		t.Fatalf("POST bad: %v", err)
	}
	badRes.Body.Close()
	if badRes.StatusCode != 400 {
		t.Fatalf("POST inválido = %d, want 400", badRes.StatusCode)
	}

	updBody, _ := json.Marshal(library.Entry{Status: "finished"})
	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/api/library/m1", bytes.NewReader(updBody))
	req.Header.Set("Content-Type", "application/json")
	put, err := client.Do(req)
	if err != nil {
		t.Fatalf("PUT: %v", err)
	}
	put.Body.Close()
	if put.StatusCode != 200 {
		t.Fatalf("PUT = %d, want 200", put.StatusCode)
	}

	filtered, err := http.Get(ts.URL + "/api/library?status=finished")
	if err != nil {
		t.Fatalf("GET filter: %v", err)
	}
	var f []library.Entry
	if err := json.NewDecoder(filtered.Body).Decode(&f); err != nil {
		t.Fatalf("decode filter: %v", err)
	}
	filtered.Body.Close()
	if len(f) != 1 {
		t.Fatalf("filter = %d, want 1", len(f))
	}

	delReq, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/library/m1", nil)
	del, err := client.Do(delReq)
	if err != nil {
		t.Fatalf("DELETE: %v", err)
	}
	del.Body.Close()
	if del.StatusCode != 204 {
		t.Fatalf("DELETE = %d, want 204", del.StatusCode)
	}
}

func TestMangadexProxyPassthrough(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":"ok","echo":"` + r.URL.RequestURI() + `"}`))
	}))
	defer upstream.Close()

	ts := newTestServer(t, upstream.URL)
	res, err := http.Get(ts.URL + "/api/mangadex/search?title=naruto&limit=5")
	if err != nil {
		t.Fatalf("GET search: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("search = %d", res.StatusCode)
	}
	var payload map[string]string
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload["result"] != "ok" {
		t.Fatalf("payload = %+v", payload)
	}
}

func TestMangadexProxyError(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer upstream.Close()

	ts := newTestServer(t, upstream.URL)
	res, err := http.Get(ts.URL + "/api/mangadex/search?title=x")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 502 {
		t.Fatalf("upstream 429 deveria virar 502, got %d", res.StatusCode)
	}
}

func TestCORS(t *testing.T) {
	ts := newTestServer(t, "http://example.com")
	req, _ := http.NewRequest("OPTIONS", ts.URL+"/api/library", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("OPTIONS: %v", err)
	}
	defer res.Body.Close()
	if res.Header.Get("Access-Control-Allow-Origin") == "" {
		t.Fatal("CORS ausente")
	}
}
