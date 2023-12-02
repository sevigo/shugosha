package backup

import (
	"fmt"
)

// DummyProvider is a simple backup provider that logs file changes.
type DummyProvider struct{}

// NewDummyProvider creates a new DummyProvider.
func NewDummyProvider() *DummyProvider {
	return &DummyProvider{}
}

// Backup logs the file change event.
func (dp *DummyProvider) Backup(path string) error {
	fmt.Printf(">>> DummyProvider: Backing up - %q\n", path)
	return nil
}

func (dp *DummyProvider) Name() string {
	return "Dummy"
}
