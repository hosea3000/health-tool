package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"health-tool/domain"
)

type CountdownFile struct {
	SavedAt time.Time               `json:"savedAt"`
	Events  []domain.CountdownEvent `json:"events"`
}

func LoadCountdownFile(path string) (CountdownFile, error) {
	var file CountdownFile
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return file, nil
	}
	if err != nil {
		return file, err
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return CountdownFile{}, err
	}
	return file, nil
}

func SaveCountdownFile(path string, file CountdownFile) error {
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func UserCountdownPath() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "health-tool", "countdowns.json"), nil
}
