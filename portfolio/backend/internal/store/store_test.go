package store_test

import (
	"path/filepath"
	"testing"

	"portfolio-backend/internal/store"
)

func TestListEmptyInitially(t *testing.T) {
	s, err := store.New(filepath.Join(t.TempDir(), "data.json"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := s.List(); len(got) != 0 {
		t.Fatalf("List = %d itens, want 0", len(got))
	}
}

func TestCreateAndGet(t *testing.T) {
	s, _ := store.New(filepath.Join(t.TempDir(), "data.json"))
	p, err := s.Create(store.Project{
		ID:          "site-investimentos",
		Title:       "Site de Investimentos",
		Description: "Carteiras e rentabilidade",
		Tech:        []string{"Go", "React"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.ID != "site-investimentos" {
		t.Fatalf("ID = %q", p.ID)
	}
	got, ok := s.Get("site-investimentos")
	if !ok || got.Title != "Site de Investimentos" {
		t.Fatalf("Get = %+v, %v", got, ok)
	}
}

func TestCreateValidates(t *testing.T) {
	s, _ := store.New(filepath.Join(t.TempDir(), "data.json"))
	if _, err := s.Create(store.Project{ID: "", Title: "Sem ID"}); err == nil {
		t.Fatal("Create sem ID deveria falhar")
	}
	if _, err := s.Create(store.Project{ID: "x", Title: ""}); err == nil {
		t.Fatal("Create sem Title deveria falhar")
	}
	if _, err := s.Create(store.Project{ID: "a", Title: "A"}); err != nil {
		t.Fatalf("Create A: %v", err)
	}
	if _, err := s.Create(store.Project{ID: "a", Title: "Duplicado"}); err == nil {
		t.Fatal("Create duplicado deveria falhar")
	}
}

func TestUpdateAndDelete(t *testing.T) {
	s, _ := store.New(filepath.Join(t.TempDir(), "data.json"))
	if _, err := s.Create(store.Project{ID: "site-manga", Title: "Site de Mangá"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	upd, err := s.Update("site-manga", store.Project{Title: "Mangá Atualizado", Description: "Leitura"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if upd.Title != "Mangá Atualizado" {
		t.Fatalf("Title = %q", upd.Title)
	}
	if _, err := s.Update("inexistente", store.Project{Title: "X"}); err == nil {
		t.Fatal("Update inexistente deveria falhar")
	}
	if !s.Delete("site-manga") {
		t.Fatal("Delete deveria retornar true")
	}
	if _, ok := s.Get("site-manga"); ok {
		t.Fatal("Get após Delete deveria falhar")
	}
	if s.Delete("site-manga") {
		t.Fatal("Delete duplicado deveria retornar false")
	}
}

func TestPersistsToFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	s, _ := store.New(path)
	if _, err := s.Create(store.Project{ID: "p1", Title: "P1"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	s2, err := store.New(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if _, ok := s2.Get("p1"); !ok {
		t.Fatal("projeto p1 deveria persistir no arquivo")
	}
}
