package domain

import (
	"errors"
	"fmt"
)

// Total is a calculated snapshot, never an independently stored grade.
type Total struct {
	studentID   StudentID
	value       float64
	graded      int
	criteria    int
	preliminary bool
}

func (t Total) StudentID() StudentID { return t.studentID }

func (t Total) Value() float64 { return t.value }

func (t Total) GradedCount() int { return t.graded }

func (t Total) CriterionCount() int { return t.criteria }

func (t Total) Complete() bool { return t.criteria > 0 && t.graded == t.criteria }

func (t Total) Preliminary() bool { return t.preliminary }

// TotalStrategy receives the actual rubric and one student's evaluations
// validated by the aggregate. It controls how missing evaluations affect a total.
type TotalStrategy interface {
	Calculate([]Criterion, []Evaluation) (Total, error)
}

// TotalCalculator is a Domain Service: it combines the rubric with current grades.
type TotalCalculator struct {
	strategy TotalStrategy
}

func NewTotalCalculator(strategy TotalStrategy) *TotalCalculator {
	return &TotalCalculator{strategy: strategy}
}

func (c *TotalCalculator) Calculate(book Gradebook, studentID StudentID) (Total, error) {
	if c.strategy == nil {
		return Total{}, errors.New("не задана стратегия расчёта итога")
	}
	evaluations, err := book.Evaluations(studentID)
	if err != nil {
		return Total{}, err
	}
	total, err := c.strategy.Calculate(book.Criteria(), evaluations)
	if err != nil {
		return Total{}, err
	}
	total.studentID = studentID
	return total, nil
}

type FinalWeightedTotal struct{}

// IncompleteEvaluationsError lets a report represent a missing final total
// without treating unrelated calculation errors as an incomplete gradebook.
type IncompleteEvaluationsError struct {
	Graded   int
	Criteria int
}

func (e *IncompleteEvaluationsError) Error() string {
	return fmt.Sprintf("для окончательного итога нужны оценки по всем критериям: выставлено %d из %d", e.Graded, e.Criteria)
}

func (FinalWeightedTotal) Calculate(criteria []Criterion, evaluations []Evaluation) (Total, error) {
	if len(criteria) == 0 {
		return Total{}, errors.New("нельзя рассчитать итог: рубрикатор пуст")
	}
	if len(evaluations) != len(criteria) {
		return Total{}, &IncompleteEvaluationsError{Graded: len(evaluations), Criteria: len(criteria)}
	}
	return weightedTotal(criteria, evaluations, false), nil
}

type PreliminaryWeightedTotal struct{}

func (PreliminaryWeightedTotal) Calculate(criteria []Criterion, evaluations []Evaluation) (Total, error) {
	if len(criteria) == 0 {
		return Total{}, errors.New("нельзя рассчитать итог: рубрикатор пуст")
	}
	return weightedTotal(criteria, evaluations, true), nil
}

func weightedTotal(criteria []Criterion, evaluations []Evaluation, preliminary bool) Total {
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
	return Total{
		value:       10 * (weightedScore / totalWeight),
		graded:      len(evaluations),
		criteria:    len(criteria),
		preliminary: preliminary,
	}
}
