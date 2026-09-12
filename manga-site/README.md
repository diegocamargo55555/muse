# Mangateca — MangaDex + estante (Go + React + Docker)

Busca e detalhe via MangaDex (proxy no backend, sem CORS no browser),
capítulos por idioma (EN/PT-BR) e estante com
Lendo · Quero ler · Concluído · Abandonado, persistida em volume.

## Testes (TDD)

```sh
cd backend
go test ./... -count=1
```

## Rodar local (sem Docker)

```sh
cd backend
PORT=8080 DATA_PATH=./lib.json go run ./cmd/api

cd ../frontend
npm install
npm run dev
```

Frontend em http://localhost:5173 (proxy /api → :8080).

## Rodar com Docker

```sh
cd manga-site
WEB_PORT=3002 docker compose up --build -d
curl -s "http://localhost:3002/api/mangadex/search?title=berserk&limit=2" | head -c 300
curl -s http://localhost:3002/api/library
```

## Notas

- Leitura das páginas não incluída (escopo: busca + detalhe + capítulos);
  o detalhe linka a página do título no MangaDex.
- Rate limit do MangaDex (5 req/s) respeitado pelo uso normal;
  em 429 o proxy responde 502 com mensagem.
