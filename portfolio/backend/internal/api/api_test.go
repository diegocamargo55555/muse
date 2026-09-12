package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"portfolio-backend/internal/api"
	"portfolio-backend/internal/store"
)

func newTestServer(t *testing.T) (*httptest.Server, *store.Store) {
	t.Helper()
	s, err := store.New(filepath.Join(t.TempDir(), "data.json"))
	if err != nil {
		t.Fatalf("New store: %v", err)
	}
	if _, err := s.Create(store.Project{ID: "site-investimentos", Title: "Site de Investimentos", Description: "Carteiras", Tech: []string{"Go"}}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	srv := api.NewServer(s, "test-token")
	ts := httptest.NewServer(srv.Routes())
	t.Cleanup(ts.Close)
	return ts, s
}

func TestHealth(t *testing.T) {
	ts, _ := newTestServer(t)
	res, err := http.Get(ts.URL + "/api/health")
	if err != nil {
		t.Fatalf("GET health: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
}

func TestListAndGetProjects(t *testing.T) {
	ts, _ := newTestServer(t)
	res, err := http.Get(ts.URL + "/api/projects")
	if err != nil {
		t.Fatalf("GET list: %v", err)
	}
	defer res.Body.Close()
	var list []store.Project
	if err := json.NewDecoder(res.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 1 || list[0].ID != "site-investimentos" {
		t.Fatalf("list = %+v", list)
	}

	res2, err := http.Get(ts.URL + "/api/projects/site-investimentos")
	if err != nil {
		t.Fatalf("GET one: %v", err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != 200 {
		t.Fatalf("get status = %d", res2.StatusCode)
	}

	res3, err := http.Get(ts.URL + "/api/projects/nao-existe")
	if err != nil {
		t.Fatalf("GET inexistente: %v", err)
	}
	defer res3.Body.Close()
	if res3.StatusCode != 404 {
		t.Fatalf("get inexistente = %d, want 404", res3.StatusCode)
	}
}

func TestWriteRequiresAuth(t *testing.T) {
	ts, _ := newTestServer(t)
	body, _ := json.Marshal(store.Project{ID: "novo", Title: "Novo"})
	res, err := http.Post(ts.URL+"/api/projects", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST sem token: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 401 {
		t.Fatalf("sem token = %d, want 401", res.StatusCode)
	}
}

func TestCreateUpdateDeleteWithAuth(t *testing.T) {
	ts, _ := newTestServer(t)
	client := ts.Client()

	do := func(method, url string, v any) *http.Response {
		var buf bytes.Buffer
		if v != nil {
			_ = json.NewEncoder(&buf).Encode(v)
		}
		req, _ := http.NewRequest(method, ts.URL+url, &buf)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("%s %s: %v", method, url, err)
		}
		return res
	}

	res := do("POST", "/api/projects", store.Project{ID: "site-manga", Title: "Site de Mangá", Tech: []string{"React"}})
	if res.StatusCode != 201 {
		t.Fatalf("POST = %d, want 201", res.StatusCode)
	}
	res.Body.Close()

	res = do("PUT", "/api/projects/site-manga", store.Project{Title: "Mangá v2"})
	if res.StatusCode != 200 {
		t.Fatalf("PUT = %d, want 200", res.StatusCode)
	}
	res.Body.Close()

	req, _ := http.NewRequest("DELETE", ts.URL+"/api/projects/site-manga", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	del, err := client.Do(req)
	if err != nil {
		t.Fatalf("DELETE: %v", err)
	}
	defer del.Body.Close()
	if del.StatusCode != 204 {
		t.Fatalf("DELETE = %d, want 204", del.StatusCode)
	}
}

func TestCORSHeaders(t *testing.T) {
	ts, _ := newTestServer(t)
	req, _ := http.NewRequest("OPTIONS", ts.URL+"/api/projects", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("OPTIONS: %v", err)
	}
	defer res.Body.Close()
	if res.Header.Get("Access-Control-Allow-Origin") == "" {
		t.Fatal("CORS header ausente")
	}
}
