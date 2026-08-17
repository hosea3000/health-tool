package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	goruntime "runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"health-tool/domain"
	"health-tool/model"
	"health-tool/store"
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
	settings      model.Settings
	settingsPath  string
	timelinePath  string
	currentDate   string
	timeline      []model.TimelineEntry
	countdownPath string
	countdowns    []domain.CountdownEvent
	cardOrderPath string
	cardOrder     []string
	countersPath  string
	counters      []domain.Counter
	quitRequested atomic.Bool
	// updateDownloadURL 是最近一次检查发现的新版本资产下载地址；检查失败或已最新时清空。
	updateDownloadURL string
	// updateLatestVersion 是最近一次检查发现的新版本号，用于确认弹窗文案。
	updateLatestVersion string
}

func NewApp() *App {
	settings := model.DefaultSettings()
	path, err := store.UserSettingsPath()
	if err == nil {
		if loaded, loadErr := store.LoadSettings(path); loadErr == nil {
			settings = loaded
		}
	}
	app := newAppWithSettings(time.Now, func() {}, settings)
	app.settingsPath = path
	if timelinePath, err := store.UserTimelinePath(); err == nil {
		app.timelinePath = timelinePath
	}
	if countdownPath, err := store.UserCountdownPath(); err == nil {
		app.countdownPath = countdownPath
	}
	if cardOrderPath, err := store.UserCardOrderPath(); err == nil {
		app.cardOrderPath = cardOrderPath
	}
	if countersPath, err := store.UserCounterPath(); err == nil {
		app.countersPath = countersPath
	}
	app.notify, app.notifyStarted = newNotifiers()
	return app
}

func newApp(now func() time.Time, notify func()) *App {
	return newAppWithSettings(now, notify, model.DefaultSettings())
}

func newAppWithSettings(now func() time.Time, notify func(), settings model.Settings) *App {
	return &App{
		monitor:       domain.NewMonitor(now(), domain.DurationFromMinutes(settings.ReminderMinutes), domain.DurationFromMinutes(settings.RestMinutes)),
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
	a.loadCountersLocked()
	a.mu.Unlock()
	// 清理上次中断下载遗留的 .part 残渣（.new 待确认缓存保留）。
	if exePath, err := os.Executable(); err == nil {
		cleanupUpdateArtifacts(exePath)
	}
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

func (a *App) Status() model.AppStatus {
	a.mu.Lock()
	now := a.now()
	reminder := a.advanceLocked(now)
	restRemaining := a.monitor.RestRemaining(now)
	restRemainingSeconds := int64(restRemaining / time.Second)
	if restRemaining%time.Second != 0 {
		restRemainingSeconds++
	}
	status := model.AppStatus{
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

func (a *App) Timeline() []model.TimelineEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := a.now()
	entries := append([]model.TimelineEntry(nil), a.timeline...)
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

// GetSettings 返回当前设置；AutoStart 字段从注册表实时读取填充，查询失败时保持 false。
func (a *App) GetSettings() model.Settings {
	a.mu.Lock()
	settings := a.settings
	a.mu.Unlock()
	if enabled, err := autoStartEnabled(); err == nil {
		settings.AutoStart = enabled
	}
	return settings
}

// SetAutoStart 开启或关闭开机自启动（写/删 HKCU Run key）。
func (a *App) SetAutoStart(enabled bool) error {
	if err := setAutoStart(enabled); err != nil {
		log.Printf("set autostart %v: %v", enabled, err)
		return err
	}
	return nil
}

// AutoStartEnabled 查询开机自启动当前状态；非 Windows 平台返回错误。
func (a *App) AutoStartEnabled() (bool, error) {
	return autoStartEnabled()
}

func (a *App) SaveSettings(reminderMinutes int, restMinutes int) bool {
	settings := model.Settings{ReminderMinutes: reminderMinutes, RestMinutes: restMinutes}
	if a.settingsPath == "" || store.SaveSettings(a.settingsPath, settings) != nil {
		return false
	}
	a.mu.Lock()
	a.settings = settings
	a.monitor.SetReminderDuration(domain.DurationFromMinutes(reminderMinutes))
	a.monitor.SetRestDuration(domain.DurationFromMinutes(restMinutes))
	a.mu.Unlock()
	return true
}

func (a *App) CountdownEvents() []model.CountdownView {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := a.now()
	views := make([]model.CountdownView, 0, len(a.countdowns))
	for _, event := range a.countdowns {
		next := event.Rule.NextOccurrence(now)
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		remaining := int(next.Sub(today) / (24 * time.Hour))
		views = append(views, model.CountdownView{
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
	file, err := store.LoadCountdownFile(a.countdownPath)
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
	if a.cardOrderPath == "" || store.SaveCardOrderFile(a.cardOrderPath, order) != nil {
		return false
	}
	a.cardOrder = append([]string(nil), order...)
	return true
}

func (a *App) loadCardOrderLocked() {
	if a.cardOrderPath == "" {
		return
	}
	order, err := store.LoadCardOrderFile(a.cardOrderPath)
	if err != nil {
		return
	}
	a.cardOrder = append([]string(nil), order...)
}

func (a *App) persistCountdownsLocked() {
	if a.countdownPath == "" {
		return
	}
	_ = store.SaveCountdownFile(a.countdownPath, store.CountdownFile{
		SavedAt: a.now(),
		Events:  append([]domain.CountdownEvent(nil), a.countdowns...),
	})
}

func (a *App) Counters() []model.CounterView {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := a.now()
	views := make([]model.CounterView, 0, len(a.counters))
	for _, counter := range a.counters {
		views = append(views, a.counterViewLocked(counter, now))
	}
	return views
}

func (a *App) counterViewLocked(counter domain.Counter, now time.Time) model.CounterView {
	count := counter.CurrentCount(now)
	return model.CounterView{
		ID:          counter.ID,
		Name:        counter.Name,
		Period:      counter.Period,
		PeriodLabel: counter.PeriodLabel(),
		Goal:        counter.Goal,
		Count:       count,
		GoalReached: counter.Goal > 0 && count >= counter.Goal,
		History:     a.counterHistoryLocked(counter, now),
	}
}

// counterHistoryLocked 最近 7 个非零历史周期，按时间倒序。永不清零只有一个桶，无历史。
func (a *App) counterHistoryLocked(counter domain.Counter, now time.Time) []model.CounterHistoryItem {
	if counter.Period == domain.CounterNever {
		return nil
	}
	current := counter.PeriodKey(now)
	type entry struct {
		key   string
		count int
	}
	entries := make([]entry, 0, len(counter.Counts))
	for key, count := range counter.Counts {
		if key == current || count <= 0 {
			continue
		}
		entries = append(entries, entry{key, count})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].key > entries[j].key })
	if len(entries) > 7 {
		entries = entries[:7]
	}
	history := make([]model.CounterHistoryItem, 0, len(entries))
	for _, e := range entries {
		history = append(history, model.CounterHistoryItem{Label: a.periodLabelFor(e.key, counter.Period), Count: e.count})
	}
	return history
}

// periodLabelFor 把桶 key 格式化成展示文案：天→"08-13"、月→"2026-08"、年→"2026"。
func (a *App) periodLabelFor(key string, period domain.CounterPeriod) string {
	switch period {
	case domain.CounterMonth:
		return key
	case domain.CounterYear:
		return key
	default:
		if len(key) >= 5 {
			return key[5:]
		}
		return key
	}
}

func (a *App) AddCounter(name string, period string, goal int) string {
	counter := domain.Counter{ID: fmt.Sprintf("%d", a.now().UnixNano()), Name: name, Period: domain.CounterPeriod(period), Goal: goal, Counts: map[string]int{}}
	if err := counter.Validate(); err != nil {
		return err.Error()
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.counters = append(a.counters, counter)
	a.persistCountersLocked()
	return ""
}

func (a *App) UpdateCounter(id string, name string, period string, goal int) string {
	counter := domain.Counter{ID: id, Name: name, Period: domain.CounterPeriod(period), Goal: goal}
	if err := counter.Validate(); err != nil {
		return err.Error()
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.counters {
		if a.counters[i].ID == id {
			counter.Counts = a.counters[i].Counts
			a.counters[i] = counter
			a.persistCountersLocked()
			return ""
		}
	}
	return "counter not found"
}

func (a *App) DeleteCounter(id string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.counters {
		if a.counters[i].ID == id {
			a.counters = append(a.counters[:i], a.counters[i+1:]...)
			a.persistCountersLocked()
			return true
		}
	}
	return false
}

func (a *App) IncrementCounter(id string) int {
	return a.mutateCounter(id, func(counter *domain.Counter, now time.Time) {
		key := counter.PeriodKey(now)
		counter.Counts[key]++
	})
}

func (a *App) DecrementCounter(id string) int {
	return a.mutateCounter(id, func(counter *domain.Counter, now time.Time) {
		key := counter.PeriodKey(now)
		if counter.Counts[key] > 0 {
			counter.Counts[key]--
		}
	})
}

func (a *App) SetCounterCount(id string, count int) int {
	if count < 0 {
		count = 0
	}
	return a.mutateCounter(id, func(counter *domain.Counter, now time.Time) {
		counter.Counts[counter.PeriodKey(now)] = count
	})
}

// mutateCounter 对指定计数器的当前周期次数做就地变更并落盘，返回变更后的当前次数；计数器不存在时返回 -1。
func (a *App) mutateCounter(id string, mutate func(*domain.Counter, time.Time)) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.counters {
		if a.counters[i].ID == id {
			now := a.now()
			mutate(&a.counters[i], now)
			a.persistCountersLocked()
			return a.counters[i].CurrentCount(now)
		}
	}
	return -1
}

func (a *App) loadCountersLocked() {
	if a.countersPath == "" {
		return
	}
	file, err := store.LoadCounterFile(a.countersPath)
	if err != nil {
		return
	}
	a.counters = append([]domain.Counter(nil), file.Counters...)
	for i := range a.counters {
		if a.counters[i].Counts == nil {
			a.counters[i].Counts = map[string]int{}
		}
	}
}

func (a *App) persistCountersLocked() {
	if a.countersPath == "" {
		return
	}
	_ = store.SaveCounterFile(a.countersPath, store.CounterFile{
		SavedAt:  a.now(),
		Counters: append([]domain.Counter(nil), a.counters...),
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
	file, err := store.LoadTimelineFile(a.timelinePath)
	if err != nil {
		return
	}
	if file.Date != now.Format("2006-01-02") {
		return
	}
	a.timeline = append([]model.TimelineEntry(nil), file.Entries...)
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
	_ = store.SaveTimelineFile(a.timelinePath, store.TimelineFile{
		Date:    now.Format("2006-01-02"),
		SavedAt: now,
		Entries: append([]model.TimelineEntry(nil), a.timeline...),
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
		a.timeline = append(a.timeline, model.TimelineEntry{Kind: after.String(), StartedAt: at})
	}
	a.persistTimelineLocked(at)
}

// updateClient 是手动检查更新使用的 HTTP 客户端，包级变量便于测试替换。
var updateClient = &http.Client{Timeout: updateCheckTimeout}

// updateDownloadClient 是一键更新下载使用的 HTTP 客户端，超时远宽于检查请求。
var updateDownloadClient = &http.Client{Timeout: updateDownloadTimeout}

// CurrentVersion 返回运行时版本号（发布版本或 dev）。
func (a *App) CurrentVersion() string {
	return version
}

// CheckForUpdates 手动检查 GitHub Release 上的最新版本。dev 版本短路，不发起网络请求。
// 发现新版本时缓存资产下载地址，供 DownloadAndApplyUpdate 直接使用。
func (a *App) CheckForUpdates() model.UpdateCheckResult {
	if version == "" || version == "dev" {
		return model.UpdateCheckResult{
			Status:         model.UpdateStatusUpToDate,
			CurrentVersion: version,
			Message:        "当前为开发版本，不检查更新",
		}
	}
	result := checkForUpdates(updateClient, version, updateAPIBaseURL)
	a.mu.Lock()
	if result.Status == model.UpdateStatusAvailable {
		a.updateDownloadURL = result.DownloadURL
		a.updateLatestVersion = result.LatestVersion
	} else {
		a.updateDownloadURL = ""
		a.updateLatestVersion = ""
	}
	a.mu.Unlock()
	// 已是最新版本：清理遗留的待更新文件，保持磁盘事实与 UI 一致。
	if result.Status == model.UpdateStatusUpToDate {
		if exePath, err := os.Executable(); err == nil {
			cleanupPendingUpdateFiles(exePath)
		}
	}
	return result
}

// PendingUpdateInfo 返回 exe 同目录是否存在已下载待应用（.new）的更新及其版本号。
// dev 版本与非 Windows 平台一律返回不存在。
func (a *App) PendingUpdateInfo() model.PendingUpdateInfo {
	if version == "" || version == "dev" || goruntime.GOOS != "windows" {
		return model.PendingUpdateInfo{}
	}
	exePath, err := os.Executable()
	if err != nil {
		return model.PendingUpdateInfo{}
	}
	_, newPath, _, _ := exeUpdatePaths(exePath)
	if _, err := os.Stat(newPath); err != nil {
		return model.PendingUpdateInfo{}
	}
	return model.PendingUpdateInfo{
		Exists:  true,
		Version: pendingUpdateVersion(exePath),
	}
}

// DownloadAndApplyUpdate 一键更新：异步下载新版本 exe 到 exe 同目录的 .part，
// 完成后落位为 .new 并通过 update:progress 事件推送进度。dev 版本与非 Windows
// 平台短路不发起网络请求。返回空串表示已开始下载，否则返回同步失败文案。
func (a *App) DownloadAndApplyUpdate() string {
	if version == "" || version == "dev" {
		return "当前为开发版本，不支持自动更新"
	}
	if goruntime.GOOS != "windows" {
		return "当前平台不支持自动更新，请通过 GitHub 手动更新"
	}
	if a.ctx == nil {
		return "应用尚未就绪，请稍后重试"
	}
	a.mu.Lock()
	url := a.updateDownloadURL
	a.mu.Unlock()
	if url == "" {
		return "暂无可用更新，请先检查更新"
	}
	exePath, err := os.Executable()
	if err != nil {
		return "无法定位程序路径，无法自动更新"
	}
	if !dirWritable(filepath.Dir(exePath)) {
		return "程序目录不可写，请通过「前往 GitHub 查看」手动更新"
	}
	partPath, newPath, versionPath, _ := exeUpdatePaths(exePath)
	client := updateDownloadClient
	go func() {
		if err := downloadUpdate(a.ctx, client, url, partPath, newPath); err != nil {
			runtime.EventsEmit(a.ctx, updateProgressEvent, model.UpdateDownloadEvent{
				Phase:   model.UpdateDownloadPhaseError,
				Message: err.Error(),
			})
			return
		}
		// 记录 .new 对应的版本号，供确认弹窗与后续「重启更新」读取。
		a.mu.Lock()
		latest := a.updateLatestVersion
		a.mu.Unlock()
		if latest != "" {
			if err := os.WriteFile(versionPath, []byte(latest), 0o644); err != nil {
				runtime.EventsEmit(a.ctx, updateProgressEvent, model.UpdateDownloadEvent{
					Phase:   model.UpdateDownloadPhaseError,
					Message: "写入版本标记失败，请通过 GitHub 手动更新",
				})
				cleanupPendingUpdateFiles(exePath)
				return
			}
		}
		// 下载完成：弹原生确认框；确认则退出替换重启，取消则通知前端切换为「重启更新」。
		ok, err := confirmRestartDialog(a.ctx, pendingUpdateVersion(exePath))
		if err != nil {
			runtime.EventsEmit(a.ctx, updateProgressEvent, model.UpdateDownloadEvent{
				Phase:   model.UpdateDownloadPhaseError,
				Message: "确认对话框打开失败，请稍后重试",
			})
			return
		}
		if !ok {
			runtime.EventsEmit(a.ctx, updateProgressEvent, model.UpdateDownloadEvent{
				Phase:   model.UpdateDownloadPhaseCancelled,
				Message: "新版本已下载，可稍后重启",
			})
			return
		}
		exePath, err := os.Executable()
		if err != nil {
			runtime.EventsEmit(a.ctx, updateProgressEvent, model.UpdateDownloadEvent{
				Phase:   model.UpdateDownloadPhaseError,
				Message: "无法定位程序路径，请通过 GitHub 手动更新",
			})
			return
		}
		if msg := applyUpdateAndRestart(exePath, func() {
			a.quitRequested.Store(true)
			runtime.Quit(a.ctx)
		}); msg != "" {
			runtime.EventsEmit(a.ctx, updateProgressEvent, model.UpdateDownloadEvent{
				Phase:   model.UpdateDownloadPhaseError,
				Message: msg,
			})
		}
	}()
	return ""
}

// confirmRestartDialog 弹出「新版本已就绪」原生确认框，返回用户是否选择立即重启。
// 注意：Windows 端 MessageDialog 使用原生 MessageBox，固定 Yes/No 按钮并返回
// "Yes"/"No"（Buttons 自定义文本被忽略）；其他平台返回自定义按钮文本。
// 因此除明确的取消值外一律视为确认。
func confirmRestartDialog(ctx context.Context, latestVersion string) (bool, error) {
	message := "新版本已下载完成，是否立即重启生效？\n（Yes 立即重启，No 稍后处理）"
	if latestVersion != "" {
		message = "新版本 v" + latestVersion + " 已下载完成，是否立即重启生效？\n（Yes 立即重启，No 稍后处理）"
	}
	choice, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "更新就绪",
		Message:       message,
		Buttons:       []string{"立即重启", "稍后"},
		DefaultButton: "立即重启",
		CancelButton:  "稍后",
	})
	if err != nil {
		return false, err
	}
	switch choice {
	case "", "No", "Cancel", "稍后":
		return false, nil
	default:
		return true, nil
	}
}

// ApplyUpdateAndRestart 确认重启：校验 .new 就绪后弹确认框，确认则生成一次性更新脚本
// 并走完整退出流程（置位退出请求并调用 runtime.Quit，触发 shutdown 持久化）。
// 返回空串表示已进入重启流程；「已取消」表示用户选择稍后；其余为失败文案。
func (a *App) ApplyUpdateAndRestart() string {
	if version == "" || version == "dev" {
		return "当前为开发版本，不支持自动更新"
	}
	if a.ctx == nil {
		return "应用尚未就绪，请稍后重试"
	}
	exePath, err := os.Executable()
	if err != nil {
		return "无法定位程序路径，无法自动更新"
	}
	_, newPath, _, _ := exeUpdatePaths(exePath)
	if _, err := os.Stat(newPath); err != nil {
		return "未找到已下载的更新文件，请先点击「立即更新」"
	}
	ok, err := confirmRestartDialog(a.ctx, pendingUpdateVersion(exePath))
	if err != nil {
		return "确认对话框打开失败，请稍后重试"
	}
	if !ok {
		return "已取消"
	}
	return applyUpdateAndRestart(exePath, func() {
		a.quitRequested.Store(true)
		runtime.Quit(a.ctx)
	})
}

func (a *App) notifyReminder(restMinutes int) {
	a.notify(restMinutes)
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "reminder")
	}
}
