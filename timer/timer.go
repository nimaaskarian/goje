package timer

import (
	"time"
)

type TimerMode int

const (
	Pomodoro TimerMode = iota
	ShortBreak
	LongBreak
	MODE_MAX
)

type TimerConfig struct {
	SessionCount    uint
	Duration        [MODE_MAX]time.Duration
	OnModeEnd       [MODE_MAX]func(*Timer)
	OnModeStart     [MODE_MAX]func(*Timer)
	OnTick          func(*Timer)
	OnSeek          func(*Timer)
	Paused          bool
	DurationPerTick time.Duration
}

var DefaultConfig = TimerConfig{
	SessionCount: 4,
	Duration: [MODE_MAX]time.Duration{
		25 * time.Minute,
		5 * time.Minute,
		30 * time.Minute,
	},
	DurationPerTick: time.Second,
	Paused:          false,
}

type Timer struct {
	Config       TimerConfig
	Duration     time.Duration
	Mode         TimerMode
	SessionCount uint
	Paused       bool
}

func (t *Timer) Reset() {
  t.SeekTo(t.Config.Duration[t.Mode])
}

func (t *Timer) Init() {
	t.Mode = Pomodoro
	t.SessionCount = t.Config.SessionCount
	t.Paused = t.Config.Paused
  t.Reset()
}

func (t *Timer) SeekTo(duration time.Duration) {
	t.Duration = duration
  if t.Config.OnSeek != nil {
    t.Config.OnSeek(t)
  }
}

func (t *Timer) SeekAdd(duration time.Duration) {
	new_duration := t.Duration + duration
	if new_duration < 0 {
		t.SeekTo(0)
	} else {
		t.SeekTo(new_duration)
	}
}

func (t *Timer) tick() {
	if t.Paused {
		time.Sleep(t.Config.DurationPerTick)
		return
	}
	if t.Duration == 0 {
		t.CycleMode()
	}
	time.Sleep(t.Config.DurationPerTick)
	t.Duration -= t.Config.DurationPerTick
	if t.Config.OnTick != nil {
		t.Config.OnTick(t)
	}
}

// Halts the current thread for ever. Use in a go routine.
func (t *Timer) Loop() {
	for {
		t.tick()
	}
}

func (t *Timer) CycleMode() {
	if t.Config.OnModeEnd[t.Mode] != nil {
		t.Config.OnModeEnd[t.Mode](t)
	}
	switch t.Mode {
	case Pomodoro:
		t.SessionCount--
		if t.SessionCount == 0 {
			t.Mode = LongBreak
		} else {
			t.Mode = ShortBreak
		}
	case LongBreak:
		t.Init()
	case ShortBreak:
		t.Mode = Pomodoro
	}
	if t.Config.OnModeStart[t.Mode] != nil {
		t.Config.OnModeStart[t.Mode](t)
	}
	t.Duration = t.Config.Duration[t.Mode]
}

func (t *Timer) String() string {
	rounded := t.Duration.Round(time.Second)
	return rounded.String()
}

func (tm TimerMode) String() string {
	switch tm {
	case Pomodoro:
		return "Pomodoro"
	case ShortBreak:
		return "Short Break"
	case LongBreak:
		return "Long Break"
	}
	return "unknown"
}
