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

type Settings struct {
	ReminderMinutes int `json:"reminderMinutes"`
	RestMinutes     int `json:"restMinutes"`
}

func DefaultSettings() Settings {
	return Settings{
		ReminderMinutes: int(domain.DefaultReminderDuration / time.Minute),
		RestMinutes:     int(domain.DefaultRestDuration / time.Minute),
	}
}
