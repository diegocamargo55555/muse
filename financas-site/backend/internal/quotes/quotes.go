// Package quotes busca cotações: brapi para ativos BR, Yahoo para exterior.
package quotes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Quote é uma cotação normalizada.
type Quote struct {
	Symbol   string  `json:"symbol"`
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
	Source   string  `json:"source"`
}

// Service busca cotações com failover query1→query2 no Yahoo.
type Service struct {
	brapiBase  string
	yahooBases []string
	token      string
	client     *http.Client
}

// New cria o serviço. Bases injetáveis para testes.
func New(brapiBase string, yahooBases []string, token string, client *http.Client) *Service {
	if brapiBase == "" {
		brapiBase = "https://brapi.dev"
	}
	if len(yahooBases) == 0 {
		yahooBases = []string{"https://query1.finance.yahoo.com", "https://query2.finance.yahoo.com"}
	}
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	return &Service{brapiBase: strings.TrimSuffix(brapiBase, "/"), yahooBases: yahooBases, token: token, client: client}
}

func (s *Service) get(ctx context.Context, target string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36")
	return s.client.Do(req)
}

type brapiResp struct {
	Results []struct {
		Symbol             string  `json:"symbol"`
		RegularMarketPrice float64 `json:"regularMarketPrice"`
		Currency           string  `json:"currency"`
	} `json:"results"`
}

// FetchBR cota tickers BR via brapi.
func (s *Service) FetchBR(ctx context.Context, tickers []string) (map[string]Quote, map[string]string) {
	got := map[string]Quote{}
	errs := map[string]string{}
	if len(tickers) == 0 {
		return got, errs
	}
	u, _ := url.Parse(s.brapiBase + "/api/quote/" + strings.Join(tickers, ","))
	qs := u.Query()
	if s.token != "" {
		qs.Set("token", s.token)
	}
	u.RawQuery = qs.Encode()
	res, err := s.get(ctx, u.String())
	if err != nil {
		for _, t := range tickers {
			errs[t] = "brapi indisponível"
		}
		return got, errs
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		for _, t := range tickers {
			errs[t] = fmt.Sprintf("brapi respondeu %s", res.Status)
		}
		return got, errs
	}
	var payload brapiResp
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		for _, t := range tickers {
			errs[t] = "resposta brapi inválida"
		}
		return got, errs
	}
	for _, r := range payload.Results {
		got[r.Symbol] = Quote{Symbol: r.Symbol, Price: r.RegularMarketPrice, Currency: r.Currency, Source: "brapi"}
	}
	for _, t := range tickers {
		if _, ok := got[t]; ok {
			continue
		}
		if _, bad := errs[t]; !bad {
			errs[t] = "ticker não retornado pela brapi"
		}
	}
	return got, errs
}

type yahooResp struct {
	Chart struct {
		Result []struct {
			Meta struct {
				RegularMarketPrice float64 `json:"regularMarketPrice"`
				Currency           string  `json:"currency"`
			} `json:"meta"`
		} `json:"result"`
	} `json:"chart"`
}

func (s *Service) fetchYahooOne(ctx context.Context, base, symbol string) (Quote, error) {
	u := fmt.Sprintf("%s/v8/finance/chart/%s?interval=1d&range=1d", strings.TrimSuffix(base, "/"), url.PathEscape(symbol))
	res, err := s.get(ctx, u)
	if err != nil {
		return Quote{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return Quote{}, fmt.Errorf("yahoo respondeu %s", res.Status)
	}
	var payload yahooResp
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return Quote{}, fmt.Errorf("resposta yahoo inválida")
	}
	if len(payload.Chart.Result) == 0 {
		return Quote{}, fmt.Errorf("símbolo sem cotação no yahoo")
	}
	m := payload.Chart.Result[0].Meta
	return Quote{Symbol: symbol, Price: m.RegularMarketPrice, Currency: m.Currency, Source: "yahoo"}, nil
}

// FetchUS cota símbolos do exterior via Yahoo com failover de base.
func (s *Service) FetchUS(ctx context.Context, symbols []string) (map[string]Quote, map[string]string) {
	got := map[string]Quote{}
	errs := map[string]string{}
	for _, sym := range symbols {
		var last error
		for _, base := range s.yahooBases {
			q, err := s.fetchYahooOne(ctx, base, sym)
			if err == nil {
				got[sym] = q
				last = nil
				break
			}
			last = err
		}
		if last != nil {
			errs[sym] = last.Error()
		}
	}
	return got, errs
}

// FetchAll combina BR + US.
func (s *Service) FetchAll(ctx context.Context, br, us []string) (map[string]Quote, map[string]string) {
	got, errs := s.FetchBR(ctx, br)
	g2, e2 := s.FetchUS(ctx, us)
	for k, v := range g2 {
		got[k] = v
	}
	for k, v := range e2 {
		errs[k] = v
	}
	return got, errs
}

// FXRate devolve USD→BRL via Yahoo (USDBRL=X).
func (s *Service) FXRate(ctx context.Context) (float64, error) {
	for _, base := range s.yahooBases {
		q, err := s.fetchYahooOne(ctx, base, "USDBRL=X")
		if err == nil && q.Price > 0 {
			return q.Price, nil
		}
	}
	return 0, fmt.Errorf("câmbio indisponível")
}
