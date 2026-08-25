package fs

import (
	"fmt"
	"os"
)

// Idempotent: get or create file
func EnsureFile(filePath string) (*os.File, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("Open file %q: %w", filePath, err)
	}
	return file, nil
}
