// Package portfolio agrega transações em posições e totais.
package portfolio

import "sort"

// TxnInput é um lançamento para agregação.
type TxnInput struct {
	AccountID uint
	Type      string
	Symbol    string
	Market    string
	Currency  string
	Qty       float64
	Price     float64
}

// QuotePrice é o preço de mercado de um símbolo.
type QuotePrice struct {
	Price    float64
	Currency string
}

// Position é a posição agregada num símbolo.
type Position struct {
	Symbol      string  `json:"symbol"`
	Market      string  `json:"market"`
	Currency    string  `json:"currency"`
	Qty         float64 `json:"qty"`
	AvgCost     float64 `json:"avgCost"`
	Price       float64 `json:"price"`
	MarketValue float64 `json:"marketValue"`
	CostBasis   float64 `json:"costBasis"`
	PnL         float64  `json:"pnl"`
	QuoteOK     bool    `json:"quoteOK"`
}

// Summary é a carteira consolidada.
type Summary struct {
	Positions       []Position       `json:"positions"`
	CashByCurrency  map[string]float64 `json:"cashByCurrency"`
	MarketByCurrency map[string]float64 `json:"marketByCurrency"`
	TotalBRL        float64          `json:"totalBRL"`
	FXUsed          float64          `json:"fxUsed"`
	FXLive          bool             `json:"fxLive"`
	MissingQuotes   []string         `json:"missingQuotes"`
}

func toBRL(v float64, cur string, fx float64) float64 {
	if cur == "USD" {
		return v * fx
	}
	return v
}

// Build agrega lançamentos. fxUSD converte USD→BRL.
func Build(txns []TxnInput, quotes map[string]QuotePrice, fxUSD float64) Summary {
	s := Summary{
		CashByCurrency:   map[string]float64{},
		MarketByCurrency: map[string]float64{},
		FXUsed:           fxUSD,
		FXLive:           fxUSD > 0,
	}
	if fxUSD <= 0 {
		fxUSD = 1
		s.FXUsed = 1
	}
	type acc struct {
		qty    float64
		cost   float64
		market string
		cur    string
	}
	pos := map[string]*acc{}
	for _, t := range txns {
		amount := t.Qty * t.Price
		switch t.Type {
		case "deposit", "dividend":
			s.CashByCurrency[t.Currency] += amount
		case "withdraw":
			s.CashByCurrency[t.Currency] -= amount
		case "buy":
			s.CashByCurrency[t.Currency] -= amount
			p := pos[t.Symbol]
			if p == nil {
				p = &acc{market: t.Market, cur: t.Currency}
				pos[t.Symbol] = p
			}
			p.qty += t.Qty
			p.cost += amount
		case "sell":
			s.CashByCurrency[t.Currency] += amount
			if p := pos[t.Symbol]; p != nil && p.qty > 0 {
				avg := p.cost / p.qty
				q := t.Qty
				if q > p.qty {
					q = p.qty
				}
				p.qty -= q
				p.cost -= avg * q
			}
		}
	}
	for sym, p := range pos {
		if p.qty <= 0 {
			continue
		}
		avg := p.cost / p.qty
		q, ok := quotes[sym]
		pp := Position{Symbol: sym, Market: p.market, Currency: p.cur, Qty: p.qty, AvgCost: avg, CostBasis: p.cost, QuoteOK: ok}
		if ok {
			pp.Price = q.Price
			pp.Currency = q.Currency
			pp.MarketValue = p.qty * q.Price
			pp.PnL = pp.MarketValue - p.cost
			s.MarketByCurrency[q.Currency] += pp.MarketValue
		} else {
			pp.MarketValue = 0
			pp.PnL = -p.cost
			s.MissingQuotes = append(s.MissingQuotes, sym)
		}
		s.Positions = append(s.Positions, pp)
	}
	sort.Slice(s.Positions, func(i, j int) bool { return s.Positions[i].Symbol < s.Positions[j].Symbol })
	sort.Strings(s.MissingQuotes)
	total := 0.0
	for cur, v := range s.CashByCurrency {
		total += toBRL(v, cur, fxUSD)
	}
	for cur, v := range s.MarketByCurrency {
		total += toBRL(v, cur, fxUSD)
	}
	s.TotalBRL = total
	return s
}
