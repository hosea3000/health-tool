package domain

import (
	"errors"
	"time"
	"unicode/utf8"
)

const MaxTitleLength = 20

type RuleType string

const (
	RuleDate     RuleType = "date"
	RuleMonthly  RuleType = "monthly"
	RuleWeekly   RuleType = "weekly"
	RuleBiweekly RuleType = "biweekly"
)

type CountdownEvent struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Rule  Rule   `json:"rule"`
}

type Rule struct {
	Type    RuleType `json:"type"`
	Target  string   `json:"target,omitempty"`  // date: YYYY-MM-DD
	Day     int      `json:"day,omitempty"`     // monthly: 1..31
	Weekday int      `json:"weekday,omitempty"` // weekly/biweekly: 0=周一..6=周日
	Phase   string   `json:"phase,omitempty"`   // biweekly: big|small
	Anchor  string   `json:"anchor,omitempty"`  // biweekly: YYYY-MM-DD
}

func (e CountdownEvent) Validate() error {
	if utf8.RuneCountInString(e.Title) == 0 {
		return errors.New("title is required")
	}
	if utf8.RuneCountInString(e.Title) > MaxTitleLength {
		return errors.New("title too long")
	}
	return e.Rule.Validate()
}

func (r Rule) Validate() error {
	switch r.Type {
	case RuleDate:
		if _, err := time.Parse("2006-01-02", r.Target); err != nil {
			return errors.New("target date must be YYYY-MM-DD")
		}
	case RuleMonthly:
		if r.Day < 1 || r.Day > 31 {
			return errors.New("day must be between 1 and 31")
		}
	case RuleWeekly:
		if r.Weekday < 0 || r.Weekday > 6 {
			return errors.New("weekday must be between 0 and 6")
		}
	case RuleBiweekly:
		if r.Weekday < 0 || r.Weekday > 6 {
			return errors.New("weekday must be between 0 and 6")
		}
		if r.Phase != "big" && r.Phase != "small" {
			return errors.New("phase must be big or small")
		}
		if _, err := time.Parse("2006-01-02", r.Anchor); err != nil {
			return errors.New("anchor date must be YYYY-MM-DD")
		}
	default:
		return errors.New("unknown rule type")
	}
	return nil
}

// NextOccurrence 返回 today 或之后的下一次到期日（日期粒度）。
func (r Rule) NextOccurrence(now time.Time) time.Time {
	today := dateOnly(now)
	switch r.Type {
	case RuleDate:
		t, err := time.ParseInLocation("2006-01-02", r.Target, today.Location())
		if err != nil {
			return time.Time{}
		}
		return t
	case RuleMonthly:
		return nextMonthly(r.Day, today)
	case RuleWeekly:
		target := toGoWeekday(r.Weekday)
		diff := (int(target) - int(today.Weekday()) + 7) % 7
		return today.AddDate(0, 0, diff)
	case RuleBiweekly:
		return nextBiweekly(r, today)
	default:
		return time.Time{}
	}
}

// toGoWeekday 将周一=0..周日=6 转换为 time.Weekday（周日=0..周六=6）。
func toGoWeekday(w int) time.Weekday {
	return time.Weekday((w + 1) % 7)
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func lastDayOfMonth(y int, m time.Month) int {
	return time.Date(y, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// nextMonthly 下一月的第 day 天，短月钳位到该月最后一天。
func nextMonthly(day int, today time.Time) time.Time {
	y, m := today.Year(), today.Month()
	candidate := monthlyDate(y, m, day, today.Location())
	if !candidate.Before(today) {
		return candidate
	}
	nm := m + 1
	ny := y
	if nm > 12 {
		nm = 1
		ny++
	}
	return monthlyDate(ny, nm, day, today.Location())
}

func monthlyDate(y int, m time.Month, day int, loc *time.Location) time.Time {
	last := lastDayOfMonth(y, m)
	if day > last {
		day = last
	}
	return time.Date(y, m, day, 0, 0, 0, 0, loc)
}

// nextBiweekly 相位匹配（大周或小周）的最近周几。
// 锚周固定为大周：事件相位为"big"落在与锚周同相位的周（本周），
// "small"落在与锚周反相位的周（下周），之后每两周一次。
func nextBiweekly(r Rule, today time.Time) time.Time {
	anchor, err := time.ParseInLocation("2006-01-02", r.Anchor, today.Location())
	if err != nil {
		anchor = today
	}
	anchor = dateOnly(anchor)
	target := toGoWeekday(r.Weekday)
	for d := 0; d < 14; d++ {
		candidate := today.AddDate(0, 0, d)
		if candidate.Weekday() != target {
			continue
		}
		daysSince := int(candidate.Sub(anchor) / (24 * time.Hour))
		isBig := floorDiv(daysSince, 7)%2 == 0
		if (r.Phase == "big") == isBig {
			return candidate
		}
	}
	return time.Time{}
}

func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}
