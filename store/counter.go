package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"health-tool/domain"
)

type CounterFile struct {
	SavedAt  time.Time        `json:"savedAt"`
	Counters []domain.Counter `json:"counters"`
}

func LoadCounterFile(path string) (CounterFile, error) {
	var file CounterFile
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return file, nil
	}
	if err != nil {
		return file, err
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return CounterFile{}, err
	}
	return file, nil
}

func SaveCounterFile(path string, file CounterFile) error {
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func UserCounterPath() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "health-tool", "counters.json"), nil
}
