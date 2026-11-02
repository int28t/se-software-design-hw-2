package memory_test

import (
	"testing"

	"se-software-design-hw-2/internal/adapters/memory"
	"se-software-design-hw-2/internal/domain"
)

func TestRepositorySnapshotIsolation(t *testing.T) {
	repo := memory.NewRepository()
	var book domain.Gradebook
	for range 5 {
		if _, err := book.AddStudent("Студент"); err != nil {
			t.Fatal(err)
		}
		if _, err := book.AddCriterion("Критерий", 10, 1); err != nil {
			t.Fatal(err)
		}
	}
	if err := repo.Save(book); err != nil {
		t.Fatal(err)
	}
	// A change after Save must not share the repository's backing arrays.
	if _, err := book.AddStudent("После Save"); err != nil {
		t.Fatal(err)
	}
	if _, err := book.AddCriterion("После Save", 20, 3); err != nil {
		t.Fatal(err)
	}
	loaded, err := repo.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Students()) != 5 || len(loaded.Criteria()) != 5 {
		t.Fatal("Save retained a reference to caller's book")
	}
	if _, err := loaded.AddStudent("После Load"); err != nil {
		t.Fatal(err)
	}
	if _, err := loaded.AddCriterion("После Load", 5, 2); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(loaded); err != nil {
		t.Fatal(err)
	}
	again, err := repo.Load()
	if err != nil {
		t.Fatal(err)
	}
	if again.Students()[5].Name() != "После Load" || again.Criteria()[5].Name() != "После Load" {
		t.Fatal("loaded snapshot was overwritten through shared backing arrays")
	}
	if _, err := loaded.AddStudent("Не сохранено"); err != nil {
		t.Fatal(err)
	}
	fresh, err := repo.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(fresh.Students()) != 6 {
		t.Fatal("Load leaked unsaved changes")
	}
}
