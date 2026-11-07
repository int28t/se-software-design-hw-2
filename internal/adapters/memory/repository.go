// Package memory stores isolated gradebook snapshots for one console session.
package memory

import "se-software-design-hw-2/internal/domain"

// Repository stores one course in memory for a single console session.
type Repository struct {
	book domain.Gradebook
}

// NewRepository creates an empty in-memory gradebook repository.
func NewRepository() *Repository { return &Repository{} }

// Load returns an independent snapshot of the stored gradebook.
func (r *Repository) Load() (domain.Gradebook, error) {
	return r.book.Clone(), nil
}

// Save replaces the stored gradebook with an independent snapshot.
func (r *Repository) Save(book domain.Gradebook) error {
	r.book = book.Clone()
	return nil
}
