package library

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// Status válidos da biblioteca.
const (
	StatusReading   = "reading"
	StatusPlan      = "plan_to_read"
	StatusFinished  = "finished"
	StatusDropped   = "dropped"
)

// Entry é um mangá salvo na biblioteca.
type Entry struct {
	MangaID   string `json:"mangaId"`
	Title     string `json:"title"`
	CoverURL  string `json:"coverUrl,omitempty"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// NormalizeStatus aceita "plan to read"/maiúsculas e devolve a forma canônica.
func NormalizeStatus(s string) (string, error) {
	n := strings.ToLower(strings.TrimSpace(s))
	n = strings.ReplaceAll(n, " ", "_")
	switch n {
	case StatusReading, StatusPlan, StatusFinished, StatusDropped:
		return n, nil
	default:
		return "", errors.New("status inválido: use reading, plan_to_read, finished ou dropped")
	}
}

// Store guarda a biblioteca em memória com persistência em JSON.
type Store struct {
	mu    sync.RWMutex
	path  string
	items map[string]Entry
}

// New carrega do arquivo path (se existir) ou inicia vazio.
func New(path string) (*Store, error) {
	s := &Store{path: path, items: map[string]Entry{}}
	if path == "" {
		return s, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return s, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return s, nil
	}
	var list []Entry
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	for _, e := range list {
		s.items[e.MangaID] = e
	}
	return s, nil
}

func (s *Store) save() error {
	if s.path == "" {
		return nil
	}
	list := s.listLocked("")
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) listLocked(filter string) []Entry {
	list := make([]Entry, 0, len(s.items))
	for _, e := range s.items {
		if filter == "" || e.Status == filter {
			list = append(list, e)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Title < list[j].Title })
	return list
}

// List retorna entradas, opcionalmente filtradas por status.
func (s *Store) List(filter string) []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.listLocked(filter)
}

// Get busca uma entrada pelo MangaID.
func (s *Store) Get(id string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.items[id]
	return e, ok
}

// Add valida e insere uma nova entrada.
func (s *Store) Add(e Entry) (Entry, error) {
	if e.MangaID == "" {
		return Entry{}, errors.New("mangaId é obrigatório")
	}
	if strings.TrimSpace(e.Title) == "" {
		return Entry{}, errors.New("title é obrigatório")
	}
	st, err := NormalizeStatus(e.Status)
	if err != nil {
		return Entry{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.items[e.MangaID]; exists {
		return Entry{}, errors.New("mangá já está na biblioteca")
	}
	e.Status = st
	e.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	s.items[e.MangaID] = e
	if err := s.save(); err != nil {
		delete(s.items, e.MangaID)
		return Entry{}, err
	}
	return e, nil
}

// Update altera status (e metadados opcionais) de uma entrada existente.
func (s *Store) Update(id string, e Entry) (Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur, ok := s.items[id]
	if !ok {
		return Entry{}, errors.New("não encontrado")
	}
	if e.Status != "" {
		st, err := NormalizeStatus(e.Status)
		if err != nil {
			return Entry{}, err
		}
		cur.Status = st
	}
	if strings.TrimSpace(e.Title) != "" {
		cur.Title = e.Title
	}
	if e.CoverURL != "" {
		cur.CoverURL = e.CoverURL
	}
	cur.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	s.items[id] = cur
	if err := s.save(); err != nil {
		return Entry{}, err
	}
	return cur, nil
}

// Delete remove uma entrada; retorna false se não existia.
func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	_ = s.save()
	return true
}
