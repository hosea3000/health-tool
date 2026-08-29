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

func TestLoadSettingsDefaultsNotificationsEnabledForLegacyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	// 老配置文件：不含 notificationsEnabled 字段
	if err := os.WriteFile(path, []byte(`{"reminderMinutes":45,"restMinutes":5}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := LoadSettings(path)
	if err != nil {
		t.Fatal(err)
	}
	if !got.NotificationsEnabled {
		t.Fatalf("legacy settings notificationsEnabled = false, want true (默认开启，老文件无迁移)")
	}
}

func TestLoadSettingsKeepsExplicitNotificationsDisabled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte(`{"reminderMinutes":45,"restMinutes":5,"notificationsEnabled":false}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := LoadSettings(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.NotificationsEnabled {
		t.Fatal("explicit notificationsEnabled=false was not preserved")
	}
}

func TestSettingsNotificationsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	want := model.Settings{ReminderMinutes: 45, RestMinutes: 5, NotificationsEnabled: false}
	if err := SaveSettings(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadSettings(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.NotificationsEnabled {
		t.Fatalf("notifications round trip = true, want false")
	}
	if !strings.Contains(func() string { data, _ := os.ReadFile(path); return string(data) }(), `"notificationsEnabled": false`) {
		t.Fatal("settings.json should persist notificationsEnabled")
	}
}
