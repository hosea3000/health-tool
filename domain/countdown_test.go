package domain

import (
	"testing"
	"time"
)

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("parse %s: %v", s, err)
	}
	return d
}

func TestNextOccurrenceDate(t *testing.T) {
	rule := Rule{Type: RuleDate, Target: "2026-12-31"}
	if got := rule.NextOccurrence(mustDate(t, "2026-08-12")); got != mustDate(t, "2026-12-31") {
		t.Errorf("expected 2026-12-31, got %s", got)
	}
	if got := rule.NextOccurrence(mustDate(t, "2027-01-02")); got != mustDate(t, "2026-12-31") {
		t.Errorf("past date returns fixed target, got %s", got)
	}
}

func TestNextOccurrenceWeekly(t *testing.T) {
	rule := Rule{Type: RuleWeekly, Weekday: 2} // 周三
	cases := []struct {
		now, want string
	}{
		{"2026-08-12", "2026-08-12"}, // 当天是周三
		{"2026-08-13", "2026-08-19"}, // 周四 → 下周三
		{"2026-08-16", "2026-08-19"}, // 周日 → 下周三
	}
	for _, c := range cases {
		if got := rule.NextOccurrence(mustDate(t, c.now)); got != mustDate(t, c.want) {
			t.Errorf("weekly %s: expected %s, got %s", c.now, c.want, got)
		}
	}
}

func TestNextOccurrenceMonthlyClamp(t *testing.T) {
	day31 := Rule{Type: RuleMonthly, Day: 31}
	cases := []struct {
		now, want string
	}{
		{"2026-01-15", "2026-01-31"},
		{"2026-01-31", "2026-01-31"}, // 当天即到期
		{"2026-02-20", "2026-02-28"}, // 2 月钳位（非闰年）
		{"2026-03-15", "2026-03-31"},
		{"2026-04-20", "2026-04-30"}, // 4 月钳位
		{"2026-04-30", "2026-04-30"}, // 钳位到期日当天
		{"2026-05-01", "2026-05-31"},
		{"2026-12-20", "2026-12-31"},
	}
	for _, c := range cases {
		if got := day31.NextOccurrence(mustDate(t, c.now)); got != mustDate(t, c.want) {
			t.Errorf("monthly31 %s: expected %s, got %s", c.now, c.want, got)
		}
	}

	leap := Rule{Type: RuleMonthly, Day: 31}
	if got := leap.NextOccurrence(mustDate(t, "2028-02-20")); got != mustDate(t, "2028-02-29") {
		t.Errorf("leap feb: expected 2028-02-29, got %s", got)
	}

	day30 := Rule{Type: RuleMonthly, Day: 30}
	if got := day30.NextOccurrence(mustDate(t, "2026-02-20")); got != mustDate(t, "2026-02-28") {
		t.Errorf("monthly30 feb: expected 2026-02-28, got %s", got)
	}

	// 月末之后指向次月
	if got := day31.NextOccurrence(mustDate(t, "2026-05-01")); got != mustDate(t, "2026-05-31") {
		t.Errorf("after clamped month-end: expected 2026-05-31, got %s", got)
	}
}

func TestNextOccurrenceBiweekly(t *testing.T) {
	bigSat := Rule{Type: RuleBiweekly, Weekday: 5, Phase: "big", Anchor: "2026-08-10"} // 锚周一为大周，周六
	cases := []struct {
		now, want string
	}{
		{"2026-08-12", "2026-08-15"}, // 大周内周三 → 本周六
		{"2026-08-15", "2026-08-15"}, // 当天即到期
		{"2026-08-16", "2026-08-29"}, // 大周周日 → 跳过小周周六
		{"2026-08-22", "2026-08-29"}, // 小周周六 → 下周六（大周）
		{"2026-08-29", "2026-08-29"}, // 大周周六当天
		{"2026-09-06", "2026-09-12"}, // 小周周日 → 大周周六
	}
	for _, c := range cases {
		if got := bigSat.NextOccurrence(mustDate(t, c.now)); got != mustDate(t, c.want) {
			t.Errorf("biweekly big/sat %s: expected %s, got %s", c.now, c.want, got)
		}
	}

	smallSun := Rule{Type: RuleBiweekly, Weekday: 6, Phase: "small", Anchor: "2026-08-10"} // 锚周一所在周固定为大周
	casesSmall := []struct {
		now, want string
	}{
		{"2026-08-12", "2026-08-23"}, // 锚周（大周）内 → 下一个小周周日
		{"2026-08-23", "2026-08-23"}, // 当天（小周周日）
		{"2026-08-24", "2026-09-06"}, // 大周周一 → 跳过大周周日，落在小周周日
	}
	for _, c := range casesSmall {
		if got := smallSun.NextOccurrence(mustDate(t, c.now)); got != mustDate(t, c.want) {
			t.Errorf("biweekly small/sun %s: expected %s, got %s", c.now, c.want, got)
		}
	}
}

func TestValidate(t *testing.T) {
	ok := []CountdownEvent{
		{Title: "发薪日", Rule: Rule{Type: RuleDate, Target: "2026-12-31"}},
		{Title: "每个月最后一天", Rule: Rule{Type: RuleMonthly, Day: 31}},
		{Title: "周三例会", Rule: Rule{Type: RuleWeekly, Weekday: 2}},
		{Title: "大小周周六", Rule: Rule{Type: RuleBiweekly, Weekday: 5, Phase: "big", Anchor: "2026-08-10"}},
		{Title: "二十个汉字就是刚刚好二十个汉字啊", Rule: Rule{Type: RuleDate, Target: "2026-12-31"}},
	}
	for _, e := range ok {
		if err := e.Validate(); err != nil {
			t.Errorf("expected valid %+v, got %v", e, err)
		}
	}

	bad := []CountdownEvent{
		{Title: "", Rule: Rule{Type: RuleDate, Target: "2026-12-31"}},
		{Title: "超长标题一二三四五六七八九十一二三四五六七八九十啊", Rule: Rule{Type: RuleDate, Target: "2026-12-31"}},
		{Title: "x", Rule: Rule{Type: RuleDate, Target: "12/31/2026"}},
		{Title: "x", Rule: Rule{Type: RuleMonthly, Day: 0}},
		{Title: "x", Rule: Rule{Type: RuleMonthly, Day: 32}},
		{Title: "x", Rule: Rule{Type: RuleWeekly, Weekday: -1}},
		{Title: "x", Rule: Rule{Type: RuleWeekly, Weekday: 7}},
		{Title: "x", Rule: Rule{Type: RuleBiweekly, Weekday: 5, Phase: "huge", Anchor: "2026-08-10"}},
		{Title: "x", Rule: Rule{Type: RuleBiweekly, Weekday: 5, Phase: "big", Anchor: "08/10/2026"}},
		{Title: "x", Rule: Rule{Type: "fortnightly", Weekday: 5}},
	}
	for _, e := range bad {
		if err := e.Validate(); err == nil {
			t.Errorf("expected invalid %+v", e)
		}
	}
}
