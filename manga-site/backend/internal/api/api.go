package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"mangasite-backend/internal/library"
)

// Server expõe proxy MangaDex + biblioteca.
type Server struct {
	lib      *library.Store
	upstream string
	client   *http.Client
}

// NewServer cria o handler. upstream é a base do MangaDex (injetável p/ testes).
func NewServer(lib *library.Store, upstream string) *Server {
	if upstream == "" {
		upstream = "https://api.mangadex.org"
	}
	return &Server{
		lib:      lib,
		upstream: strings.TrimSuffix(upstream, "/"),
		client:   &http.Client{Timeout: 15 * time.Second},
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Routes registra as rotas.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/mangadex/search", s.handleSearch)
	mux.HandleFunc("/api/mangadex/manga/", s.handleMangaSub)
	mux.HandleFunc("/api/library", s.handleLibrary)
	mux.HandleFunc("/api/library/", s.handleLibraryByID)
	return cors(mux)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func clampInt(v string, def, max int) int {
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	if n > max {
		return max
	}
	return n
}

// proxy repassa GET ao MangaDex e devolve o corpo como está.
func (s *Server) proxy(w http.ResponseWriter, target string) {
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "mangadex indisponível"})
		return
	}
	res, err := s.client.Do(req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "mangadex indisponível"})
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "mangadex respondeu " + res.Status})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, res.Body)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	u, _ := url.Parse(s.upstream + "/manga")
	qs := u.Query()
	if title := strings.TrimSpace(q.Get("title")); title != "" {
		qs.Set("title", title)
	}
	qs.Set("limit", strconv.Itoa(clampInt(q.Get("limit"), 12, 24)))
	qs.Set("offset", strconv.Itoa(clampInt(q.Get("offset"), 0, 10000)))
	qs.Set("includes[]", "cover_art")
	qs.Add("includes[]", "author")
	qs.Set("order[relevance]", "desc")
	qs.Set("contentRating[]", "safe")
	qs.Add("contentRating[]", "suggestive")
	qs.Set("availableTranslatedLanguage[]", "en")
	u.RawQuery = qs.Encode()
	s.proxy(w, u.String())
}

func (s *Server) handleMangaSub(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, "/api/mangadex/manga/")
	parts := strings.Split(rest, "/")
	if len(parts) == 1 && parts[0] != "" {
		u, _ := url.Parse(s.upstream + "/manga/" + parts[0])
		qs := u.Query()
		qs.Set("includes[]", "cover_art")
		qs.Add("includes[]", "author")
		qs.Add("includes[]", "artist")
		u.RawQuery = qs.Encode()
		s.proxy(w, u.String())
		return
	}
	if len(parts) == 2 && parts[1] == "chapters" && parts[0] != "" {
		q := r.URL.Query()
		u, _ := url.Parse(s.upstream + "/manga/" + parts[0] + "/feed")
		qs := u.Query()
		qs.Set("limit", strconv.Itoa(clampInt(q.Get("limit"), 20, 100)))
		qs.Set("offset", strconv.Itoa(clampInt(q.Get("offset"), 0, 10000)))
		qs.Set("order[chapter]", "asc")
		lang := q.Get("lang")
		if lang == "" {
			lang = "en"
		}
		qs.Set("translatedLanguage[]", lang)
		u.RawQuery = qs.Encode()
		s.proxy(w, u.String())
		return
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "não encontrado"})
}

func (s *Server) handleLibrary(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		status := r.URL.Query().Get("status")
		filter := ""
		if status != "" {
			n, err := library.NormalizeStatus(status)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			filter = n
		}
		writeJSON(w, http.StatusOK, s.lib.List(filter))
	case http.MethodPost:
		var e library.Entry
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "json inválido"})
			return
		}
		created, err := s.lib.Add(e)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleLibraryByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/library/")
	if id == "" || strings.Contains(id, "/") {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "não encontrado"})
		return
	}
	switch r.Method {
	case http.MethodGet:
		e, ok := s.lib.Get(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "não encontrado"})
			return
		}
		writeJSON(w, http.StatusOK, e)
	case http.MethodPut:
		var e library.Entry
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "json inválido"})
			return
		}
		updated, err := s.lib.Update(id, e)
		if err != nil {
			if err.Error() == "não encontrado" {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, updated)
	case http.MethodDelete:
		if !s.lib.Delete(id) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "não encontrado"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
