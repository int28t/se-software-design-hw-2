package csv

import (
	"bytes"
	encodingcsv "encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"se-software-design-hw-2/internal/application"
	"se-software-design-hw-2/internal/domain"
)

func TestExportRoundTrip(t *testing.T) {
	var book domain.Gradebook
	for _, name := range []string{"Анна, \"Аня\"\nИванова", "Борис", "Пётр"} {
		if _, err := book.AddStudent(name); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := book.AddCriterion("Экзамен", 10, 1); err != nil {
		t.Fatal(err)
	}
	var rows []application.ResultRow
	for _, grade := range []struct {
		id    domain.StudentID
		score float64
	}{{1, 8}, {2, 0}} {
		if _, err := book.AddEvaluation(grade.id, 1, grade.score); err != nil {
			t.Fatal(err)
		}
		total, err := domain.NewTotalCalculator(domain.FinalWeightedTotal{}).Calculate(book, grade.id)
		if err != nil {
			t.Fatal(err)
		}
		rows = append(rows, application.ResultRow{
			StudentID: grade.id, StudentName: book.Students()[int(grade.id)-1].Name(),
			Total: new(total), GradedCount: 1, CriterionCount: 1,
		})
	}
	rows = append(rows, application.ResultRow{StudentID: 3, StudentName: "Пётр", CriterionCount: 1})
	destination := filepath.Join(t.TempDir(), "results.csv")
	if err := NewExporter().Export(destination, rows); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	got, err := encodingcsv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"student_id", "student_name", "total", "graded_count", "criterion_count", "status"},
		{"1", "Анна, \"Аня\"\nИванова", "8.00", "1", "1", "final"},
		{"2", "Борис", "0.00", "1", "1", "final"},
		{"3", "Пётр", "", "0", "1", "incomplete"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CSV changed values or order:\ngot %#v\nwant %#v", got, want)
	}
}

func TestExportPreservesExistingFileAndReportsOpenErrors(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "existing.csv")
	if err := os.WriteFile(destination, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := NewExporter().Export(destination, nil); !errors.Is(err, os.ErrExist) {
		t.Fatalf("existing file was not rejected: %v", err)
	}
	data, err := os.ReadFile(destination)
	if err != nil || string(data) != "original" {
		t.Fatalf("existing file changed: %q, %v", data, err)
	}
	missing := filepath.Join(t.TempDir(), "missing", "results.csv")
	if err := NewExporter().Export(missing, nil); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing parent directory was not reported: %v", err)
	}
}

type failingOutput struct {
	writeErr error
	closeErr error
	closed   bool
}

func (w *failingOutput) Write(data []byte) (int, error) {
	if w.writeErr != nil {
		return 0, w.writeErr
	}
	return len(data), nil
}

func (w *failingOutput) Close() error {
	w.closed = true
	return w.closeErr
}

func TestWriteAndCloseReportsAllIOFailures(t *testing.T) {
	writeErr, closeErr := errors.New("write failure"), errors.New("close failure")
	for _, output := range []*failingOutput{
		{writeErr: writeErr}, {closeErr: closeErr}, {writeErr: writeErr, closeErr: closeErr},
	} {
		err := writeAndClose(output, nil)
		if !output.closed || err == nil {
			t.Fatal("writer was not closed or failure was hidden")
		}
		if output.writeErr != nil && !errors.Is(err, writeErr) {
			t.Fatalf("lost write/flush error: %v", err)
		}
		if output.closeErr != nil && !errors.Is(err, closeErr) {
			t.Fatalf("lost close error: %v", err)
		}
	}
}
