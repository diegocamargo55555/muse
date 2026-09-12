# Portfólio — Go + React + Docker (TDD)

Dark minimalista PT-BR. API Go com admin via token, frontend React, tudo dockerizado.
Cabe com folga em 4 vCPU / 8 GB (limites: API 1 CPU/512MB, web 0.5 CPU/256MB).

## TDD

Contrato primeiro em `backend/internal/store/store_test.go` e
`backend/internal/api/api_test.go` (falham sem implementação).
Implementação em `store.go` + `api.go`. Rodar:

```sh
cd backend
go test ./... -count=1
```

## Rodar local (sem Docker)

```sh
cd backend
PORT=8080 ADMIN_TOKEN=dev-token DATA_PATH=./data.json go run ./cmd/api

cd ../frontend
npm install
npm run dev
```

Frontend em http://localhost:5173 (proxy /api → :8080).

## Rodar com Docker

```sh
cd portfolio
ADMIN_TOKEN=troque-este-token WEB_PORT=3000 docker compose up --build -d
curl -s http://localhost:3000/api/health
curl -s http://localhost:3000/ | head -c 200
```

Admin no navegador: abra a URL, vá em Admin, cole o mesmo ADMIN_TOKEN.

## Estrutura

- `backend/` Go stdlib (sem framework), persistência JSON em volume.
- `frontend/` React + Vite, nginx serve estático e proxy /api.
- `docker-compose.yml` limites dimensionados para 4 cores / 8 GB.
