package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"health-tool/domain"
)

type App struct {
	ctx           context.Context
	mu            sync.Mutex
	monitor       *domain.Monitor
	now           func() time.Time
	notify        func(int)
	notifyStarted func(int)
	stopInput     func()
	lastTick      time.Time
	settings      Settings
	settingsPath  string
}

func NewApp() *App {
	settings := defaultSettings()
	path, err := userSettingsPath()
	if err == nil {
		if loaded, loadErr := loadSettings(path); loadErr == nil {
			settings = loaded
		}
	}
	app := newAppWithSettings(time.Now, func() {}, settings)
	app.settingsPath = path
	app.notify, app.notifyStarted = newNotifiers()
	return app
}

func newApp(now func() time.Time, notify func()) *App {
	return newAppWithSettings(now, notify, defaultSettings())
}

func newAppWithSettings(now func() time.Time, notify func(), settings Settings) *App {
	return &App{
		monitor:       domain.NewMonitor(now(), durationFromMinutes(settings.ReminderMinutes), durationFromMinutes(settings.RestMinutes)),
		now:           now,
		notify:        func(int) { notify() },
		notifyStarted: func(int) {},
		settings:      settings,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	var err error
	a.stopInput, err = startInputMonitor(a.recordActivity)
	if err != nil {
		log.Printf("input monitoring unavailable: %v", err)
	}
	a.lastTick = a.now()
	a.startTray()
	go a.tick()
}

func (a *App) shutdown(_ context.Context) {
	if a.stopInput != nil {
		a.stopInput()
	}
	stopTray()
}

func (a *App) beforeClose(ctx context.Context) bool {
	runtime.WindowHide(ctx)
	return true
}

type AppStatus struct {
	State                string `json:"state"`
	ElapsedSeconds       int64  `json:"elapsedSeconds"`
	ReminderMinutes      int    `json:"reminderMinutes"`
	RestMinutes          int    `json:"restMinutes"`
	RestRemainingSeconds int64  `json:"restRemainingSeconds"`
}

func (a *App) Status() AppStatus {
	a.mu.Lock()
	now := a.now()
	reminder := a.advanceLocked(now)
	restRemaining := a.monitor.RestRemaining(now)
	restRemainingSeconds := int64(restRemaining / time.Second)
	if restRemaining%time.Second != 0 {
		restRemainingSeconds++
	}
	status := AppStatus{
		State:                a.monitor.State().String(),
		ElapsedSeconds:       int64(a.monitor.Elapsed(now).Seconds()),
		ReminderMinutes:      a.settings.ReminderMinutes,
		RestMinutes:          a.settings.RestMinutes,
		RestRemainingSeconds: restRemainingSeconds,
	}
	a.mu.Unlock()
	if reminder {
		a.notifyReminder(a.settings.RestMinutes)
	}
	return status
}

func (a *App) GetSettings() Settings {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settings
}

func (a *App) SaveSettings(reminderMinutes int, restMinutes int) bool {
	settings := Settings{ReminderMinutes: reminderMinutes, RestMinutes: restMinutes}
	if a.settingsPath == "" || saveSettings(a.settingsPath, settings) != nil {
		return false
	}
	a.mu.Lock()
	a.settings = settings
	a.monitor.SetReminderDuration(durationFromMinutes(reminderMinutes))
	a.monitor.SetRestDuration(durationFromMinutes(restMinutes))
	a.mu.Unlock()
	return true
}

func (a *App) recordActivity(activity domain.EffectiveActivity) {
	a.mu.Lock()
	reminder := a.advanceLocked(activity.At)
	started := false
	if result := a.monitor.EffectiveActivity(activity); result.Reminder {
		reminder = true
	} else if result.Changed && a.monitor.State() == domain.Working {
		started = true
	}
	restMinutes := a.settings.RestMinutes
	reminderMinutes := a.settings.ReminderMinutes
	updateTrayState(a.monitor.State().String())
	a.mu.Unlock()
	if started {
		a.notifyStarted(reminderMinutes)
	}
	if reminder {
		a.notifyReminder(restMinutes)
	}
}

func (a *App) tick() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			a.mu.Lock()
			now := a.now()
			reminder := false
			if workstationLocked() || (a.monitor.State() == domain.Working && now.Sub(a.lastTick) >= 5*time.Minute) {
				a.monitor.PauseForIdle()
				updateTrayState(a.monitor.State().String())
			} else {
				reminder = a.advanceLocked(now)
			}
			restMinutes := a.settings.RestMinutes
			a.lastTick = now
			a.mu.Unlock()
			if reminder {
				a.notifyReminder(restMinutes)
			}
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *App) advanceLocked(now time.Time) bool {
	if result := a.monitor.Advance(now); result.Reminder {
		updateTrayState(a.monitor.State().String())
		return true
	}
	updateTrayState(a.monitor.State().String())
	return false
}

func (a *App) notifyReminder(restMinutes int) {
	a.notify(restMinutes)
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "reminder")
	}
}
