package application

import (
	"fmt"

	"se-software-design-hw-2/internal/domain"
)

// GradebookRepository returns independent snapshots; Save commits a snapshot.
type GradebookRepository interface {
	Load() (domain.Gradebook, error)
	Save(domain.Gradebook) error
}

// Service is the facade for the console's application scenarios.
type Service struct {
	repository GradebookRepository
	exporter   ResultExporter
}

func NewService(repository GradebookRepository, exporter ResultExporter) *Service {
	return &Service{repository: repository, exporter: exporter}
}

func (s *Service) AddStudent(name string) (domain.Student, error) {
	book, err := s.repository.Load()
	if err != nil {
		return domain.Student{}, fmt.Errorf("загрузка ведомости: %w", err)
	}
	student, err := book.AddStudent(name)
	if err != nil {
		return domain.Student{}, err
	}
	if err := s.repository.Save(book); err != nil {
		return domain.Student{}, fmt.Errorf("сохранение студента: %w", err)
	}
	return student, nil
}

func (s *Service) Students() ([]domain.Student, error) {
	book, err := s.repository.Load()
	if err != nil {
		return nil, fmt.Errorf("загрузка студентов: %w", err)
	}
	return book.Students(), nil
}

func (s *Service) AddCriterion(name string, maxScore, weight float64) (domain.Criterion, error) {
	book, err := s.repository.Load()
	if err != nil {
		return domain.Criterion{}, fmt.Errorf("загрузка ведомости: %w", err)
	}
	criterion, err := book.AddCriterion(name, maxScore, weight)
	if err != nil {
		return domain.Criterion{}, err
	}
	if err := s.repository.Save(book); err != nil {
		return domain.Criterion{}, fmt.Errorf("сохранение критерия: %w", err)
	}
	return criterion, nil
}

func (s *Service) Criteria() ([]domain.Criterion, error) {
	book, err := s.repository.Load()
	if err != nil {
		return nil, fmt.Errorf("загрузка критериев: %w", err)
	}
	return book.Criteria(), nil
}
