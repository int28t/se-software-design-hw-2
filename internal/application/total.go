package application

import (
	"fmt"

	"se-software-design-hw-2/internal/domain"
)

func (s *Service) Total(studentID domain.StudentID) (domain.Total, error) {
	return s.calculateTotal(studentID, domain.FinalWeightedTotal{})
}

func (s *Service) PreliminaryTotal(studentID domain.StudentID) (domain.Total, error) {
	return s.calculateTotal(studentID, domain.PreliminaryWeightedTotal{})
}

func (s *Service) calculateTotal(studentID domain.StudentID, strategy domain.TotalStrategy) (domain.Total, error) {
	book, err := s.repository.Load()
	if err != nil {
		return domain.Total{}, fmt.Errorf("загрузка ведомости для расчёта итога: %w", err)
	}
	calculator := domain.NewTotalCalculator(strategy)
	return calculator.Calculate(book, studentID)
}
