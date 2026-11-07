package application

import (
	"errors"
	"fmt"

	"se-software-design-hw-2/internal/domain"
)

// ResultRow is a report DTO. A nil Total means the final total is unavailable.
type ResultRow struct {
	StudentID      domain.StudentID
	StudentName    string
	Total          *domain.Total
	GradedCount    int
	CriterionCount int
}

// ResultExporter writes a calculated report to an external destination.
type ResultExporter interface {
	Export(destination string, rows []ResultRow) error
}

// Results builds the report from one snapshot, leaving incomplete totals unavailable.
func (s *Service) Results() ([]ResultRow, error) {
	book, err := s.repository.Load()
	if err != nil {
		return nil, fmt.Errorf("загрузка итоговой ведомости: %w", err)
	}
	students := book.Students()
	if len(students) == 0 {
		return nil, errors.New("нет студентов для формирования итоговой ведомости")
	}
	calculator := domain.NewTotalCalculator(domain.FinalWeightedTotal{})
	rows := make([]ResultRow, 0, len(students))
	for _, student := range students {
		row := ResultRow{StudentID: student.ID(), StudentName: student.Name()}
		total, err := calculator.Calculate(book, student.ID())
		if err != nil {
			incomplete, ok := errors.AsType[*domain.IncompleteEvaluationsError](err)
			if !ok {
				return nil, fmt.Errorf("итог студента ID=%d: %w", student.ID(), err)
			}
			row.GradedCount, row.CriterionCount = incomplete.Graded, incomplete.Criteria
			rows = append(rows, row)
			continue
		}
		row.Total = new(total)
		row.GradedCount, row.CriterionCount = total.GradedCount(), total.CriterionCount()
		rows = append(rows, row)
	}
	return rows, nil
}

// ExportResults builds the current report and delegates writing to the export port.
func (s *Service) ExportResults(destination string) error {
	if s.exporter == nil {
		return errors.New("экспортёр итогов не подключён")
	}
	rows, err := s.Results()
	if err != nil {
		return err
	}
	if err := s.exporter.Export(destination, rows); err != nil {
		return fmt.Errorf("экспорт итоговой ведомости: %w", err)
	}
	return nil
}
