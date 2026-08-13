package domain

import (
	"errors"
	"time"
	"unicode/utf8"
)

type CounterPeriod string

const (
	CounterDay   CounterPeriod = "day"
	CounterMonth CounterPeriod = "month"
	CounterYear  CounterPeriod = "year"
	CounterNever CounterPeriod = "never"
)

// Counter 计数器：按重置周期把次数存进计数桶 map，换周期自然开新桶。
type Counter struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Period CounterPeriod  `json:"period"`
	Goal   int            `json:"goal"` // 0 = 不设目标
	Counts map[string]int `json:"counts"`
}

func (c Counter) Validate() error {
	if utf8.RuneCountInString(c.Name) == 0 {
		return errors.New("name is required")
	}
	if utf8.RuneCountInString(c.Name) > MaxTitleLength {
		return errors.New("name too long")
	}
	switch c.Period {
	case CounterDay, CounterMonth, CounterYear, CounterNever:
	default:
		return errors.New("unknown counter period")
	}
	if c.Goal < 0 {
		return errors.New("goal must not be negative")
	}
	return nil
}

// PeriodKey 当前周期对应的计数桶 key。
func (c Counter) PeriodKey(t time.Time) string {
	switch c.Period {
	case CounterMonth:
		return t.Format("2006-01")
	case CounterYear:
		return t.Format("2006")
	case CounterNever:
		return "all"
	default:
		return t.Format("2006-01-02")
	}
}

// CurrentCount 当前周期已累计的次数。
func (c Counter) CurrentCount(t time.Time) int {
	return c.Counts[c.PeriodKey(t)]
}

// PeriodLabel 周期的中文文案。
func (c Counter) PeriodLabel() string {
	switch c.Period {
	case CounterDay:
		return "今日"
	case CounterMonth:
		return "本月"
	case CounterYear:
		return "今年"
	default:
		return "累计"
	}
}
