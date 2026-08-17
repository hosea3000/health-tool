package model

import (
	"time"

	"health-tool/domain"
)

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

type CountdownView struct {
	ID            string      `json:"id"`
	Title         string      `json:"title"`
	Rule          domain.Rule `json:"rule"`
	NextDate      string      `json:"nextDate"`
	RemainingDays int         `json:"remainingDays"`
}

type CounterView struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Period      domain.CounterPeriod `json:"period"`
	PeriodLabel string               `json:"periodLabel"`
	Goal        int                  `json:"goal"`
	Count       int                  `json:"count"`
	GoalReached bool                 `json:"goalReached"`
	History     []CounterHistoryItem `json:"history"`
}

type CounterHistoryItem struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// 更新检查结果状态常量。
const (
	UpdateStatusUpToDate  = "up-to-date"
	UpdateStatusAvailable = "update-available"
	UpdateStatusError     = "error"
)

type Settings struct {
	ReminderMinutes int `json:"reminderMinutes"`
	RestMinutes     int `json:"restMinutes"`
	// AutoStart 反映开机自启动的实际状态，由注册表实时读取填充，不持久化到 settings.json。
	AutoStart bool `json:"autoStart,omitempty"`
}

type UpdateCheckResult struct {
	// Status 取值：up-to-date / update-available / error。
	Status         string `json:"status"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseURL     string `json:"releaseUrl"`
	// DownloadURL 是 health-tool.exe 资产的下载地址；资产缺失时为空。
	DownloadURL string `json:"downloadUrl"`
	Message     string `json:"message"`
}

// 更新下载事件阶段常量。
const (
	UpdateDownloadPhaseDownloading = "downloading"
	UpdateDownloadPhaseCompleted   = "completed"
	UpdateDownloadPhaseCancelled   = "cancelled"
	UpdateDownloadPhaseError       = "error"
)

// UpdateDownloadEvent 是一次更新下载的进度事件，通过 Wails events（update:progress）推送给前端。
type UpdateDownloadEvent struct {
	// Phase 取值：downloading / completed / cancelled / error。
	Phase      string `json:"phase"`
	Downloaded int64  `json:"downloaded"`
	Total      int64  `json:"total"`
	Percent    int    `json:"percent"`
	// Message 为错误阶段的提示文案。
	Message string `json:"message"`
}

// PendingUpdateInfo 描述已下载待应用（.new 就绪）的更新状态。
type PendingUpdateInfo struct {
	// Exists 表示 exe 同目录是否存在 .new 待应用文件。
	Exists bool `json:"exists"`
	// Version 是 .new 对应的版本号（来自 .new.version 元数据），读取失败时为空。
	Version string `json:"version"`
}

func DefaultSettings() Settings {
	return Settings{
		ReminderMinutes: int(domain.DefaultReminderDuration / time.Minute),
		RestMinutes:     int(domain.DefaultRestDuration / time.Minute),
	}
}
