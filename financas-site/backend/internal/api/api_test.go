package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"financas-backend/internal/api"
	"financas-backend/internal/model"
	"financas-backend/internal/quotes"
)

type fakeStore struct {
	accounts []model.Account
	txns     []model.Transaction
	seq      uint
}

func (f *fakeStore) CreateAccount(a *model.Account) error {
	f.seq++
	a.ID = f.seq
	f.accounts = append(f.accounts, *a)
	return nil
}

func (f *fakeStore) ListAccounts() ([]model.Account, error) { return f.accounts, nil }

func (f *fakeStore) GetAccount(id uint) (model.Account, error) {
	for _, a := range f.accounts {
		if a.ID == id {
			return a, nil
		}
	}
	return model.Account{}, errors.New("não encontrada")
}

func (f *fakeStore) UpdateAccount(a *model.Account) error {
	for i := range f.accounts {
		if f.accounts[i].ID == a.ID {
			f.accounts[i] = *a
			return nil
		}
	}
	return errors.New("não encontrada")
}

func (f *fakeStore) DeleteAccount(id uint) error {
	for i := range f.accounts {
		if f.accounts[i].ID == id {
			f.accounts = append(f.accounts[:i], f.accounts[i+1:]...)
			return nil
		}
	}
	return errors.New("não encontrada")
}

func (f *fakeStore) CreateTransaction(t *model.Transaction) error {
	f.seq++
	t.ID = f.seq
	f.txns = append(f.txns, *t)
	return nil
}

func (f *fakeStore) ListTransactions(accountID uint) ([]model.Transaction, error) {
	var out []model.Transaction
	for _, t := range f.txns {
		if accountID == 0 || t.AccountID == accountID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (f *fakeStore) DeleteTransaction(id uint) error {
	for i := range f.txns {
		if f.txns[i].ID == id {
			f.txns = append(f.txns[:i], f.txns[i+1:]...)
			return nil
		}
	}
	return errors.New("não encontrada")
}

func (f *fakeStore) AllTransactions() ([]model.Transaction, error) { return f.txns, nil }

type fakeQuotes struct{}

func (fakeQuotes) FetchAll(_ context.Context, br, us []string) (map[string]quotes.Quote, map[string]string) {
	out := map[string]quotes.Quote{}
	for _, s := range br {
		out[s] = quotes.Quote{Symbol: s, Price: 49, Currency: "BRL", Source: "brapi"}
	}
	for _, s := range us {
		out[s] = quotes.Quote{Symbol: s, Price: 230, Currency: "USD", Source: "yahoo"}
	}
	return out, map[string]string{}
}

func (fakeQuotes) FXRate(_ context.Context) (float64, error) { return 5.0, nil }

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return api.NewRouter(&fakeStore{}, fakeQuotes{})
}

func TestHealth(t *testing.T) {
	r := newTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/health", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestAccountsCRUD(t *testing.T) {
	r := newTestRouter()
	post := func(body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/accounts", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		return w
	}
	if w := post(`{"name":"Corretora BR","currency":"BRL"}`); w.Code != 201 {
		t.Fatalf("POST = %d: %s", w.Code, w.Body.String())
	}
	if w := post(`{"name":"X","currency":"reais"}`); w.Code != 400 {
		t.Fatalf("moeda inválida = %d, want 400", w.Code)
	}
	if w := post(`{"name":"","currency":"USD"}`); w.Code != 400 {
		t.Fatalf("sem nome = %d, want 400", w.Code)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/accounts", nil)
	r.ServeHTTP(w, req)
	var list []model.Account
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil || len(list) != 1 {
		t.Fatalf("list = %s", w.Body.String())
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/accounts/999", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != 404 {
		t.Fatalf("inexistente = %d, want 404", w2.Code)
	}
}

func TestTransactionsValidation(t *testing.T) {
	r := newTestRouter()
	seed := httptest.NewRecorder()
	seedReq, _ := http.NewRequest("POST", "/api/accounts", bytes.NewBufferString(`{"name":"Principal","currency":"BRL"}`))
	seedReq.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(seed, seedReq)
	if seed.Code != 201 {
		t.Fatalf("seed conta = %d", seed.Code)
	}
	post := func(body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		return w
	}
	if w := post(`{"accountId":1,"type":"loteria","currency":"BRL","qty":1,"price":10}`); w.Code != 400 {
		t.Fatalf("tipo inválido = %d, want 400", w.Code)
	}
	if w := post(`{"accountId":1,"type":"buy","currency":"BRL","qty":10,"price":40}`); w.Code != 400 {
		t.Fatalf("buy sem symbol = %d, want 400", w.Code)
	}
	if w := post(`{"accountId":1,"type":"buy","symbol":"PETR4","market":"BR","currency":"BRL","qty":10,"price":40}`); w.Code != 201 {
		t.Fatalf("POST = %d: %s", w.Code, w.Body.String())
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/transactions?accountId=1", nil)
	r.ServeHTTP(w, req)
	var list []model.Transaction
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil || len(list) != 1 {
		t.Fatalf("list = %s", w.Body.String())
	}
}

func TestQuotesEndpoint(t *testing.T) {
	r := newTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/quotes?br=PETR4&us=AAPL", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
	var payload map[string]quotes.Quote
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload["PETR4"].Source != "brapi" || payload["AAPL"].Source != "yahoo" {
		t.Fatalf("payload = %+v", payload)
	}
}

func TestPortfolioEndpoint(t *testing.T) {
	r := newTestRouter()
	seed := httptest.NewRecorder()
	seedReq, _ := http.NewRequest("POST", "/api/accounts", bytes.NewBufferString(`{"name":"Principal","currency":"BRL"}`))
	seedReq.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(seed, seedReq)
	if seed.Code != 201 {
		t.Fatalf("seed conta = %d", seed.Code)
	}
	for _, b := range []string{
		`{"accountId":1,"type":"deposit","currency":"BRL","qty":1,"price":10000}`,
		`{"accountId":1,"type":"buy","symbol":"PETR4","market":"BR","currency":"BRL","qty":100,"price":40}`,
	} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBufferString(b))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 201 {
			t.Fatalf("seed = %d", w.Code)
		}
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/portfolio", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
	var s struct {
		Positions []struct {
			Symbol string
			Qty    float64
		}
		TotalBRL float64 `json:"totalBRL"`
	}
	if err := json.NewDecoder(w.Body).Decode(&s); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(s.Positions) != 1 || s.Positions[0].Qty != 100 {
		t.Fatalf("positions = %+v", s.Positions)
	}
	// mercado 100*49=4900 + caixa 6000 = 10900
	if s.TotalBRL != 10900 {
		t.Fatalf("total=%v, want 10900", s.TotalBRL)
	}
}
