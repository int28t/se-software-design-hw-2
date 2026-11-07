package domain_test

import (
	"errors"
	"math"
	"slices"
	"testing"

	"se-software-design-hw-2/internal/domain"
)

func TestFinalTotalReportsIncompleteCounts(t *testing.T) {
	book := totalBook(t)
	if _, err := book.AddEvaluation(1, 1, 0); err != nil {
		t.Fatal(err)
	}
	_, err := domain.NewTotalCalculator(domain.FinalWeightedTotal{}).Calculate(book, 1)
	incomplete, ok := errors.AsType[*domain.IncompleteEvaluationsError](err)
	if !ok || incomplete.Graded != 1 || incomplete.Criteria != 2 {
		t.Fatalf("missing typed completeness information: %v", err)
	}
}

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

type configuredTotalStrategy struct {
	calculation domain.TotalCalculation
	err         error
}

func (s configuredTotalStrategy) Calculate([]domain.Criterion, []domain.Evaluation) (domain.TotalCalculation, error) {
	return s.calculation, s.err
}

func TestTotalCalculatorDerivesMetadataFromGradebook(t *testing.T) {
	book := totalBook(t)
	for _, criterion := range book.Criteria() {
		if _, err := book.AddEvaluation(1, criterion.ID(), 0); err != nil {
			t.Fatal(err)
		}
	}
	calculator := domain.NewTotalCalculator(configuredTotalStrategy{})
	assertTotal(t, calculator, book, 1, 0, 2, 2, false)
}

type sumRatioStrategy struct{}

func (sumRatioStrategy) Calculate(
	criteria []domain.Criterion,
	evaluations []domain.Evaluation,
) (domain.TotalCalculation, error) {
	var score, maximum float64
	for _, evaluation := range evaluations {
		score += evaluation.Score().Value()
	}
	for _, criterion := range criteria {
		maximum += criterion.MaxScore()
	}
	return domain.TotalCalculation{Value: 10 * score / maximum}, nil
}

func TestTotalStrategyCanDefineANewFormulaOutsideDomain(t *testing.T) {
	book := totalBook(t)
	for _, grade := range []struct {
		criterion domain.CriterionID
		score     float64
	}{{1, 8}, {2, 10}} {
		if _, err := book.AddEvaluation(1, grade.criterion, grade.score); err != nil {
			t.Fatal(err)
		}
	}
	calculator := domain.NewTotalCalculator(sumRatioStrategy{})
	assertTotal(t, calculator, book, 1, 6, 2, 2, false)
}

func TestTotalCalculatorRejectsInvalidStrategyValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value float64
	}{
		{name: "negative", value: -0.01},
		{name: "above scale", value: 10.01},
		{name: "NaN", value: math.NaN()},
		{name: "positive infinity", value: math.Inf(1)},
		{name: "negative infinity", value: math.Inf(-1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			strategy := configuredTotalStrategy{
				calculation: domain.TotalCalculation{Value: tc.value, Preliminary: true},
			}
			calculator := domain.NewTotalCalculator(strategy)
			if _, err := calculator.Calculate(totalBook(t), 1); err == nil {
				t.Fatalf("accepted an invalid strategy result: %g", tc.value)
			}
		})
	}
}

func TestTotalCalculatorEnforcesFinalCompletenessForCustomStrategy(t *testing.T) {
	strategy := configuredTotalStrategy{calculation: domain.TotalCalculation{Value: 7}}
	_, err := domain.NewTotalCalculator(strategy).Calculate(totalBook(t), 1)
	incomplete, ok := errors.AsType[*domain.IncompleteEvaluationsError](err)
	if !ok || incomplete.Graded != 0 || incomplete.Criteria != 2 {
		t.Fatalf("custom strategy bypassed final completeness: %v", err)
	}
}

func TestTotalCalculatorRejectsEmptyRubricForCustomStrategy(t *testing.T) {
	var book domain.Gradebook
	if _, err := book.AddStudent("Анна"); err != nil {
		t.Fatal(err)
	}
	strategy := configuredTotalStrategy{calculation: domain.TotalCalculation{Value: 7, Preliminary: true}}
	if _, err := domain.NewTotalCalculator(strategy).Calculate(book, 1); err == nil {
		t.Fatal("custom strategy bypassed empty rubric validation")
	}
}

func TestTotalCalculatorPropagatesStrategyFailure(t *testing.T) {
	wantErr := errors.New("calculation failure")
	strategy := configuredTotalStrategy{err: wantErr}
	_, err := domain.NewTotalCalculator(strategy).Calculate(totalBook(t), 1)
	if !errors.Is(err, wantErr) {
		t.Fatalf("lost strategy failure: %v", err)
	}
}

type clearingTotalStrategy struct{}

func (clearingTotalStrategy) Calculate(
	criteria []domain.Criterion,
	evaluations []domain.Evaluation,
) (domain.TotalCalculation, error) {
	clear(criteria)
	clear(evaluations)
	return domain.TotalCalculation{Preliminary: true}, nil
}

func TestTotalStrategyCannotMutateGradebook(t *testing.T) {
	book := totalBook(t)
	if _, err := book.AddEvaluation(1, 1, 8); err != nil {
		t.Fatal(err)
	}
	beforeCriteria := book.Criteria()
	beforeEvaluations, err := book.Evaluations(1)
	if err != nil {
		t.Fatal(err)
	}
	calculator := domain.NewTotalCalculator(clearingTotalStrategy{})
	assertTotal(t, calculator, book, 1, 0, 1, 2, true)
	afterEvaluations, err := book.Evaluations(1)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(beforeCriteria, book.Criteria()) || !slices.Equal(beforeEvaluations, afterEvaluations) {
		t.Fatal("strategy changed the aggregate through its input collections")
	}
}
