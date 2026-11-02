package application_test

import (
	"errors"
	"math"
	"testing"

	"se-software-design-hw-2/internal/adapters/memory"
	"se-software-design-hw-2/internal/application"
	"se-software-design-hw-2/internal/domain"
)

func checkApplicationTotal(t *testing.T, service *application.Service, want float64, graded, criteria int) {
	t.Helper()
	got, err := service.PreliminaryTotal(1)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got.Value()-want) > 1e-12 || got.GradedCount() != graded || got.CriterionCount() != criteria || !got.Preliminary() {
		t.Fatalf("unexpected current preliminary total: %#v, want %g", got, want)
	}
	final, err := service.Total(1)
	if graded != criteria {
		if err == nil {
			t.Fatal("incomplete final total was accepted")
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(final.Value()-want) > 1e-12 || final.Preliminary() || !final.Complete() {
		t.Fatalf("unexpected final total: %#v, want %g", final, want)
	}
}

func TestApplicationTotalReadsCurrentState(t *testing.T) {
	repo := memory.NewRepository()
	if err := repo.Save(evaluationScenarioBook(t)); err != nil {
		t.Fatal(err)
	}
	service := application.NewService(repo)
	checkApplicationTotal(t, service, 4, 1, 2)
	if _, err := service.AddEvaluation(1, 2, 6); err != nil {
		t.Fatal(err)
	}
	checkApplicationTotal(t, service, 7, 2, 2)
	if _, err := service.CorrectEvaluation(1, 1, 10); err != nil {
		t.Fatal(err)
	}
	checkApplicationTotal(t, service, 8, 2, 2)
	if err := service.CancelEvaluation(1, 2); err != nil {
		t.Fatal(err)
	}
	checkApplicationTotal(t, service, 5, 1, 2)
	if _, err := service.AddCriterion("Активность", 20, 2); err != nil {
		t.Fatal(err)
	}
	checkApplicationTotal(t, service, 2.5, 1, 3)
}

func TestTotalScenariosPropagateErrorsAndNeverSave(t *testing.T) {
	wantErr := errors.New("load failure")
	for _, preliminary := range []bool{false, true} {
		run := func(s *application.Service, studentID domain.StudentID) (domain.Total, error) {
			if preliminary {
				return s.PreliminaryTotal(studentID)
			}
			return s.Total(studentID)
		}
		repo := &evaluationRepository{loadErr: wantErr}
		if _, err := run(application.NewService(repo), 1); !errors.Is(err, wantErr) {
			t.Fatalf("lost repository error: %v", err)
		}
		if repo.saves != 0 {
			t.Fatal("calculation saved after load failure")
		}
		repo = &evaluationRepository{book: evaluationScenarioBook(t)}
		service := application.NewService(repo)
		if _, err := run(service, 99); err == nil {
			t.Fatal("accepted unknown student")
		}
		if _, err := run(service, 1); (err == nil) != preliminary {
			t.Fatalf("wrong incomplete rubric behavior: preliminary=%v err=%v", preliminary, err)
		}
		if _, err := repo.book.AddEvaluation(1, 2, 6); err != nil {
			t.Fatal(err)
		}
		if _, err := run(service, 1); err != nil {
			t.Fatal(err)
		}
		if repo.saves != 0 {
			t.Fatal("read-only total scenario called Save")
		}
	}
}
