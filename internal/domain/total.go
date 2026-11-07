package domain

import (
	"errors"
	"fmt"
	"math"
)

// Total is a calculated snapshot, never an independently stored grade.
type Total struct {
	studentID   StudentID
	value       float64
	graded      int
	criteria    int
	preliminary bool
}

// StudentID returns the identity of the student whose total was calculated.
func (t Total) StudentID() StudentID { return t.studentID }

// Value returns the total on the 0–10 scale without display rounding.
func (t Total) Value() float64 { return t.value }

// GradedCount returns the number of evaluations included in the calculation.
func (t Total) GradedCount() int { return t.graded }

// CriterionCount returns the number of rubric items at calculation time.
func (t Total) CriterionCount() int { return t.criteria }

// Complete reports whether every item in a nonempty rubric was graded.
func (t Total) Complete() bool { return t.criteria > 0 && t.graded == t.criteria }

// Preliminary reports whether the total is preliminary rather than final.
func (t Total) Preliminary() bool { return t.preliminary }

// TotalCalculation carries the strategy output before the calculator validates it.
// Strategies supply the value and mode; student identity and counts belong to the calculator.
type TotalCalculation struct {
	Value       float64
	Preliminary bool
}

// TotalStrategy receives the actual rubric and one student's evaluations
// validated by the aggregate. It returns a value on the 0–10 scale and its mode.
// The input collections are independent snapshots of validated aggregate data.
type TotalStrategy interface {
	Calculate([]Criterion, []Evaluation) (TotalCalculation, error)
}

// TotalCalculator is a Domain Service: it combines the rubric with current grades.
type TotalCalculator struct {
	strategy TotalStrategy
}

// NewTotalCalculator creates a calculator using the supplied total strategy.
// The strategy must be a non-nil implementation.
func NewTotalCalculator(strategy TotalStrategy) *TotalCalculator {
	return &TotalCalculator{strategy: strategy}
}

// Calculate derives a total from the current rubric and evaluations without changing the book.
func (c *TotalCalculator) Calculate(book Gradebook, studentID StudentID) (Total, error) {
	if c.strategy == nil {
		return Total{}, errors.New("не задана стратегия расчёта итога")
	}
	evaluations, err := book.Evaluations(studentID)
	if err != nil {
		return Total{}, err
	}
	criteria := book.Criteria()
	if len(criteria) == 0 {
		return Total{}, errors.New("нельзя рассчитать итог: рубрикатор пуст")
	}
	calculation, err := c.strategy.Calculate(criteria, evaluations)
	if err != nil {
		return Total{}, err
	}
	nonFinite := math.IsNaN(calculation.Value) || math.IsInf(calculation.Value, 0)
	outsideScale := calculation.Value < 0 || calculation.Value > 10
	if nonFinite || outsideScale {
		return Total{}, errors.New("стратегия должна возвращать конечный итог в диапазоне от 0 до 10")
	}
	if !calculation.Preliminary && len(evaluations) != len(criteria) {
		return Total{}, &IncompleteEvaluationsError{Graded: len(evaluations), Criteria: len(criteria)}
	}
	return Total{
		studentID:   studentID,
		value:       calculation.Value,
		graded:      len(evaluations),
		criteria:    len(criteria),
		preliminary: calculation.Preliminary,
	}, nil
}

// FinalWeightedTotal calculates a weighted total only when every criterion is graded.
type FinalWeightedTotal struct{}

// IncompleteEvaluationsError lets a report represent a missing final total
// without treating unrelated calculation errors as an incomplete gradebook.
type IncompleteEvaluationsError struct {
	Graded   int
	Criteria int
}

// Error describes how many evaluations are missing from the final total.
func (e *IncompleteEvaluationsError) Error() string {
	return fmt.Sprintf(
		"для окончательного итога нужны оценки по всем критериям: выставлено %d из %d",
		e.Graded, e.Criteria,
	)
}

// Calculate returns the final weighted value for the actual rubric and one student's
// validated evaluations. It rejects an empty rubric or missing evaluations.
func (FinalWeightedTotal) Calculate(criteria []Criterion, evaluations []Evaluation) (TotalCalculation, error) {
	if len(criteria) == 0 {
		return TotalCalculation{}, errors.New("нельзя рассчитать итог: рубрикатор пуст")
	}
	if len(evaluations) != len(criteria) {
		return TotalCalculation{}, &IncompleteEvaluationsError{Graded: len(evaluations), Criteria: len(criteria)}
	}
	return TotalCalculation{Value: weightedTotal(criteria, evaluations)}, nil
}

// PreliminaryWeightedTotal calculates a weighted total with missing scores treated as zero.
type PreliminaryWeightedTotal struct{}

// Calculate returns the preliminary weighted value for the actual rubric and one
// student's validated evaluations, including missing scores as zero.
func (PreliminaryWeightedTotal) Calculate(criteria []Criterion, evaluations []Evaluation) (TotalCalculation, error) {
	if len(criteria) == 0 {
		return TotalCalculation{}, errors.New("нельзя рассчитать итог: рубрикатор пуст")
	}
	return TotalCalculation{Value: weightedTotal(criteria, evaluations), Preliminary: true}, nil
}

func weightedTotal(criteria []Criterion, evaluations []Evaluation) float64 {
	scores := make(map[CriterionID]float64, len(evaluations))
	for _, evaluation := range evaluations {
		scores[evaluation.CriterionID()] = evaluation.Score().Value()
	}
	var largestWeight float64
	for _, criterion := range criteria {
		largestWeight = max(largestWeight, criterion.Weight().Value())
	}
	var weightedScore, totalWeight float64
	for _, criterion := range criteria {
		// Scaling all weights by the same positive constant preserves the total
		// and prevents overflow when valid weights are close to MaxFloat64.
		weight := criterion.Weight().Value() / largestWeight
		ratio := scores[criterion.ID()] / criterion.MaxScore()
		weightedScore += weight * ratio
		totalWeight += weight
	}
	return 10 * (weightedScore / totalWeight)
}
