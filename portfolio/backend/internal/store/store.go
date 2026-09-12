package store

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"sync"
)

// Project é o registro de um projeto do portfólio.
type Project struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tech        []string `json:"tech"`
	URL         string   `json:"url,omitempty"`
	Repo        string   `json:"repo,omitempty"`
	Featured    bool     `json:"featured"`
}

// Store guarda projetos em memória com persistência em JSON.
type Store struct {
	mu    sync.RWMutex
	path  string
	items map[string]Project
}

// New carrega do arquivo path (se existir) ou inicia vazio.
func New(path string) (*Store, error) {
	s := &Store{path: path, items: map[string]Project{}}
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
	var list []Project
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	for _, p := range list {
		s.items[p.ID] = p
	}
	return s, nil
}

func (s *Store) save() error {
	if s.path == "" {
		return nil
	}
	list := s.listLocked()
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

func (s *Store) listLocked() []Project {
	list := make([]Project, 0, len(s.items))
	for _, p := range s.items {
		list = append(list, p)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	return list
}

// List retorna todos os projetos ordenados por ID.
func (s *Store) List() []Project {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.listLocked()
}

// Get busca um projeto por ID.
func (s *Store) Get(id string) (Project, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.items[id]
	return p, ok
}

// Create valida e insere um novo projeto.
func (s *Store) Create(p Project) (Project, error) {
	if p.ID == "" {
		return Project{}, errors.New("id é obrigatório")
	}
	if p.Title == "" {
		return Project{}, errors.New("title é obrigatório")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.items[p.ID]; exists {
		return Project{}, errors.New("id já existe")
	}
	if p.Tech == nil {
		p.Tech = []string{}
	}
	s.items[p.ID] = p
	if err := s.save(); err != nil {
		delete(s.items, p.ID)
		return Project{}, err
	}
	return p, nil
}

// Update altera um projeto existente (ID da URL prevalece).
func (s *Store) Update(id string, p Project) (Project, error) {
	if p.Title == "" {
		return Project{}, errors.New("title é obrigatório")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cur, ok := s.items[id]
	if !ok {
		return Project{}, errors.New("não encontrado")
	}
	cur.Title = p.Title
	cur.Description = p.Description
	if p.Tech != nil {
		cur.Tech = p.Tech
	}
	if p.URL != "" {
		cur.URL = p.URL
	}
	if p.Repo != "" {
		cur.Repo = p.Repo
	}
	cur.Featured = p.Featured
	s.items[id] = cur
	if err := s.save(); err != nil {
		return Project{}, err
	}
	return cur, nil
}

// Delete remove um projeto; retorna false se não existia.
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

// SeedIfEmpty insere os projetos iniciais quando o store está vazio.
func (s *Store) SeedIfEmpty() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.items) > 0 {
		return
	}
	samples := []Project{
		{
			ID:          "site-investimentos",
			Title:       "Site de Investimentos",
			Description: "Acompanhamento de carteiras, aportes e rentabilidade.",
			Tech:        []string{"Go", "React"},
			Featured:    true,
		},
		{
			ID:          "site-manga",
			Title:       "Site de Mangá",
			Description: "Catálogo e leitura com favoritos e progresso.",
			Tech:        []string{"React", "Go"},
			Featured:    true,
		},
	}
	for _, p := range samples {
		s.items[p.ID] = p
	}
	_ = s.save()
}
