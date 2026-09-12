package main

import (
	"log"
	"net/http"
	"os"

	"mangasite-backend/internal/api"
	"mangasite-backend/internal/library"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		dataPath = "data/library.json"
	}
	upstream := os.Getenv("MANGADEX_API")
	if upstream == "" {
		upstream = "https://api.mangadex.org"
	}

	lib, err := library.New(dataPath)
	if err != nil {
		log.Fatalf("library: %v", err)
	}

	srv := api.NewServer(lib, upstream)
	log.Printf("manga-site API na porta %s (upstream %s)", port, upstream)
	if err := http.ListenAndServe(":"+port, srv.Routes()); err != nil {
		log.Fatal(err)
	}
}
