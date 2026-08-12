package main

import (
	"context"
	"fmt"
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
	countdownPath   string
	countdowns      []domain.CountdownEvent
	cardOrderPath   string
	cardOrder       []string
	quitRequested   atomic.Bool
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
	if countdownPath, err := userCountdownPath(); err == nil {
		app.countdownPath = countdownPath
	}
	if cardOrderPath, err := userCardOrderPath(); err == nil {
		app.cardOrderPath = cardOrderPath
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
	a.loadCountdownsLocked()
	a.loadCardOrderLocked()
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

type CountdownView struct {
	ID            string      `json:"id"`
	Title         string      `json:"title"`
	Rule          domain.Rule `json:"rule"`
	NextDate      string      `json:"nextDate"`
	RemainingDays int         `json:"remainingDays"`
}

func (a *App) CountdownEvents() []CountdownView {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := a.now()
	views := make([]CountdownView, 0, len(a.countdowns))
	for _, event := range a.countdowns {
		next := event.Rule.NextOccurrence(now)
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		remaining := int(next.Sub(today) / (24 * time.Hour))
		views = append(views, CountdownView{
			ID:            event.ID,
			Title:         event.Title,
			Rule:          event.Rule,
			NextDate:      next.Format("2006-01-02"),
			RemainingDays: remaining,
		})
	}
	sort.SliceStable(views, func(i, j int) bool {
		ri, rj := views[i].RemainingDays, views[j].RemainingDays
		if (ri >= 0) != (rj >= 0) {
			return ri >= 0
		}
		if ri < 0 {
			return ri > rj
		}
		return ri < rj
	})
	return views
}

func (a *App) AddCountdown(title string, rule domain.Rule) string {
	event := domain.CountdownEvent{ID: fmt.Sprintf("%d", a.now().UnixNano()), Title: title, Rule: rule}
	if err := event.Validate(); err != nil {
		return err.Error()
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.countdowns = append(a.countdowns, event)
	a.persistCountdownsLocked()
	return ""
}

func (a *App) UpdateCountdown(id string, title string, rule domain.Rule) string {
	event := domain.CountdownEvent{ID: id, Title: title, Rule: rule}
	if err := event.Validate(); err != nil {
		return err.Error()
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.countdowns {
		if a.countdowns[i].ID == id {
			a.countdowns[i] = event
			a.persistCountdownsLocked()
			return ""
		}
	}
	return "event not found"
}

func (a *App) DeleteCountdown(id string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.countdowns {
		if a.countdowns[i].ID == id {
			a.countdowns = append(a.countdowns[:i], a.countdowns[i+1:]...)
			a.persistCountdownsLocked()
			return true
		}
	}
	return false
}

func (a *App) loadCountdownsLocked() {
	if a.countdownPath == "" {
		return
	}
	file, err := loadCountdownFile(a.countdownPath)
	if err != nil {
		return
	}
	a.countdowns = append([]domain.CountdownEvent(nil), file.Events...)
}

func (a *App) GetCardOrder() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]string{}, a.cardOrder...)
}

func (a *App) SaveCardOrder(order []string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cardOrderPath == "" || saveCardOrderFile(a.cardOrderPath, order) != nil {
		return false
	}
	a.cardOrder = append([]string(nil), order...)
	return true
}

func (a *App) loadCardOrderLocked() {
	if a.cardOrderPath == "" {
		return
	}
	order, err := loadCardOrderFile(a.cardOrderPath)
	if err != nil {
		return
	}
	a.cardOrder = append([]string(nil), order...)
}

func (a *App) persistCountdownsLocked() {
	if a.countdownPath == "" {
		return
	}
	_ = saveCountdownFile(a.countdownPath, countdownFile{
		SavedAt: a.now(),
		Events:  append([]domain.CountdownEvent(nil), a.countdowns...),
	})
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
