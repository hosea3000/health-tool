package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"health-tool/domain"
	"health-tool/model"
)

func LoadSettings(path string) (model.Settings, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return model.DefaultSettings(), nil
	}
	if err != nil {
		return model.Settings{}, err
	}

	settings := model.DefaultSettings()
	if err := json.Unmarshal(data, &settings); err != nil {
		return model.Settings{}, err
	}
	if !domain.ValidReminderDuration(domain.DurationFromMinutes(settings.ReminderMinutes)) || !domain.ValidRestDuration(domain.DurationFromMinutes(settings.RestMinutes)) {
		return model.DefaultSettings(), nil
	}
	return settings, nil
}

func SaveSettings(path string, settings model.Settings) error {
	if !domain.ValidReminderDuration(domain.DurationFromMinutes(settings.ReminderMinutes)) {
		return errors.New("reminder minutes must be between 1 and 180 in steps of 5")
	}
	if !domain.ValidRestDuration(domain.DurationFromMinutes(settings.RestMinutes)) {
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

func UserSettingsPath() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "health-tool", "settings.json"), nil
}
