package domain

import (
	"errors"
	"strings"
)

// StudentID identifies a student within the course.
type StudentID int

// Student is identified by ID: different students may have the same name.
type Student struct {
	id   StudentID
	name string
}

func newStudent(id StudentID, name string) (Student, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Student{}, errors.New("имя студента не должно быть пустым")
	}
	return Student{id: id, name: name}, nil
}

// ID returns the student identity.
func (s Student) ID() StudentID { return s.id }

// Name returns the trimmed student name.
func (s Student) Name() string { return s.name }
