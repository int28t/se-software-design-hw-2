package domain

import "slices"

// Gradebook is the aggregate root for one course. Its zero value is an empty book.
type Gradebook struct {
	students []Student
	criteria []Criterion
}

func (g *Gradebook) AddStudent(name string) (Student, error) {
	student, err := newStudent(StudentID(len(g.students)+1), name)
	if err != nil {
		return Student{}, err
	}
	g.students = append(g.students, student)
	return student, nil
}

func (g *Gradebook) AddCriterion(name string, maxScore, weight float64) (Criterion, error) {
	criterion, err := newCriterion(CriterionID(len(g.criteria)+1), name, maxScore, weight)
	if err != nil {
		return Criterion{}, err
	}
	g.criteria = append(g.criteria, criterion)
	return criterion, nil
}

func (g Gradebook) Students() []Student { return slices.Clone(g.students) }

func (g Gradebook) Criteria() []Criterion { return slices.Clone(g.criteria) }

// Clone isolates the mutable collections for repository load/save boundaries.
func (g Gradebook) Clone() Gradebook {
	return Gradebook{students: g.Students(), criteria: g.Criteria()}
}
