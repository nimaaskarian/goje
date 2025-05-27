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
	AfterTick          func(*Timer)
	AfterSeek          func(*Timer)
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
	config       TimerConfig
	Duration     time.Duration
	Mode         TimerMode
	SessionCount uint
	Paused       bool
}

func (t *Timer) Reset() {
  t.SeekTo(t.config.Duration[t.Mode])
}

func (t *Timer) SetConfig(config TimerConfig) {
  t.config = config
}

func (t *Timer) Init() {
	t.Mode = Pomodoro
	t.SessionCount = t.config.SessionCount
	t.Paused = t.config.Paused
  t.Reset()
}

func (t *Timer) SeekTo(duration time.Duration) {
	t.Duration = duration
  if t.config.AfterSeek != nil {
    t.config.AfterSeek(t)
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
		time.Sleep(t.config.DurationPerTick)
		return
	}
	if t.Duration <= 0 {
		t.CycleMode()
	}
	time.Sleep(t.config.DurationPerTick)
	t.Duration -= t.config.DurationPerTick
	if t.config.AfterTick != nil {
		t.config.AfterTick(t)
	}
}

// Halts the current thread for ever. Use in a go routine.
func (t *Timer) Loop() {
	for {
		t.tick()
	}
}

func (t *Timer) CycleMode() {
	if t.config.OnModeEnd[t.Mode] != nil {
		t.config.OnModeEnd[t.Mode](t)
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
	if t.config.OnModeStart[t.Mode] != nil {
		t.config.OnModeStart[t.Mode](t)
	}
	t.Duration = t.config.Duration[t.Mode]
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
