package main

import (
	"context"
	"log"
	"sort"
	"sync"
	"sync/atomic"
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
	timelinePath  string
	currentDate   string
	timeline      []TimelineEntry
	quitRequested atomic.Bool
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
	if timelinePath, err := userTimelinePath(); err == nil {
		app.timelinePath = timelinePath
	}
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
	a.mu.Lock()
	a.currentDate = a.now().Format("2006-01-02")
	a.loadTimelineLocked(a.now())
	a.mu.Unlock()
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
	a.mu.Lock()
	if len(a.timeline) > 0 && a.timeline[len(a.timeline)-1].EndedAt == nil {
		last := &a.timeline[len(a.timeline)-1]
		endedAt := a.now()
		last.EndedAt = &endedAt
	}
	a.persistTimelineLocked(a.now())
	a.mu.Unlock()
}

func (a *App) beforeClose(ctx context.Context) bool {
	if a.quitRequested.Load() {
		return false
	}
	runtime.WindowHide(ctx)
	return true
}

func (a *App) requestQuit() {
	a.quitRequested.Store(true)
}

type AppStatus struct {
	State                string `json:"state"`
	ElapsedSeconds       int64  `json:"elapsedSeconds"`
	ReminderMinutes      int    `json:"reminderMinutes"`
	RestMinutes          int    `json:"restMinutes"`
	RestRemainingSeconds int64  `json:"restRemainingSeconds"`
}

type TimelineEntry struct {
	Kind            string     `json:"kind"`
	StartedAt       time.Time  `json:"startedAt"`
	EndedAt         *time.Time `json:"endedAt,omitempty"`
	DurationSeconds int64      `json:"durationSeconds"`
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

func (a *App) Timeline() []TimelineEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := a.now()
	entries := append([]TimelineEntry(nil), a.timeline...)
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].StartedAt.Before(entries[j].StartedAt)
	})
	for i := range entries {
		end := now
		if entries[i].EndedAt != nil {
			end = *entries[i].EndedAt
		}
		if end.After(entries[i].StartedAt) {
			entries[i].DurationSeconds = int64(end.Sub(entries[i].StartedAt) / time.Second)
		}
	}
	return entries
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
	before := a.monitor.State()
	result := a.monitor.EffectiveActivity(activity)
	a.recordTimelineTransitionLocked(before, a.monitor.State(), activity.At)
	if result.Reminder {
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
			a.rolloverLocked(now)
			reminder := false
			if workstationLocked() || (a.monitor.State() == domain.Working && now.Sub(a.lastTick) >= 5*time.Minute) {
				before := a.monitor.State()
				a.monitor.PauseForIdle()
				a.recordTimelineTransitionLocked(before, a.monitor.State(), now)
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
	before := a.monitor.State()
	if result := a.monitor.Advance(now); result.Reminder {
		a.recordTimelineTransitionLocked(before, a.monitor.State(), now)
		updateTrayState(a.monitor.State().String())
		return true
	}
	a.recordTimelineTransitionLocked(before, a.monitor.State(), now)
	updateTrayState(a.monitor.State().String())
	return false
}

func (a *App) rolloverLocked(now time.Time) {
	if now.Format("2006-01-02") == a.currentDate {
		return
	}
	a.currentDate = now.Format("2006-01-02")
	a.timeline = nil
	a.persistTimelineLocked(now)
}

func (a *App) loadTimelineLocked(now time.Time) {
	if a.timelinePath == "" {
		return
	}
	file, err := loadTimelineFile(a.timelinePath)
	if err != nil {
		return
	}
	if file.Date != now.Format("2006-01-02") {
		return
	}
	a.timeline = append([]TimelineEntry(nil), file.Entries...)
	if len(a.timeline) > 0 && a.timeline[len(a.timeline)-1].EndedAt == nil {
		last := &a.timeline[len(a.timeline)-1]
		savedAt := file.SavedAt
		last.EndedAt = &savedAt
	}
}

func (a *App) persistTimelineLocked(now time.Time) {
	if a.timelinePath == "" {
		return
	}
	_ = saveTimelineFile(a.timelinePath, timelineFile{
		Date:    now.Format("2006-01-02"),
		SavedAt: now,
		Entries: append([]TimelineEntry(nil), a.timeline...),
	})
}

func (a *App) recordTimelineTransitionLocked(before, after domain.State, at time.Time) {
	if before == after {
		return
	}
	if len(a.timeline) > 0 && a.timeline[len(a.timeline)-1].EndedAt == nil {
		last := &a.timeline[len(a.timeline)-1]
		if !at.Before(last.StartedAt) {
			endedAt := at
			last.EndedAt = &endedAt
		}
	}
	if after == domain.Working || after == domain.Resting || after == domain.IdlePaused {
		a.timeline = append(a.timeline, TimelineEntry{Kind: after.String(), StartedAt: at})
	}
	a.persistTimelineLocked(at)
}

func (a *App) notifyReminder(restMinutes int) {
	a.notify(restMinutes)
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "reminder")
	}
}
