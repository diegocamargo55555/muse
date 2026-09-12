package main

import (
	"log"
	"net/http"
	"os"

	"portfolio-backend/internal/api"
	"portfolio-backend/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		dataPath = "data/projects.json"
	}
	token := os.Getenv("ADMIN_TOKEN")

	s, err := store.New(dataPath)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	s.SeedIfEmpty()

	srv := api.NewServer(s, token)
	log.Printf("portfolio API na porta %s", port)
	if err := http.ListenAndServe(":"+port, srv.Routes()); err != nil {
		log.Fatal(err)
	}
}
