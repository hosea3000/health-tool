package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type CardOrderFile struct {
	Order []string `json:"order"`
}

func LoadCardOrderFile(path string) ([]string, error) {
	var file CardOrderFile
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return file.Order, nil
}

func SaveCardOrderFile(path string, order []string) error {
	data, err := json.MarshalIndent(CardOrderFile{Order: order}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func UserCardOrderPath() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "health-tool", "card_order.json"), nil
}
