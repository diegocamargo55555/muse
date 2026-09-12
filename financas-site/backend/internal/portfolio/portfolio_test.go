package portfolio_test

import (
	"testing"

	"financas-backend/internal/portfolio"
)

func txns() []portfolio.TxnInput {
	return []portfolio.TxnInput{
		{AccountID: 1, Type: "deposit", Currency: "BRL", Qty: 1, Price: 10000},
		{AccountID: 1, Type: "buy", Symbol: "PETR4", Market: "BR", Currency: "BRL", Qty: 100, Price: 40},
		{AccountID: 1, Type: "buy", Symbol: "PETR4", Market: "BR", Currency: "BRL", Qty: 100, Price: 50},
		{AccountID: 2, Type: "deposit", Currency: "USD", Qty: 1, Price: 5000},
		{AccountID: 2, Type: "buy", Symbol: "AAPL", Market: "US", Currency: "USD", Qty: 10, Price: 200},
	}
}

func quotes() map[string]portfolio.QuotePrice {
	return map[string]portfolio.QuotePrice{
		"PETR4": {Price: 49, Currency: "BRL"},
		"AAPL":  {Price: 230, Currency: "USD"},
	}
}

func TestPositionsAvgCost(t *testing.T) {
	s := portfolio.Build(txns(), quotes(), 5.0)
	var petr *portfolio.Position
	for i := range s.Positions {
		if s.Positions[i].Symbol == "PETR4" {
			petr = &s.Positions[i]
		}
	}
	if petr == nil {
		t.Fatal("PETR4 deveria existir")
	}
	if petr.Qty != 200 || petr.AvgCost != 45 {
		t.Fatalf("qty=%v avg=%v", petr.Qty, petr.AvgCost)
	}
	if petr.MarketValue != 200*49 || petr.PnL != 200*49-200*45 {
		t.Fatalf("mv=%v pnl=%v", petr.MarketValue, petr.PnL)
	}
}

func TestSellReducesPosition(t *testing.T) {
	all := append(txns(), portfolio.TxnInput{AccountID: 1, Type: "sell", Symbol: "PETR4", Market: "BR", Currency: "BRL", Qty: 50, Price: 48})
	s := portfolio.Build(all, quotes(), 5.0)
	for _, p := range s.Positions {
		if p.Symbol == "PETR4" && p.Qty != 150 {
			t.Fatalf("qty=%v, want 150", p.Qty)
		}
	}
}

func TestCashAndTotalsBRL(t *testing.T) {
	s := portfolio.Build(txns(), quotes(), 5.0)
	// Caixa BRL: 10000 - 200*45(avg 45? custo real 4000+5000=9000) = 1000
	if s.CashByCurrency["BRL"] != 1000 {
		t.Fatalf("caixa BRL=%v, want 1000", s.CashByCurrency["BRL"])
	}
	if s.CashByCurrency["USD"] != 3000 {
		t.Fatalf("caixa USD=%v, want 3000", s.CashByCurrency["USD"])
	}
	// Total BRL: 9800 (PETR4) + 1000 (caixa BRL) + 2300*5 (AAPL) + 3000*5 (caixa USD) = 37300
	if s.TotalBRL != 37300 {
		t.Fatalf("total=%v, want 37300", s.TotalBRL)
	}
}

func TestMissingQuoteFlagged(t *testing.T) {
	s := portfolio.Build(txns(), map[string]portfolio.QuotePrice{}, 5.0)
	if len(s.MissingQuotes) != 2 {
		t.Fatalf("missing=%v", s.MissingQuotes)
	}
	for _, p := range s.Positions {
		if p.QuoteOK {
			t.Fatalf("QuoteOK deveria ser false: %+v", p)
		}
	}
}

func TestEmptyPortfolio(t *testing.T) {
	s := portfolio.Build(nil, quotes(), 5.0)
	if len(s.Positions) != 0 || s.TotalBRL != 0 {
		t.Fatalf("vazio deveria zerar: %+v", s)
	}
}
