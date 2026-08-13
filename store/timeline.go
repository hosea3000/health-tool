package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"health-tool/model"
)

type TimelineFile struct {
	Date    string               `json:"date"`
	SavedAt time.Time            `json:"savedAt"`
	Entries []model.TimelineEntry `json:"entries"`
}

func LoadTimelineFile(path string) (TimelineFile, error) {
	var file TimelineFile
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return file, nil
	}
	if err != nil {
		return file, err
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return TimelineFile{}, err
	}
	return file, nil
}

func SaveTimelineFile(path string, file TimelineFile) error {
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func UserTimelinePath() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "health-tool", "timeline.json"), nil
}
