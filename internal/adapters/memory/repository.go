package memory

import "se-software-design-hw-2/internal/domain"

// Repository stores one course in memory for a single console session.
type Repository struct {
	book domain.Gradebook
}

func NewRepository() *Repository { return &Repository{} }

func (r *Repository) Load() (domain.Gradebook, error) {
	return r.book.Clone(), nil
}

func (r *Repository) Save(book domain.Gradebook) error {
	r.book = book.Clone()
	return nil
}
