package echo

import (
	"fmt"

	"github.com/sevigo/shugosha/pkg/config"
	"github.com/sevigo/shugosha/pkg/model"
)

// Provider is a simple backup provider that logs file changes.
type provider struct {
	directoryList []string
}

// NewEchoProvider creates a new EchoProvider.
func NewEchoProvider(providerConfig *config.ProviderConfig) (model.Provider, error) {
	return &provider{
		directoryList: providerConfig.DirectoryList,
	}, nil
}

// Backup logs the file change event.
func (p *provider) Backup(path string) error {
	fmt.Printf("[Echo] Backing up - %q\n", path)
	return nil
}

func (p *provider) Name() string {
	return "Echo"
}

func (p *provider) DirectoryList() []string {
	return p.directoryList
}
