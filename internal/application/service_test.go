package application_test

import (
	"errors"
	"testing"

	"se-software-design-hw-2/internal/adapters/memory"
	"se-software-design-hw-2/internal/application"
	"se-software-design-hw-2/internal/domain"
)

func TestStudentAndCriterionScenarios(t *testing.T) {
	s := application.NewService(memory.NewRepository())
	for _, name := range []string{"Анна", "Борис"} {
		if _, err := s.AddStudent(name); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.AddCriterion("Проект", 20, 3); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddCriterion("Экзамен", 10, 1); err != nil {
		t.Fatal(err)
	}
	students, err := s.Students()
	if err != nil {
		t.Fatal(err)
	}
	criteria, err := s.Criteria()
	if err != nil {
		t.Fatal(err)
	}
	if len(students) != 2 || students[0].Name() != "Анна" || students[1].ID() != 2 {
		t.Fatalf("unexpected students: %#v", students)
	}
	if len(criteria) != 2 || criteria[0].Name() != "Проект" || criteria[1].ID() != 2 {
		t.Fatalf("unexpected criteria: %#v", criteria)
	}
}

type failingRepository struct {
	loadErr error
	saveErr error
	saves   int
}

func (r *failingRepository) Load() (domain.Gradebook, error) {
	return domain.Gradebook{}, r.loadErr
}

func (r *failingRepository) Save(domain.Gradebook) error {
	r.saves++
	return r.saveErr
}

func TestInvalidInputDoesNotSave(t *testing.T) {
	repo := &failingRepository{}
	s := application.NewService(repo)
	if _, err := s.AddStudent(" "); err == nil {
		t.Fatal("accepted empty name")
	}
	if _, err := s.AddCriterion("Экзамен", 0, 1); err == nil {
		t.Fatal("accepted zero maximum")
	}
	if repo.saves != 0 {
		t.Fatal("saved invalid changes")
	}
}

func TestRepositoryErrors(t *testing.T) {
	wantErr := errors.New("storage failure")
	for _, operation := range []struct {
		name string
		run  func(*application.Service) error
	}{
		{"add student", func(s *application.Service) error { _, err := s.AddStudent("Анна"); return err }},
		{"list students", func(s *application.Service) error { _, err := s.Students(); return err }},
		{"add criterion", func(s *application.Service) error { _, err := s.AddCriterion("Проект", 10, 1); return err }},
		{"list criteria", func(s *application.Service) error { _, err := s.Criteria(); return err }},
	} {
		t.Run(operation.name, func(t *testing.T) {
			repo := &failingRepository{loadErr: wantErr}
			if err := operation.run(application.NewService(repo)); !errors.Is(err, wantErr) {
				t.Fatalf("lost load error: %v", err)
			}
			if repo.saves != 0 {
				t.Fatal("saved after load error")
			}
		})
	}
	t.Run("save student", func(t *testing.T) {
		repo := &failingRepository{saveErr: wantErr}
		student, err := application.NewService(repo).AddStudent("Анна")
		if !errors.Is(err, wantErr) || student.ID() != 0 || repo.saves != 1 {
			t.Fatalf("save failure reported as success: student=%#v err=%v saves=%d", student, err, repo.saves)
		}
	})
	t.Run("save criterion", func(t *testing.T) {
		repo := &failingRepository{saveErr: wantErr}
		criterion, err := application.NewService(repo).AddCriterion("Проект", 10, 1)
		if !errors.Is(err, wantErr) || criterion.ID() != 0 || repo.saves != 1 {
			t.Fatalf("save failure reported as success: criterion=%#v err=%v saves=%d", criterion, err, repo.saves)
		}
	})
}
