package memory_test

import (
	"testing"

	"se-software-design-hw-2/internal/adapters/memory"
	"se-software-design-hw-2/internal/domain"
)

func TestEvaluationSnapshots(t *testing.T) {
	var book domain.Gradebook
	if _, err := book.AddStudent("Анна"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Проект", "Экзамен"} {
		c, err := book.AddCriterion(name, 10, 1)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := book.AddEvaluation(1, c.ID(), 8); err != nil {
			t.Fatal(err)
		}
	}
	repo := memory.NewRepository()
	if err := repo.Save(book); err != nil {
		t.Fatal(err)
	}
	checkOriginal := func() {
		t.Helper()
		loaded, err := repo.Load()
		if err != nil {
			t.Fatal(err)
		}
		got, err := loaded.Evaluations(1)
		if err != nil || len(got) != 2 || got[0].Score().Value() != 8 || got[1].Score().Value() != 8 {
			t.Fatalf("unsaved mutation reached repository: %#v, %v", got, err)
		}
	}
	if _, err := book.CorrectEvaluation(1, 1, 3); err != nil {
		t.Fatal(err)
	}
	if err := book.CancelEvaluation(1, 2); err != nil {
		t.Fatal(err)
	}
	checkOriginal()
	loaded, err := repo.Load()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := loaded.CorrectEvaluation(1, 1, 5); err != nil {
		t.Fatal(err)
	}
	if err := loaded.CancelEvaluation(1, 2); err != nil {
		t.Fatal(err)
	}
	checkOriginal()
	if err := repo.Save(loaded); err != nil {
		t.Fatal(err)
	}
	fresh, err := repo.Load()
	if err != nil {
		t.Fatal(err)
	}
	got, err := fresh.Evaluations(1)
	if err != nil || len(got) != 1 || got[0].Score().Value() != 5 {
		t.Fatalf("explicit save did not persist correction/cancellation: %#v, %v", got, err)
	}
}
