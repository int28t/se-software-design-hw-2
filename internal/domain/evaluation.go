package domain

import (
	"errors"
	"fmt"
	"math"
	"slices"
)

// Score is a finite nonnegative value; its criterion's upper bound is checked by the factory.
type Score struct {
	value float64
}

// NewScore validates and creates a finite nonnegative score.
func NewScore(value float64) (Score, error) {
	if value < 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return Score{}, errors.New("балл должен быть неотрицательным конечным числом")
	}
	return Score{value: value}, nil
}

// Value returns the numeric score.
func (s Score) Value() float64 { return s.value }

// Evaluation is identified by the student/criterion pair, including when corrected.
type Evaluation struct {
	studentID   StudentID
	criterionID CriterionID
	score       Score
}

// StudentID returns the identity of the evaluated student.
func (e Evaluation) StudentID() StudentID { return e.studentID }

// CriterionID returns the identity of the evaluated rubric item.
func (e Evaluation) CriterionID() CriterionID { return e.criterionID }

// Score returns the current evaluation score.
func (e Evaluation) Score() Score { return e.score }

// EvaluationFactory validates references against the actual course rubric.
// Both initial grading and corrections use this creation path.
type EvaluationFactory struct{}

// Create validates both references and the score against the rubric without changing the book.
func (EvaluationFactory) Create(
	book Gradebook,
	studentID StudentID,
	criterionID CriterionID,
	value float64,
) (Evaluation, error) {
	if err := book.requireStudent(studentID); err != nil {
		return Evaluation{}, err
	}
	index := slices.IndexFunc(book.criteria, func(c Criterion) bool { return c.ID() == criterionID })
	if index < 0 {
		return Evaluation{}, fmt.Errorf("критерий с ID=%d не найден в рубрикаторе", criterionID)
	}
	score, err := NewScore(value)
	if err != nil {
		return Evaluation{}, err
	}
	if value > book.criteria[index].MaxScore() {
		return Evaluation{}, fmt.Errorf("балл должен быть в диапазоне от 0 до %g", book.criteria[index].MaxScore())
	}
	return Evaluation{studentID: studentID, criterionID: criterionID, score: score}, nil
}

// AddEvaluation adds a validated evaluation and rejects an already graded pair.
func (g *Gradebook) AddEvaluation(studentID StudentID, criterionID CriterionID, value float64) (Evaluation, error) {
	evaluation, err := (EvaluationFactory{}).Create(*g, studentID, criterionID, value)
	if err != nil {
		return Evaluation{}, err
	}
	if g.evaluationIndex(studentID, criterionID) >= 0 {
		return Evaluation{}, errors.New("оценка уже выставлена; используйте исправление")
	}
	g.evaluations = append(g.evaluations, evaluation)
	return evaluation, nil
}

// CorrectEvaluation replaces an existing score only after all validation succeeds.
func (g *Gradebook) CorrectEvaluation(studentID StudentID, criterionID CriterionID, value float64) (Evaluation, error) {
	index := g.evaluationIndex(studentID, criterionID)
	if index < 0 {
		return Evaluation{}, missingEvaluation(studentID, criterionID)
	}
	evaluation, err := (EvaluationFactory{}).Create(*g, studentID, criterionID, value)
	if err != nil {
		return Evaluation{}, err
	}
	g.evaluations[index] = evaluation
	return evaluation, nil
}

// CancelEvaluation removes an existing evaluation or leaves the book unchanged.
func (g *Gradebook) CancelEvaluation(studentID StudentID, criterionID CriterionID) error {
	index := g.evaluationIndex(studentID, criterionID)
	if index < 0 {
		return missingEvaluation(studentID, criterionID)
	}
	g.evaluations = slices.Delete(g.evaluations, index, index+1)
	return nil
}

// Evaluations returns an independent list of the student's current evaluations.
func (g *Gradebook) Evaluations(studentID StudentID) ([]Evaluation, error) {
	if err := g.requireStudent(studentID); err != nil {
		return nil, err
	}
	var result []Evaluation
	for _, evaluation := range g.evaluations {
		if evaluation.StudentID() == studentID {
			result = append(result, evaluation)
		}
	}
	return result, nil
}

func (g *Gradebook) requireStudent(id StudentID) error {
	if slices.IndexFunc(g.students, func(s Student) bool { return s.ID() == id }) < 0 {
		return fmt.Errorf("студент с ID=%d не найден", id)
	}
	return nil
}

func (g *Gradebook) evaluationIndex(studentID StudentID, criterionID CriterionID) int {
	return slices.IndexFunc(g.evaluations, func(e Evaluation) bool {
		return e.StudentID() == studentID && e.CriterionID() == criterionID
	})
}

func missingEvaluation(studentID StudentID, criterionID CriterionID) error {
	return fmt.Errorf("оценка студента ID=%d по критерию ID=%d не найдена", studentID, criterionID)
}
