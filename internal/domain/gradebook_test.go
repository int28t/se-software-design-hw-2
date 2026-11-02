package domain_test

import (
	"math"
	"strconv"
	"testing"

	"se-software-design-hw-2/internal/domain"
)

func TestStudents(t *testing.T) {
	var book domain.Gradebook
	if len(book.Students()) != 0 {
		t.Fatal("new book must be empty")
	}
	for _, name := range []string{"", " \t\n"} {
		if _, err := book.AddStudent(name); err == nil {
			t.Fatalf("accepted empty name %q", name)
		}
	}
	for i := range 2 {
		s, err := book.AddStudent("  Иван Иванов  ")
		if err != nil {
			t.Fatal(err)
		}
		if s.ID() != domain.StudentID(i+1) || s.Name() != "Иван Иванов" {
			t.Fatalf("unexpected student: %#v", s)
		}
	}
	students := book.Students()
	students[0] = domain.Student{}
	if got := book.Students(); len(got) != 2 || got[0].ID() != 1 || got[1].ID() != 2 {
		t.Fatalf("list exposed aggregate state or invalid input changed IDs: %#v", got)
	}
}

func TestCriterionValidation(t *testing.T) {
	for _, value := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		t.Run("max="+strconv.FormatFloat(value, 'g', -1, 64), func(t *testing.T) {
			var book domain.Gradebook
			if _, err := book.AddCriterion("Экзамен", value, 1); err == nil {
				t.Fatal("accepted invalid maximum")
			}
			if len(book.Criteria()) != 0 {
				t.Fatal("invalid criterion was added")
			}
		})
		t.Run("weight="+strconv.FormatFloat(value, 'g', -1, 64), func(t *testing.T) {
			var book domain.Gradebook
			if _, err := book.AddCriterion("Экзамен", 10, value); err == nil {
				t.Fatal("accepted invalid weight")
			}
			if len(book.Criteria()) != 0 {
				t.Fatal("invalid criterion was added")
			}
		})
	}
	var book domain.Gradebook
	if _, err := book.AddCriterion(" \t", 10, 1); err == nil {
		t.Fatal("accepted empty criterion name")
	}
	for i := range 2 {
		c, err := book.AddCriterion("  Экзамен  ", 20.5, 3)
		if err != nil {
			t.Fatal(err)
		}
		if c.ID() != domain.CriterionID(i+1) || c.Name() != "Экзамен" || c.MaxScore() != 20.5 || c.Weight().Value() != 3 {
			t.Fatalf("unexpected criterion: %#v", c)
		}
	}
	criteria := book.Criteria()
	criteria[0] = domain.Criterion{}
	if got := book.Criteria(); len(got) != 2 || got[0].ID() != 1 || got[1].ID() != 2 {
		t.Fatalf("list exposed aggregate state or invalid input changed IDs: %#v", got)
	}
}

func TestCloneIsolatesAppends(t *testing.T) {
	var book domain.Gradebook
	for range 5 {
		if _, err := book.AddStudent("Студент"); err != nil {
			t.Fatal(err)
		}
		if _, err := book.AddCriterion("Критерий", 10, 1); err != nil {
			t.Fatal(err)
		}
	}
	clone := book.Clone()
	if _, err := clone.AddStudent("Из копии"); err != nil {
		t.Fatal(err)
	}
	if _, err := clone.AddCriterion("Из копии", 1, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := book.AddStudent("Из оригинала"); err != nil {
		t.Fatal(err)
	}
	if _, err := book.AddCriterion("Из оригинала", 2, 2); err != nil {
		t.Fatal(err)
	}
	if clone.Students()[5].Name() != "Из копии" || clone.Criteria()[5].Name() != "Из копии" {
		t.Fatal("clone shares slice backing arrays")
	}
}
