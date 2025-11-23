package models

import (
	"encoding/json"
	"os"
	"sync"
)

// AppData represents all persistent application data
type AppData struct {
	Services          map[string]*MonitoredService `json:"services"`
	TelegramConfig    *TelegramConfig              `json:"telegram_config"`
	Histories         map[string]*ServiceHistory   `json:"histories"`
	SystemAlertConfig *SystemAlertConfig           `json:"system_alert_config"`
}

// PersistenceManager handles saving and loading data to/from disk
type PersistenceManager struct {
	filePath string
	mu       sync.RWMutex
}

// NewPersistenceManager creates a new persistence manager
func NewPersistenceManager(filePath string) *PersistenceManager {
	return &PersistenceManager{
		filePath: filePath,
	}
}

// Save writes the app data to disk
func (p *PersistenceManager) Save(data *AppData) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(p.filePath, jsonData, 0644)
}

// Load reads the app data from disk
func (p *PersistenceManager) Load() (*AppData, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// If file doesn't exist, return empty data
	if _, err := os.Stat(p.filePath); os.IsNotExist(err) {
		return &AppData{
			Services:   make(map[string]*MonitoredService),
			Histories:  make(map[string]*ServiceHistory),
			TelegramConfig: &TelegramConfig{
				Enabled: false,
			},
			SystemAlertConfig: &SystemAlertConfig{
				DiskSpaceThreshold: 80.0,
				CPUThreshold:       90.0,
				MemoryThreshold:    90.0,
				Enabled:            true,
			},
		}, nil
	}

	data, err := os.ReadFile(p.filePath)
	if err != nil {
		return nil, err
	}

	var appData AppData
	if err := json.Unmarshal(data, &appData); err != nil {
		return nil, err
	}

	// Ensure maps are initialized
	if appData.Services == nil {
		appData.Services = make(map[string]*MonitoredService)
	}
	if appData.Histories == nil {
		appData.Histories = make(map[string]*ServiceHistory)
	}
	if appData.TelegramConfig == nil {
		appData.TelegramConfig = &TelegramConfig{
			Enabled: false,
		}
	}
	if appData.SystemAlertConfig == nil {
		appData.SystemAlertConfig = &SystemAlertConfig{
			DiskSpaceThreshold: 80.0,
			CPUThreshold:       90.0,
			MemoryThreshold:    90.0,
			Enabled:            true,
		}
	}

	return &appData, nil
}
