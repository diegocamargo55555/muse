FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -o /api ./cmd/api

FROM alpine:3.20
RUN adduser -D -u 10001 app && mkdir -p /data && chown app /data
USER app
COPY --from=build /api /api
ENV PORT=8080 DATA_PATH=/data/projects.json
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --retries=3 CMD wget -qO- http://127.0.0.1:8080/api/health || exit 1
ENTRYPOINT ["/api"]
