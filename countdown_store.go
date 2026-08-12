package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"health-tool/domain"
)

type countdownFile struct {
	SavedAt time.Time               `json:"savedAt"`
	Events  []domain.CountdownEvent `json:"events"`
}

func loadCountdownFile(path string) (countdownFile, error) {
	var file countdownFile
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return file, nil
	}
	if err != nil {
		return file, err
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return countdownFile{}, err
	}
	return file, nil
}

func saveCountdownFile(path string, file countdownFile) error {
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func userCountdownPath() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "health-tool", "countdowns.json"), nil
}
