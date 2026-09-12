package library_test

import (
	"path/filepath"
	"testing"

	"mangasite-backend/internal/library"
)

func TestListEmptyInitially(t *testing.T) {
	s, err := library.New(filepath.Join(t.TempDir(), "lib.json"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := s.List(""); len(got) != 0 {
		t.Fatalf("List = %d, want 0", len(got))
	}
}

func TestAddAndGet(t *testing.T) {
	s, _ := library.New(filepath.Join(t.TempDir(), "lib.json"))
	e, err := s.Add(library.Entry{MangaID: "abc", Title: "Berserk", Status: "reading"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if e.MangaID != "abc" || e.Status != "reading" {
		t.Fatalf("Add = %+v", e)
	}
	got, ok := s.Get("abc")
	if !ok || got.Title != "Berserk" {
		t.Fatalf("Get = %+v, %v", got, ok)
	}
}

func TestAddValidates(t *testing.T) {
	s, _ := library.New(filepath.Join(t.TempDir(), "lib.json"))
	if _, err := s.Add(library.Entry{MangaID: "", Title: "X", Status: "reading"}); err == nil {
		t.Fatal("sem MangaID deveria falhar")
	}
	if _, err := s.Add(library.Entry{MangaID: "a", Title: "", Status: "reading"}); err == nil {
		t.Fatal("sem Title deveria falhar")
	}
	for _, bad := range []string{"", "lendo", "readding", "READ"} {
		if _, err := s.Add(library.Entry{MangaID: "m-" + bad, Title: "T", Status: bad}); err == nil {
			t.Fatalf("status %q deveria falhar", bad)
		}
	}
	if _, err := s.Add(library.Entry{MangaID: "a", Title: "A", Status: "plan to read"}); err != nil {
		t.Fatalf("status com espaço deveria normalizar: %v", err)
	}
	if _, err := s.Add(library.Entry{MangaID: "a", Title: "Duplicado", Status: "reading"}); err == nil {
		t.Fatal("duplicado deveria falhar")
	}
}

func TestUpdateAndDelete(t *testing.T) {
	s, _ := library.New(filepath.Join(t.TempDir(), "lib.json"))
	if _, err := s.Add(library.Entry{MangaID: "m1", Title: "One Piece", Status: "plan_to_read"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	upd, err := s.Update("m1", library.Entry{Status: "reading"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if upd.Status != "reading" {
		t.Fatalf("Status = %q", upd.Status)
	}
	if _, err := s.Update("m1", library.Entry{Status: "invalido"}); err == nil {
		t.Fatal("status inválido deveria falhar")
	}
	if _, err := s.Update("nope", library.Entry{Status: "reading"}); err == nil {
		t.Fatal("inexistente deveria falhar")
	}
	if got := s.List("reading"); len(got) != 1 {
		t.Fatalf("List(reading) = %d, want 1", len(got))
	}
	if !s.Delete("m1") {
		t.Fatal("Delete deveria retornar true")
	}
	if s.Delete("m1") {
		t.Fatal("Delete duplicado deveria retornar false")
	}
}

func TestPersistsToFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lib.json")
	s, _ := library.New(path)
	if _, err := s.Add(library.Entry{MangaID: "m9", Title: "Vagabond", Status: "finished"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	s2, err := library.New(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if _, ok := s2.Get("m9"); !ok {
		t.Fatal("m9 deveria persistir")
	}
}
