# Carteira — finanças e investimentos (Gin + GORM + Postgres)

Contas em várias moedas, lançamentos (aporte, retirada, compra, venda,
dividendo) e carteira com cotações: **brapi** para BR, **Yahoo** para exterior.
Câmbio USD→BRL ao vivo via Yahoo (com aviso quando indisponível).

## Testes

```sh
cd backend
go test ./... -count=1
```

## Rodar com Docker

```sh
cd financas-site
WEB_PORT=3003 docker compose up --build -d
curl -s http://localhost:3003/api/health
```

Opcional: `BRAPI_TOKEN=...` para maior cota da brapi.

## Uso

1. Contas → crie uma por moeda/corretora (ex. `Corretora BR/BRL`, `Broker US/USD`).
2. Lançamentos → aporte na conta, depois compras (`PETR4` mercado BR, `AAPL` mercado US).
3. Carteira → posições com preço médio, P/L e total em BRL.

## Notas

- Yahoo limita por IP (429); a API retorna `quoteErrors` por símbolo sem
  quebrar a carteira, e o frontend avisa para tentar de novo em 1 min.
- `DELETE /api/accounts/:id` remove a conta e seus lançamentos.
