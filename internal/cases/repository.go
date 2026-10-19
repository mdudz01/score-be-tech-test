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
// The API checks for this error to send a 404 response.
var ErrNotFound = errors.New("case not found")

// FileRepository holds cases loaded from a JSON file at startup.
// byID is a map: case ID -> pointer to the case.
type FileRepository struct {
	byID map[string]*domain.Case
}

// LoadFile reads a file containing a single case and returns a repository.
func LoadFile(path string) (*FileRepository, error) {
	// Read the whole file into memory as bytes.
	raw, err := os.ReadFile(path)
	if err != nil {
		// %w wraps the original error so it isn't lost.
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	// Turn the JSON into a Case. &c passes a pointer so Unmarshal can fill it in.
	var c domain.Case
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	// A case with no ID could never be looked up, so treat it as an error.
	if c.CaseID == "" {
		return nil, fmt.Errorf("%s: case has no case_id", path)
	}

	// Store the case in a map under its ID.
	return &FileRepository{byID: map[string]*domain.Case{c.CaseID: &c}}, nil
}

// GetCase returns the case with the given ID, or ErrNotFound.
// ctx is unused here, but a database version would use it for timeouts.
func (r *FileRepository) GetCase(ctx context.Context, id string) (*domain.Case, error) {
	// ok is false if the ID isn't in the map.
	c, ok := r.byID[id]
	if !ok {
		return nil, ErrNotFound
	}
	return c, nil
}
