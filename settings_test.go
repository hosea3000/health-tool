package main

import (
	"path/filepath"
	"testing"
)

func TestLoadSettingsDefaultsWhenFileIsMissing(t *testing.T) {
	settings, err := loadSettings(filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if settings.ReminderMinutes != 60 || settings.RestMinutes != 3 {
		t.Fatalf("defaults = %+v, want 60 reminder and 3 rest minutes", settings)
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	want := Settings{ReminderMinutes: 45, RestMinutes: 5}
	if err := saveSettings(path, want); err != nil {
		t.Fatal(err)
	}

	got, err := loadSettings(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("settings = %+v, want %+v", got, want)
	}
}

func TestSaveSettingsRejectsInvalidReminderMinutes(t *testing.T) {
	if err := saveSettings(filepath.Join(t.TempDir(), "settings.json"), Settings{ReminderMinutes: 44}); err == nil {
		t.Fatal("invalid reminder duration was accepted")
	}
}
