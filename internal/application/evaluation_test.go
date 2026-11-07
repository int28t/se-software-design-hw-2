package application_test

import (
	"errors"
	"slices"
	"testing"

	"se-software-design-hw-2/internal/adapters/memory"
	"se-software-design-hw-2/internal/application"
	"se-software-design-hw-2/internal/domain"
)

func evaluationScenarioBook(t *testing.T) domain.Gradebook {
	t.Helper()
	var book domain.Gradebook
	if _, err := book.AddStudent("Анна"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Проект", "Экзамен"} {
		if _, err := book.AddCriterion(name, 10, 1); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := book.AddEvaluation(1, 1, 8); err != nil {
		t.Fatal(err)
	}
	return book
}

func TestEvaluationScenarios(t *testing.T) {
	repo := memory.NewRepository()
	if err := repo.Save(evaluationScenarioBook(t)); err != nil {
		t.Fatal(err)
	}
	service := application.NewService(repo, nil)
	if _, err := service.AddEvaluation(1, 2, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CorrectEvaluation(1, 1, 10); err != nil {
		t.Fatal(err)
	}
	got, err := service.Evaluations(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Score().Value() != 10 || got[1].Score().Value() != 0 {
		t.Fatalf("unexpected persisted grades: %#v", got)
	}
	if err := service.CancelEvaluation(1, 1); err != nil {
		t.Fatal(err)
	}
	got, err = service.Evaluations(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].CriterionID() != 2 || got[0].Score().Value() != 0 {
		t.Fatalf("cancellation was not saved: %#v", got)
	}
}

type evaluationRepository struct {
	book    domain.Gradebook
	loadErr error
	saveErr error
	saves   int
}

func (r *evaluationRepository) Load() (domain.Gradebook, error) {
	return r.book.Clone(), r.loadErr
}

func (r *evaluationRepository) Save(book domain.Gradebook) error {
	r.saves++
	if r.saveErr != nil {
		return r.saveErr
	}
	r.book = book.Clone()
	return nil
}

func TestEvaluationRepositoryErrors(t *testing.T) {
	wantErr := errors.New("storage failure")
	for _, operation := range []struct {
		name    string
		mutates bool
		run     func(*application.Service) error
	}{
		{"add", true, func(s *application.Service) error {
			_, err := s.AddEvaluation(1, 2, 5)
			return err
		}},
		{"list", false, func(s *application.Service) error {
			_, err := s.Evaluations(1)
			return err
		}},
		{"correct", true, func(s *application.Service) error {
			_, err := s.CorrectEvaluation(1, 1, 10)
			return err
		}},
		{"cancel", true, func(s *application.Service) error {
			return s.CancelEvaluation(1, 1)
		}},
	} {
		t.Run(operation.name+"/load", func(t *testing.T) {
			repo := &evaluationRepository{loadErr: wantErr}
			if err := operation.run(application.NewService(repo, nil)); !errors.Is(err, wantErr) {
				t.Fatalf("lost load error: %v", err)
			}
			if repo.saves != 0 {
				t.Fatal("saved after load error")
			}
		})
		if operation.mutates {
			t.Run(operation.name+"/save", func(t *testing.T) {
				repo := &evaluationRepository{book: evaluationScenarioBook(t), saveErr: wantErr}
				before, err := repo.book.Evaluations(1)
				if err != nil {
					t.Fatal(err)
				}
				if err := operation.run(application.NewService(repo, nil)); !errors.Is(err, wantErr) {
					t.Fatalf("lost save error: %v", err)
				}
				after, err := repo.book.Evaluations(1)
				if err != nil {
					t.Fatal(err)
				}
				if repo.saves != 1 || !slices.Equal(after, before) {
					t.Fatalf("save failure changed repository: saves=%d grades=%#v", repo.saves, after)
				}
			})
		}
	}
}

func TestInvalidEvaluationScenariosDoNotSave(t *testing.T) {
	repo := &evaluationRepository{book: evaluationScenarioBook(t)}
	service := application.NewService(repo, nil)
	operations := []func() error{
		func() error { _, err := service.AddEvaluation(1, 1, 5); return err },
		func() error { _, err := service.AddEvaluation(1, 2, 11); return err },
		func() error { _, err := service.AddEvaluation(99, 2, 5); return err },
		func() error { _, err := service.CorrectEvaluation(1, 1, -1); return err },
		func() error { _, err := service.CorrectEvaluation(1, 2, 5); return err },
		func() error { return service.CancelEvaluation(1, 2) },
		func() error { _, err := service.Evaluations(99); return err },
	}
	for i, operation := range operations {
		if err := operation(); err == nil {
			t.Fatalf("operation %d accepted invalid input", i)
		}
	}
	if repo.saves != 0 {
		t.Fatalf("invalid operation called Save %d times", repo.saves)
	}
}
