package application

import (
	"fmt"

	"se-software-design-hw-2/internal/domain"
)

// AddEvaluation validates and saves a new evaluation without overwriting an existing one.
func (s *Service) AddEvaluation(
	studentID domain.StudentID,
	criterionID domain.CriterionID,
	score float64,
) (domain.Evaluation, error) {
	book, err := s.repository.Load()
	if err != nil {
		return domain.Evaluation{}, fmt.Errorf("загрузка ведомости: %w", err)
	}
	evaluation, err := book.AddEvaluation(studentID, criterionID, score)
	if err != nil {
		return domain.Evaluation{}, err
	}
	if err := s.repository.Save(book); err != nil {
		return domain.Evaluation{}, fmt.Errorf("сохранение оценки: %w", err)
	}
	return evaluation, nil
}

// Evaluations loads the current evaluations of an existing student.
func (s *Service) Evaluations(studentID domain.StudentID) ([]domain.Evaluation, error) {
	book, err := s.repository.Load()
	if err != nil {
		return nil, fmt.Errorf("загрузка оценок: %w", err)
	}
	return book.Evaluations(studentID)
}

// CorrectEvaluation validates a replacement score and saves the updated gradebook.
func (s *Service) CorrectEvaluation(
	studentID domain.StudentID,
	criterionID domain.CriterionID,
	score float64,
) (domain.Evaluation, error) {
	book, err := s.repository.Load()
	if err != nil {
		return domain.Evaluation{}, fmt.Errorf("загрузка ведомости: %w", err)
	}
	evaluation, err := book.CorrectEvaluation(studentID, criterionID, score)
	if err != nil {
		return domain.Evaluation{}, err
	}
	if err := s.repository.Save(book); err != nil {
		return domain.Evaluation{}, fmt.Errorf("сохранение исправления оценки: %w", err)
	}
	return evaluation, nil
}

// CancelEvaluation removes an existing evaluation and saves the updated gradebook.
func (s *Service) CancelEvaluation(studentID domain.StudentID, criterionID domain.CriterionID) error {
	book, err := s.repository.Load()
	if err != nil {
		return fmt.Errorf("загрузка ведомости: %w", err)
	}
	if err := book.CancelEvaluation(studentID, criterionID); err != nil {
		return err
	}
	if err := s.repository.Save(book); err != nil {
		return fmt.Errorf("сохранение отмены оценки: %w", err)
	}
	return nil
}
