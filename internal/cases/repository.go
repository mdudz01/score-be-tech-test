package cases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/mdudz01/score-be-tech-test/internal/domain"
)

// ErrNotFound is returned when no case exists for the requested ID.
var ErrNotFound = errors.New("case not found")

// FileRepository serves cases loaded from a JSON file at startup.
type FileRepository struct {
	byID map[string]*domain.Case
}

// LoadFile reads a file containing a single case object.
func LoadFile(path string) (*FileRepository, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var c domain.Case
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if c.CaseID == "" {
		return nil, fmt.Errorf("%s: case has no case_id", path)
	}

	return &FileRepository{byID: map[string]*domain.Case{c.CaseID: &c}}, nil
}

// GetCase returns the case with the given ID, or ErrNotFound.
func (r *FileRepository) GetCase(ctx context.Context, id string) (*domain.Case, error) {
	c, ok := r.byID[id]
	if !ok {
		return nil, ErrNotFound
	}
	return c, nil
}
