package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"health-tool/model"
)

func TestLoadSettingsDefaultsWhenFileIsMissing(t *testing.T) {
	settings, err := LoadSettings(filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if settings.ReminderMinutes != 60 || settings.RestMinutes != 3 {
		t.Fatalf("defaults = %+v, want 60 reminder and 3 rest minutes", settings)
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	want := model.Settings{ReminderMinutes: 45, RestMinutes: 5}
	if err := SaveSettings(path, want); err != nil {
		t.Fatal(err)
	}

	got, err := LoadSettings(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("settings = %+v, want %+v", got, want)
	}
}

func TestSaveSettingsRejectsInvalidReminderMinutes(t *testing.T) {
	if err := SaveSettings(filepath.Join(t.TempDir(), "settings.json"), model.Settings{ReminderMinutes: 44}); err == nil {
		t.Fatal("invalid reminder duration was accepted")
	}
}

func TestSaveSettingsDoesNotPersistAutoStart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := SaveSettings(path, model.Settings{ReminderMinutes: 45, RestMinutes: 5, AutoStart: true}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "autoStart") {
		t.Fatalf("settings.json 不应持久化 autoStart 字段: %s", data)
	}
	got, err := LoadSettings(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.AutoStart {
		t.Fatal("加载结果不应携带 autoStart 状态")
	}
}
