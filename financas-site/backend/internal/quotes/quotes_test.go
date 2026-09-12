package quotes_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"financas-backend/internal/quotes"
)

func TestFetchBRParsesBrapi(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"symbol":"PETR4","regularMarketPrice":49.10,"currency":"BRL"}]}`))
	}))
	defer up.Close()

	svc := quotes.New(up.URL, []string{up.URL}, "", nil)
	got, errs := svc.FetchBR(context.Background(), []string{"PETR4"})
	if len(errs) != 0 {
		t.Fatalf("errs = %v", errs)
	}
	q, ok := got["PETR4"]
	if !ok || q.Price != 49.10 || q.Currency != "BRL" || q.Source != "brapi" {
		t.Fatalf("quote = %+v", q)
	}
}

func TestFetchBRUpstreamError(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer up.Close()

	svc := quotes.New(up.URL, []string{up.URL}, "", nil)
	_, errs := svc.FetchBR(context.Background(), []string{"PETR4"})
	if errs["PETR4"] == "" {
		t.Fatal("esperava erro para PETR4")
	}
}

func TestFetchUSParsesYahoo(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"chart":{"result":[{"meta":{"regularMarketPrice":231.5,"currency":"USD"}}],"error":null}}`))
	}))
	defer up.Close()

	svc := quotes.New("http://example.com", []string{up.URL}, "", nil)
	got, errs := svc.FetchUS(context.Background(), []string{"AAPL"})
	if len(errs) != 0 {
		t.Fatalf("errs = %v", errs)
	}
	if got["AAPL"].Price != 231.5 || got["AAPL"].Source != "yahoo" {
		t.Fatalf("quote = %+v", got["AAPL"])
	}
}

func TestFetchUSFailoverToSecondBase(t *testing.T) {
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer bad.Close()
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"chart":{"result":[{"meta":{"regularMarketPrice":100.0,"currency":"USD"}}],"error":null}}`))
	}))
	defer good.Close()

	svc := quotes.New("http://example.com", []string{bad.URL, good.URL}, "", nil)
	got, errs := svc.FetchUS(context.Background(), []string{"MSFT"})
	if len(errs) != 0 {
		t.Fatalf("errs = %v", errs)
	}
	if got["MSFT"].Price != 100.0 {
		t.Fatalf("quote = %+v", got["MSFT"])
	}
}

func TestFetchUSAllBasesDown(t *testing.T) {
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer bad.Close()

	svc := quotes.New("http://example.com", []string{bad.URL}, "", nil)
	_, errs := svc.FetchUS(context.Background(), []string{"AAPL"})
	if errs["AAPL"] == "" {
		t.Fatal("esperava erro quando Yahoo fora")
	}
}

func TestFXRate(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"chart":{"result":[{"meta":{"regularMarketPrice":5.42,"currency":"BRL"}}],"error":null}}`))
	}))
	defer up.Close()

	svc := quotes.New("http://example.com", []string{up.URL}, "", nil)
	rate, err := svc.FXRate(context.Background())
	if err != nil {
		t.Fatalf("FXRate: %v", err)
	}
	if rate != 5.42 {
		t.Fatalf("rate = %v", rate)
	}
}
