package domain

import "time"

const (
	idleAfter               = 5 * time.Minute
	DefaultReminderDuration = 60 * time.Minute
	MinReminderDuration     = 1 * time.Minute
	MaxReminderDuration     = 180 * time.Minute
	ReminderDurationStep    = 5 * time.Minute
	DefaultRestDuration     = 3 * time.Minute
	MinRestDuration         = 1 * time.Minute
	MaxRestDuration         = 30 * time.Minute
)

// DurationFromMinutes 将分钟数换算为 time.Duration。
func DurationFromMinutes(minutes int) time.Duration {
	return time.Duration(minutes) * time.Minute
}

type State uint8

const (
	Waiting State = iota
	Working
	IdlePaused
	Resting
)

func (s State) String() string {
	switch s {
	case Waiting:
		return "waiting"
	case Working:
		return "working"
	case IdlePaused:
		return "idle-paused"
	case Resting:
		return "resting"
	default:
		return "unknown"
	}
}

type ActivityKind uint8

const (
	KeyPress ActivityKind = iota
	Click
	Wheel
	MouseMove
)

type EffectiveActivity struct {
	Kind ActivityKind
	At   time.Time
}

type Result struct {
	Changed  bool
	Reminder bool
}

type Monitor struct {
	state                State
	startedAt            time.Time
	lastActivity         time.Time
	restStartedAt        time.Time
	reminderAfter        time.Duration
	restAfter            time.Duration
	activeReminderAfter  time.Duration
	activeRestAfter      time.Duration
	notificationsEnabled bool
}

func NewMonitor(now time.Time, reminderAfter time.Duration, restAfter ...time.Duration) *Monitor {
	rest := DefaultRestDuration
	if len(restAfter) > 0 {
		rest = restAfter[0]
	}
	if !ValidReminderDuration(reminderAfter) {
		reminderAfter = DefaultReminderDuration
	}
	if !ValidRestDuration(rest) {
		rest = DefaultRestDuration
	}
	return &Monitor{state: Waiting, lastActivity: now, reminderAfter: reminderAfter, restAfter: rest, activeReminderAfter: reminderAfter, activeRestAfter: rest, notificationsEnabled: true}
}

func ValidReminderDuration(duration time.Duration) bool {
	return duration >= MinReminderDuration && duration <= MaxReminderDuration &&
		(duration == MinReminderDuration || duration%ReminderDurationStep == 0)
}

func ValidRestDuration(duration time.Duration) bool {
	return duration >= MinRestDuration && duration <= MaxRestDuration
}

func (m *Monitor) State() State {
	return m.state
}

func (m *Monitor) Elapsed(now time.Time) time.Duration {
	if m.state != Working || now.Before(m.startedAt) {
		return 0
	}
	return now.Sub(m.startedAt)
}

func (m *Monitor) RestRemaining(now time.Time) time.Duration {
	if m.state != Resting {
		return 0
	}
	remaining := m.activeRestAfter - now.Sub(m.restStartedAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (m *Monitor) SetReminderDuration(duration time.Duration) bool {
	if !ValidReminderDuration(duration) {
		return false
	}
	m.reminderAfter = duration
	return true
}

func (m *Monitor) SetRestDuration(duration time.Duration) bool {
	if !ValidRestDuration(duration) {
		return false
	}
	m.restAfter = duration
	return true
}

// SetNotificationsEnabled 设置久坐提醒通知开关，关闭立即生效（当前工作段不再触发提醒）。
// 开启时若当前处于工作段，本轮提醒时点从当前时刻重新计算
// （顺延为「已工作时长 + 提醒时长」），避免开启后立刻补弹提醒；
// 非 Working 状态下开启无需处理，下一个工作段创建时按设置值正常计时。
func (m *Monitor) SetNotificationsEnabled(enabled bool, now time.Time) {
	m.notificationsEnabled = enabled
	if enabled && m.state == Working && !now.Before(m.startedAt) {
		m.activeReminderAfter = now.Sub(m.startedAt) + m.reminderAfter
	}
}

func (m *Monitor) NotificationsEnabled() bool {
	return m.notificationsEnabled
}

func (m *Monitor) EffectiveActivity(activity EffectiveActivity) Result {
	if !validActivity(activity) {
		return Result{}
	}
	if m.state == Resting {
		result := m.Advance(activity.At)
		if m.state == Resting {
			return result
		}
	}
	if m.state == IdlePaused {
		m.state = Waiting
	}
	if m.state != Waiting && m.state != Working {
		return Result{}
	}
	if m.state == Working && activity.At.Before(m.lastActivity) {
		return Result{}
	}

	result := m.Advance(activity.At)
	if m.state == Resting {
		return result
	}
	if m.state == Waiting {
		m.state = Working
		m.startedAt = activity.At
		m.activeReminderAfter = m.reminderAfter
		m.activeRestAfter = m.restAfter
		result.Changed = true
	}
	m.lastActivity = activity.At
	return result
}

func (m *Monitor) Advance(now time.Time) Result {
	if m.state == Resting {
		if now.Sub(m.restStartedAt) >= m.activeRestAfter {
			m.state = Waiting
			return Result{Changed: true}
		}
		return Result{}
	}
	if m.state != Working || now.Before(m.lastActivity) {
		return Result{}
	}
	if now.Sub(m.lastActivity) >= idleAfter {
		m.state = IdlePaused
		return Result{Changed: true}
	}
	if now.Sub(m.startedAt) >= m.activeReminderAfter {
		// 通知关闭（静默记录模式）：到点不进入提醒休息期、不发送提醒，
		// 工作段保持计时持续增长，状态转换交给真实的闲置或退出。
		if !m.notificationsEnabled {
			return Result{}
		}
		m.state = Resting
		m.restStartedAt = now
		return Result{Changed: true, Reminder: true}
	}
	return Result{}
}

func (m *Monitor) PauseForIdle() bool {
	if m.state != Waiting && m.state != Working {
		return false
	}
	m.state = IdlePaused
	return true
}

func validActivity(activity EffectiveActivity) bool {
	if activity.At.IsZero() {
		return false
	}
	switch activity.Kind {
	case KeyPress, Click, Wheel, MouseMove:
		return true
	default:
		return false
	}
}
