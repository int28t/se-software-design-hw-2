package domain_test

import (
	"math"
	"testing"

	"se-software-design-hw-2/internal/domain"
)

func totalBook(t *testing.T) domain.Gradebook {
	t.Helper()
	var book domain.Gradebook
	for _, name := range []string{"Анна", "Борис"} {
		if _, err := book.AddStudent(name); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := book.AddCriterion("Проект", 10, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := book.AddCriterion("Экзамен", 20, 3); err != nil {
		t.Fatal(err)
	}
	return book
}

func assertTotal(t *testing.T, calculator *domain.TotalCalculator, book domain.Gradebook, studentID domain.StudentID, want float64, graded, criteria int, preliminary bool) {
	t.Helper()
	got, err := calculator.Calculate(book, studentID)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got.Value()-want) > 1e-12 || math.IsNaN(got.Value()) || math.IsInf(got.Value(), 0) {
		t.Fatalf("total = %g, want %g", got.Value(), want)
	}
	if got.StudentID() != studentID || got.GradedCount() != graded || got.CriterionCount() != criteria || got.Preliminary() != preliminary || got.Complete() != (graded == criteria) {
		t.Fatalf("unexpected total metadata: %#v", got)
	}
}

func TestWeightedTotalStrategies(t *testing.T) {
	book := totalBook(t)
	final := domain.NewTotalCalculator(domain.FinalWeightedTotal{})
	preliminary := domain.NewTotalCalculator(domain.PreliminaryWeightedTotal{})
	assertTotal(t, preliminary, book, 1, 0, 0, 2, true)
	if _, err := final.Calculate(book, 1); err == nil {
		t.Fatal("final total accepted missing grades")
	}
	if _, err := book.AddEvaluation(1, 1, 8); err != nil {
		t.Fatal(err)
	}
	assertTotal(t, preliminary, book, 1, 2, 1, 2, true)
	if _, err := final.Calculate(book, 1); err == nil {
		t.Fatal("final total accepted a partially graded rubric")
	}
	if _, err := book.AddEvaluation(1, 2, 10); err != nil {
		t.Fatal(err)
	}
	assertTotal(t, final, book, 1, 5.75, 2, 2, false)
	assertTotal(t, preliminary, book, 1, 5.75, 2, 2, true)
	// Another student's grades must not participate in the calculation.
	if _, err := book.AddEvaluation(2, 1, 10); err != nil {
		t.Fatal(err)
	}
	if _, err := book.AddEvaluation(2, 2, 20); err != nil {
		t.Fatal(err)
	}
	assertTotal(t, final, book, 2, 10, 2, 2, false)
	assertTotal(t, final, book, 1, 5.75, 2, 2, false)
}

func TestTotalUsesCurrentEvaluationsAndRubric(t *testing.T) {
	book := totalBook(t)
	for _, grade := range []struct {
		criterion domain.CriterionID
		score     float64
	}{{1, 8}, {2, 10}} {
		if _, err := book.AddEvaluation(1, grade.criterion, grade.score); err != nil {
			t.Fatal(err)
		}
	}
	final := domain.NewTotalCalculator(domain.FinalWeightedTotal{})
	preliminary := domain.NewTotalCalculator(domain.PreliminaryWeightedTotal{})
	assertTotal(t, final, book, 1, 5.75, 2, 2, false)
	if _, err := book.CorrectEvaluation(1, 1, 10); err != nil {
		t.Fatal(err)
	}
	assertTotal(t, final, book, 1, 6.25, 2, 2, false)
	if err := book.CancelEvaluation(1, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := final.Calculate(book, 1); err == nil {
		t.Fatal("cancelled grade remained in the final total")
	}
	assertTotal(t, preliminary, book, 1, 2.5, 1, 2, true)
	if _, err := book.AddEvaluation(1, 2, 20); err != nil {
		t.Fatal(err)
	}
	assertTotal(t, final, book, 1, 10, 2, 2, false)
	if _, err := book.AddCriterion("Активность", 5, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := final.Calculate(book, 1); err == nil {
		t.Fatal("new criterion was ignored")
	}
	assertTotal(t, preliminary, book, 1, 8, 2, 3, true)
}

func TestTotalZeroIsACompletedGrade(t *testing.T) {
	book := totalBook(t)
	for _, id := range []domain.CriterionID{1, 2} {
		if _, err := book.AddEvaluation(1, id, 0); err != nil {
			t.Fatal(err)
		}
	}
	assertTotal(t, domain.NewTotalCalculator(domain.FinalWeightedTotal{}), book, 1, 0, 2, 2, false)
}

func TestTotalRejectsEmptyRubricAndUnknownStudent(t *testing.T) {
	var empty domain.Gradebook
	if _, err := empty.AddStudent("Анна"); err != nil {
		t.Fatal(err)
	}
	for _, strategy := range []domain.TotalStrategy{domain.FinalWeightedTotal{}, domain.PreliminaryWeightedTotal{}} {
		calculator := domain.NewTotalCalculator(strategy)
		if _, err := calculator.Calculate(empty, 1); err == nil {
			t.Fatal("accepted an empty rubric")
		}
		if _, err := calculator.Calculate(totalBook(t), 99); err == nil {
			t.Fatal("accepted an unknown student")
		}
	}
	if _, err := domain.NewTotalCalculator(nil).Calculate(totalBook(t), 1); err == nil {
		t.Fatal("accepted missing strategy")
	}
}

func TestTotalDoesNotOverflowOrRoundIntermediateValues(t *testing.T) {
	for _, weight := range []float64{1, math.MaxFloat64, math.SmallestNonzeroFloat64} {
		var book domain.Gradebook
		if _, err := book.AddStudent("Анна"); err != nil {
			t.Fatal(err)
		}
		for range 2 {
			criterion, err := book.AddCriterion("Критерий", 3, weight)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := book.AddEvaluation(1, criterion.ID(), 1); err != nil {
				t.Fatal(err)
			}
		}
		assertTotal(t, domain.NewTotalCalculator(domain.FinalWeightedTotal{}), book, 1, 10.0/3, 2, 2, false)
	}
}
