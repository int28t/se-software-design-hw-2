package domain

import (
	"errors"
	"math"
	"strings"
)

type CriterionID int

// Weight is a positive finite contribution to a future weighted total.
// It is not a percentage and need not be at most one.
type Weight struct {
	value float64
}

func NewWeight(value float64) (Weight, error) {
	if !positiveFinite(value) {
		return Weight{}, errors.New("вес должен быть положительным конечным числом")
	}
	return Weight{value: value}, nil
}

func (w Weight) Value() float64 { return w.value }

type Criterion struct {
	id       CriterionID
	name     string
	maxScore float64
	weight   Weight
}

func newCriterion(id CriterionID, name string, maxScore, weight float64) (Criterion, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Criterion{}, errors.New("название критерия не должно быть пустым")
	}
	if !positiveFinite(maxScore) {
		return Criterion{}, errors.New("максимальный балл должен быть положительным конечным числом")
	}
	w, err := NewWeight(weight)
	if err != nil {
		return Criterion{}, err
	}
	return Criterion{id: id, name: name, maxScore: maxScore, weight: w}, nil
}

func (c Criterion) ID() CriterionID { return c.id }

func (c Criterion) Name() string { return c.name }

func (c Criterion) MaxScore() float64 { return c.maxScore }

func (c Criterion) Weight() Weight { return c.weight }

func positiveFinite(value float64) bool {
	return value > 0 && !math.IsInf(value, 0) && !math.IsNaN(value)
}
