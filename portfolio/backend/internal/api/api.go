package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"portfolio-backend/internal/store"
)

// Server expõe o portfólio via HTTP.
type Server struct {
	store *store.Store
	token string
}

// NewServer cria o handler com store e token admin para escrita.
func NewServer(s *store.Store, adminToken string) *Server {
	return &Server{store: s, token: adminToken}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Routes registra as rotas da API.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/projects", s.handleProjects)
	mux.HandleFunc("/api/projects/", s.handleProjectByID)
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

func (s *Server) requireAuth(r *http.Request) bool {
	if s.token == "" {
		return false
	}
	auth := r.Header.Get("Authorization")
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return false
	}
	return parts[1] == s.token
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.store.List())
	case http.MethodPost:
		if !s.requireAuth(r) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autorizado"})
			return
		}
		var p store.Project
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "json inválido"})
			return
		}
		created, err := s.store.Create(p)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleProjectByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/projects/")
	if id == "" || strings.Contains(id, "/") {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "não encontrado"})
		return
	}
	switch r.Method {
	case http.MethodGet:
		p, ok := s.store.Get(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "não encontrado"})
			return
		}
		writeJSON(w, http.StatusOK, p)
	case http.MethodPut:
		if !s.requireAuth(r) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autorizado"})
			return
		}
		var p store.Project
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "json inválido"})
			return
		}
		updated, err := s.store.Update(id, p)
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
		if !s.requireAuth(r) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autorizado"})
			return
		}
		if !s.store.Delete(id) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "não encontrado"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
