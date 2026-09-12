// Package api expõe a API Gin: contas, lançamentos, cotações e carteira.
package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"financas-backend/internal/model"
	"financas-backend/internal/portfolio"
	"financas-backend/internal/quotes"
	"financas-backend/internal/store"
)

// QuoteService busca cotações e câmbio.
type QuoteService interface {
	FetchAll(ctx context.Context, br, us []string) (map[string]quotes.Quote, map[string]string)
	FXRate(ctx context.Context) (float64, error)
}

var txnTypes = map[string]bool{"deposit": true, "withdraw": true, "buy": true, "sell": true, "dividend": true}

func validCurrency(c string) bool {
	if len(c) != 3 {
		return false
	}
	for _, r := range c {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

// NewRouter cria o engine Gin com as rotas.
func NewRouter(st store.Store, qs QuoteService) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors())

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/api/accounts", func(c *gin.Context) {
		list, err := st.ListAccounts()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao listar"})
			return
		}
		if list == nil {
			list = []model.Account{}
		}
		c.JSON(http.StatusOK, list)
	})
	r.POST("/api/accounts", func(c *gin.Context) {
		var in struct {
			Name     string `json:"name"`
			Currency string `json:"currency"`
		}
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "json inválido"})
			return
		}
		in.Name = strings.TrimSpace(in.Name)
		in.Currency = strings.ToUpper(strings.TrimSpace(in.Currency))
		if in.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name é obrigatório"})
			return
		}
		if !validCurrency(in.Currency) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "currency deve ter 3 letras (ex. BRL, USD)"})
			return
		}
		a := model.Account{Name: in.Name, Currency: in.Currency, CreatedAt: time.Now()}
		if err := st.CreateAccount(&a); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao criar"})
			return
		}
		c.JSON(http.StatusCreated, a)
	})
	r.GET("/api/accounts/:id", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		a, err := st.GetAccount(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "não encontrada"})
			return
		}
		c.JSON(http.StatusOK, a)
	})
	r.PUT("/api/accounts/:id", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var in struct {
			Name     string `json:"name"`
			Currency string `json:"currency"`
		}
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "json inválido"})
			return
		}
		a, err := st.GetAccount(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "não encontrada"})
			return
		}
		if strings.TrimSpace(in.Name) != "" {
			a.Name = strings.TrimSpace(in.Name)
		}
		if strings.TrimSpace(in.Currency) != "" {
			cur := strings.ToUpper(strings.TrimSpace(in.Currency))
			if !validCurrency(cur) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "currency deve ter 3 letras"})
				return
			}
			a.Currency = cur
		}
		if err := st.UpdateAccount(&a); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao atualizar"})
			return
		}
		c.JSON(http.StatusOK, a)
	})
	r.DELETE("/api/accounts/:id", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := st.DeleteAccount(uint(id)); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "não encontrada"})
			return
		}
		c.Status(http.StatusNoContent)
	})

	r.GET("/api/transactions", func(c *gin.Context) {
		var accountID uint
		if v := c.Query("accountId"); v != "" {
			n, _ := strconv.ParseUint(v, 10, 64)
			accountID = uint(n)
		}
		list, err := st.ListTransactions(accountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao listar"})
			return
		}
		if list == nil {
			list = []model.Transaction{}
		}
		c.JSON(http.StatusOK, list)
	})
	r.POST("/api/transactions", func(c *gin.Context) {
		var in struct {
			AccountID uint    `json:"accountId"`
			Type      string  `json:"type"`
			Symbol    string  `json:"symbol"`
			Market    string  `json:"market"`
			Qty       float64 `json:"qty"`
			Price     float64 `json:"price"`
			Currency  string  `json:"currency"`
			Date      string  `json:"date"`
			Note      string  `json:"note"`
		}
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "json inválido"})
			return
		}
		in.Type = strings.ToLower(strings.TrimSpace(in.Type))
		if !txnTypes[in.Type] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "type inválido: deposit, withdraw, buy, sell ou dividend"})
			return
		}
		if in.AccountID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "accountId é obrigatório"})
			return
		}
		if _, err := st.GetAccount(in.AccountID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "conta não existe"})
			return
		}
		if in.Type == "buy" || in.Type == "sell" {
			if strings.TrimSpace(in.Symbol) == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "symbol é obrigatório para compra/venda"})
				return
			}
			if in.Qty <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "qty deve ser > 0"})
				return
			}
		}
		if in.Price < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "price inválido"})
			return
		}
		cur := strings.ToUpper(strings.TrimSpace(in.Currency))
		if !validCurrency(cur) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "currency deve ter 3 letras"})
			return
		}
		date := time.Now()
		if in.Date != "" {
			if d, err := time.Parse("2006-01-02", in.Date[:min(10, len(in.Date))]); err == nil {
				date = d
			}
		}
		market := strings.ToUpper(strings.TrimSpace(in.Market))
		if market != "BR" && market != "US" {
			market = "BR"
		}
		t := model.Transaction{
			AccountID: in.AccountID, Type: in.Type, Symbol: strings.ToUpper(strings.TrimSpace(in.Symbol)),
			Market: market, Qty: in.Qty, Price: in.Price, Currency: cur, Date: date, Note: in.Note,
		}
		if err := st.CreateTransaction(&t); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao criar"})
			return
		}
		c.JSON(http.StatusCreated, t)
	})
	r.DELETE("/api/transactions/:id", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := st.DeleteTransaction(uint(id)); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "não encontrada"})
			return
		}
		c.Status(http.StatusNoContent)
	})

	r.GET("/api/quotes", func(c *gin.Context) {
		br := splitSyms(c.Query("br"))
		us := splitSyms(c.Query("us"))
		got, errs := qs.FetchAll(c.Request.Context(), br, us)
		if got == nil {
			got = map[string]quotes.Quote{}
		}
		if len(errs) > 0 {
			c.JSON(http.StatusOK, gin.H{"quotes": got, "errors": errs})
			return
		}
		c.JSON(http.StatusOK, got)
	})

	r.GET("/api/portfolio", func(c *gin.Context) {
		var accountID uint
		if v := c.Query("accountId"); v != "" {
			n, _ := strconv.ParseUint(v, 10, 64)
			accountID = uint(n)
		}
		var txns []model.Transaction
		var err error
		if accountID != 0 {
			txns, err = st.ListTransactions(accountID)
		} else {
			txns, err = st.AllTransactions()
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao agregar"})
			return
		}
		inputs := make([]portfolio.TxnInput, 0, len(txns))
		brSet, usSet := map[string]bool{}, map[string]bool{}
		for _, t := range txns {
			inputs = append(inputs, portfolio.TxnInput{
				AccountID: t.AccountID, Type: t.Type, Symbol: t.Symbol,
				Market: t.Market, Currency: t.Currency, Qty: t.Qty, Price: t.Price,
			})
			if t.Type == "buy" || t.Type == "sell" {
				if t.Market == "US" {
					usSet[t.Symbol] = true
				} else {
					brSet[t.Symbol] = true
				}
			}
		}
		got, errs := qs.FetchAll(c.Request.Context(), keys(brSet), keys(usSet))
		qm := map[string]portfolio.QuotePrice{}
		for k, v := range got {
			qm[k] = portfolio.QuotePrice{Price: v.Price, Currency: v.Currency}
		}
		fx, err := qs.FXRate(c.Request.Context())
		fxLive := err == nil && fx > 0
		if !fxLive {
			fx = 0
		}
		s := portfolio.Build(inputs, qm, fx)
		c.JSON(http.StatusOK, gin.H{
			"positions": s.Positions, "cashByCurrency": s.CashByCurrency,
			"marketByCurrency": s.MarketByCurrency, "totalBRL": s.TotalBRL,
			"fxUsed": s.FXUsed, "fxLive": fxLive, "missingQuotes": s.MissingQuotes,
			"quoteErrors": errs,
		})
	})

	return r
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func splitSyms(v string) []string {
	var out []string
	for _, p := range strings.Split(v, ",") {
		if s := strings.ToUpper(strings.TrimSpace(p)); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
