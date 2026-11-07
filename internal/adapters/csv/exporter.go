// Package csv exports result reports to new local CSV files.
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

// Exporter writes result reports as CSV without overwriting existing files.
type Exporter struct{}

// NewExporter creates a CSV result exporter.
func NewExporter() *Exporter { return &Exporter{} }

// Export writes a new CSV file and reports write, close, and cleanup failures.
func (e *Exporter) Export(destination string, rows []application.ResultRow) error {
	file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("создание CSV %q (нужен новый путь): %w", destination, err)
	}
	if err := e.writeAndClose(file, rows); err != nil {
		cleanupErr := os.Remove(destination)
		return fmt.Errorf("запись CSV %q: %w", destination, errors.Join(err, cleanupErr))
	}
	return nil
}

// writeAndClose writes the report and closes the owned output, even after a write failure.
func (e *Exporter) writeAndClose(output io.WriteCloser, rows []application.ResultRow) error {
	writer := encodingcsv.NewWriter(output)
	writeErr := writer.WriteAll(resultRecords(rows))
	return errors.Join(writeErr, output.Close())
}

func resultRecords(rows []application.ResultRow) [][]string {
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
	return records
}
