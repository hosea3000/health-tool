package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type timelineFile struct {
	Date    string          `json:"date"`
	SavedAt time.Time       `json:"savedAt"`
	Entries []TimelineEntry `json:"entries"`
}

func loadTimelineFile(path string) (timelineFile, error) {
	var file timelineFile
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return file, nil
	}
	if err != nil {
		return file, err
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return timelineFile{}, err
	}
	return file, nil
}

func saveTimelineFile(path string, file timelineFile) error {
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func userTimelinePath() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "health-tool", "timeline.json"), nil
}
