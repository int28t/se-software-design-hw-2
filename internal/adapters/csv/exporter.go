package csv

import (
	encodingcsv "encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"se-software-design-hw-2/internal/application"
)

type Exporter struct{}

func NewExporter() *Exporter { return &Exporter{} }

func (e *Exporter) Export(destination string, rows []application.ResultRow) error {
	file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("создание CSV %q (нужен новый путь): %w", destination, err)
	}
	if err := writeAndClose(file, rows); err != nil {
		cleanupErr := os.Remove(destination)
		return fmt.Errorf("запись CSV %q: %w", destination, errors.Join(err, cleanupErr))
	}
	return nil
}

func writeAndClose(output io.WriteCloser, rows []application.ResultRow) error {
	records := make([][]string, 0, len(rows)+1)
	records = append(records, []string{"student_id", "student_name", "total", "graded_count", "criterion_count", "status"})
	for _, row := range rows {
		value, status := "", "incomplete"
		if row.Total != nil {
			value = strconv.FormatFloat(row.Total.Value(), 'f', 2, 64)
			status = "final"
		}
		records = append(records, []string{
			strconv.Itoa(int(row.StudentID)), row.StudentName, value,
			strconv.Itoa(row.GradedCount), strconv.Itoa(row.CriterionCount), status,
		})
	}
	writer := encodingcsv.NewWriter(output)
	writeErr := writer.WriteAll(records)
	return errors.Join(writeErr, output.Close())
}
