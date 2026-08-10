package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"health-tool/domain"
)

type Settings struct {
	ReminderMinutes int `json:"reminderMinutes"`
	RestMinutes     int `json:"restMinutes"`
}

func defaultSettings() Settings {
	return Settings{
		ReminderMinutes: int(domain.DefaultReminderDuration / time.Minute),
		RestMinutes:     int(domain.DefaultRestDuration / time.Minute),
	}
}

func loadSettings(path string) (Settings, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return defaultSettings(), nil
	}
	if err != nil {
		return Settings{}, err
	}

	settings := defaultSettings()
	if err := json.Unmarshal(data, &settings); err != nil {
		return Settings{}, err
	}
	if !domain.ValidReminderDuration(durationFromMinutes(settings.ReminderMinutes)) || !domain.ValidRestDuration(durationFromMinutes(settings.RestMinutes)) {
		return defaultSettings(), nil
	}
	return settings, nil
}

func saveSettings(path string, settings Settings) error {
	if !domain.ValidReminderDuration(durationFromMinutes(settings.ReminderMinutes)) {
		return errors.New("reminder minutes must be between 1 and 180 in steps of 5")
	}
	if !domain.ValidRestDuration(durationFromMinutes(settings.RestMinutes)) {
		return errors.New("rest minutes must be between 1 and 30")
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func userSettingsPath() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "health-tool", "settings.json"), nil
}

func durationFromMinutes(minutes int) time.Duration {
	return time.Duration(minutes) * time.Minute
}
