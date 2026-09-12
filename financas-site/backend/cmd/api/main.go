package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"financas-backend/internal/api"
	"financas-backend/internal/quotes"
	"financas-backend/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL é obrigatória (ex. postgres://fin:fin@db:5432/financas?sslmode=disable)")
	}

	gin.SetMode(gin.ReleaseMode)

	st, err := store.Open(dsn)
	if err != nil {
		log.Fatalf("banco: %v", err)
	}
	qs := quotes.New(
		getenv("BRAPI_API", "https://brapi.dev"),
		[]string{"https://query1.finance.yahoo.com", "https://query2.finance.yahoo.com"},
		os.Getenv("BRAPI_TOKEN"),
		nil,
	)

	log.Printf("finanças API na porta %s", port)
	if err := api.NewRouter(st, qs).Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
