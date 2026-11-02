package domain_test

import (
	"math"
	"slices"
	"strconv"
	"testing"

	"se-software-design-hw-2/internal/domain"
)

func evaluationBook(t *testing.T) domain.Gradebook {
	t.Helper()
	var book domain.Gradebook
	for _, name := range []string{"Анна", "Борис"} {
		if _, err := book.AddStudent(name); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"Проект", "Экзамен"} {
		if _, err := book.AddCriterion(name, 10, 1); err != nil {
			t.Fatal(err)
		}
	}
	return book
}

func evaluations(t *testing.T, book domain.Gradebook, studentID domain.StudentID) []domain.Evaluation {
	t.Helper()
	got, err := book.Evaluations(studentID)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestEvaluationLifecycle(t *testing.T) {
	book := evaluationBook(t)
	if got := evaluations(t, book, 1); len(got) != 0 {
		t.Fatal("new student has evaluations")
	}
	for _, pair := range [][2]int{{1, 1}, {1, 2}, {2, 1}} {
		e, err := book.AddEvaluation(domain.StudentID(pair[0]), domain.CriterionID(pair[1]), 0)
		if err != nil {
			t.Fatal(err)
		}
		if int(e.StudentID()) != pair[0] || int(e.CriterionID()) != pair[1] || e.Score().Value() != 0 {
			t.Fatalf("unexpected evaluation: %#v", e)
		}
	}
	if got := evaluations(t, book, 1); len(got) != 2 || got[0].CriterionID() != 1 || got[1].CriterionID() != 2 {
		t.Fatalf("incorrect student filter or order: %#v", got)
	}
	if _, err := book.AddEvaluation(1, 1, 5); err == nil {
		t.Fatal("duplicate evaluation overwrote a legitimate zero")
	}
	corrected, err := book.CorrectEvaluation(1, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if corrected.StudentID() != 1 || corrected.CriterionID() != 1 || corrected.Score().Value() != 10 {
		t.Fatalf("incorrect corrected evaluation: %#v", corrected)
	}
	if got := evaluations(t, book, 1); len(got) != 2 || got[0] != corrected {
		t.Fatalf("correction changed identity or position: %#v", got)
	}
	if err := book.CancelEvaluation(1, 1); err != nil {
		t.Fatal(err)
	}
	if got := evaluations(t, book, 1); len(got) != 1 || got[0].CriterionID() != 2 {
		t.Fatalf("cancelled wrong evaluation: %#v", got)
	}
	if got := evaluations(t, book, 2); len(got) != 1 || got[0].Score().Value() != 0 {
		t.Fatalf("another student's evaluation changed: %#v", got)
	}
	if err := book.CancelEvaluation(1, 1); err == nil {
		t.Fatal("cancelled missing evaluation")
	}
	if _, err := book.CorrectEvaluation(1, 1, 5); err == nil {
		t.Fatal("correction created a missing evaluation")
	}
	if _, err := book.AddEvaluation(1, 1, 7.5); err != nil {
		t.Fatalf("cannot regrade after cancellation: %v", err)
	}
}

func TestInvalidEvaluationDoesNotChangeState(t *testing.T) {
	book := evaluationBook(t)
	if _, err := book.AddEvaluation(1, 1, 8); err != nil {
		t.Fatal(err)
	}
	before := evaluations(t, book, 1)
	for _, value := range []float64{-1, 10.01, math.NaN(), math.Inf(1), math.Inf(-1)} {
		t.Run(strconv.FormatFloat(value, 'g', -1, 64), func(t *testing.T) {
			if _, err := book.AddEvaluation(1, 2, value); err == nil {
				t.Fatal("accepted invalid score")
			}
			if _, err := book.CorrectEvaluation(1, 1, value); err == nil {
				t.Fatal("accepted invalid correction")
			}
			if got := evaluations(t, book, 1); !slices.Equal(got, before) {
				t.Fatalf("invalid operation changed evaluations: %#v", got)
			}
		})
	}
	for _, pair := range [][2]int{{0, 1}, {-1, 1}, {99, 1}, {1, 0}, {1, -1}, {1, 99}} {
		studentID, criterionID := domain.StudentID(pair[0]), domain.CriterionID(pair[1])
		if _, err := book.AddEvaluation(studentID, criterionID, 5); err == nil {
			t.Fatalf("accepted unknown reference: %v", pair)
		}
		if _, err := book.CorrectEvaluation(studentID, criterionID, 5); err == nil {
			t.Fatalf("corrected unknown reference: %v", pair)
		}
		if err := book.CancelEvaluation(studentID, criterionID); err == nil {
			t.Fatalf("cancelled unknown reference: %v", pair)
		}
	}
	if got := evaluations(t, book, 1); !slices.Equal(got, before) {
		t.Fatal("invalid reference changed state")
	}
	if _, err := book.Evaluations(99); err == nil {
		t.Fatal("unknown student treated as an empty list")
	}
}

func TestEvaluationCopiesAreIsolated(t *testing.T) {
	book := evaluationBook(t)
	if _, err := book.AddEvaluation(1, 1, 8); err != nil {
		t.Fatal(err)
	}
	if _, err := book.AddEvaluation(1, 2, 4); err != nil {
		t.Fatal(err)
	}
	list := evaluations(t, book, 1)
	list[0] = domain.Evaluation{}
	if evaluations(t, book, 1)[0].Score().Value() != 8 {
		t.Fatal("list exposes mutable state")
	}
	clone := book.Clone()
	if _, err := clone.CorrectEvaluation(1, 1, 3); err != nil {
		t.Fatal(err)
	}
	if err := clone.CancelEvaluation(1, 2); err != nil {
		t.Fatal(err)
	}
	if got := evaluations(t, book, 1); len(got) != 2 || got[0].Score().Value() != 8 || got[1].Score().Value() != 4 {
		t.Fatalf("clone correction/cancellation mutated original: %#v", got)
	}
}
