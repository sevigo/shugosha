package config

import (
	"encoding/json"
	"testing"

	"github.com/sevigo/shugosha/mocks"
	"github.com/sevigo/shugosha/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestManager_SaveAndLoadConfig(t *testing.T) {
	testConfig := &model.BackupConfig{
		Providers: []model.ProviderConfig{
			{
				Type: "Test",
				Name: "Foo",
			},
		},
	}

	marshaledConfig, err := json.Marshal(testConfig)
	assert.NoError(t, err)

	mockDB := mocks.NewDB(t)
	m := &Manager{
		db: mockDB,
	}

	mockDB.On("Set", "config:backupConfig", marshaledConfig).Return(nil)
	mockDB.On("Get", "config:backupConfig").Return(marshaledConfig, nil)

	err = m.SaveConfig(testConfig)
	assert.NoError(t, err)

	loadedConf, err := m.LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, testConfig, loadedConf)

	mockDB.AssertExpectations(t)
}
