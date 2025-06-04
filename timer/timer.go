package timer

import (
	"strings"
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
	Sessions        uint
	Duration        [MODE_MAX]time.Duration
	OnModeEnd       []func(*Timer) `json:"-"`
	OnModeStart     []func(*Timer) `json:"-"`
	OnChange        []func(*Timer) `json:"-"`
	OnPause         []func(*Timer) `json:"-"`
	Paused          bool
	DurationPerTick time.Duration `mapstructure:"duration-per-tick"`
}

var DefaultConfig = TimerConfig{
	Sessions: 4,
	Duration: [MODE_MAX]time.Duration{
		25 * time.Minute,
		5 * time.Minute,
		30 * time.Minute,
	},
	DurationPerTick: time.Second,
	Paused:          false,
}

type Timer struct {
	Config           *TimerConfig
	Duration         time.Duration
	Mode             TimerMode
	FinishedSessions uint
	Paused           bool
}

func (t *Timer) Reset() {
	t.SeekTo(t.Config.Duration[t.Mode])
	t.onStart()
}

func (t *Timer) onStart() {
	if t.Config.OnModeStart != nil {
		for _, onStart := range t.Config.OnModeStart {
			onStart(t)
		}
	}
}

func (t *Timer) SetConfig(config *TimerConfig) {
	t.Config = config
}

func (t *Timer) Init() {
  t.naiveInit()
  t.Reset()
}

// naive init inits the timer to initial state without setting the current time
// nor firing the start event
func (t *Timer) naiveInit() {
	t.Mode = Pomodoro
	t.FinishedSessions = 0
	t.Paused = t.Config.Paused
  if t.Paused {
    t.OnPause()
  }
}

func (t *Timer) OnChange() {
	if t.Config.OnChange != nil {
		for _, onChange := range t.Config.OnChange {
			onChange(t)
		}
	}
}

func (t *Timer) Pause() {
  t.Paused = !t.Paused
  t.OnPause()
}

func (t *Timer) OnPause() {
	if t.Config.OnPause != nil {
    for _, onPause := range t.Config.OnPause {
      onPause(t)
    }
  }
}

func (t *Timer) SeekTo(duration time.Duration) {
	t.Duration = duration
	t.OnChange()
}

func (t *Timer) SeekAdd(duration time.Duration) {
	new_duration := t.Duration + duration
	if new_duration < 0 {
		t.SeekTo(0)
	} else {
		t.SeekTo(new_duration)
	}
}

func (t *Timer) onEnd() {
	if t.Config.OnModeEnd != nil {
		for _, onEnd := range t.Config.OnModeEnd {
			onEnd(t)
		}
	}
}

func (t *Timer) tick() {
	if t.Duration <= 0 {
		t.onEnd()
		t.SwitchNextMode()
	}
	time.Sleep(t.Config.DurationPerTick)
	if t.Paused {
		return
	}
	t.Duration -= t.Config.DurationPerTick
	t.OnChange()
}

// Halts the current thread for ever. Use in a go routine.
func (t *Timer) Loop() {
	for {
		t.tick()
	}
}

func (t *Timer) SwitchNextMode() {
	switch t.Mode {
	case Pomodoro:
		t.FinishedSessions++
		if t.FinishedSessions >= t.Config.Sessions {
			t.Mode = LongBreak
		} else {
			t.Mode = ShortBreak
		}
	case LongBreak:
		t.naiveInit()
	case ShortBreak:
		t.Mode = Pomodoro
	}
	t.Reset()
}

func (t *Timer) SwitchPrevMode() {
	switch t.Mode {
	case Pomodoro:
		if t.FinishedSessions == 0 {
			t.FinishedSessions = t.Config.Sessions
			t.Mode = LongBreak
		} else {
			t.Mode = ShortBreak
		}
	case LongBreak:
		t.FinishedSessions--
		t.Mode = Pomodoro
	case ShortBreak:
		t.FinishedSessions--
		t.Mode = Pomodoro
	}
	t.Reset()
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

func (tm TimerMode) SnakeCase() string {
	return strings.ReplaceAll(strings.ToLower(tm.String()), " ", "_")
}

func (tm TimerMode) WormCase() string {
	return strings.ReplaceAll(strings.ToLower(tm.String()), " ", "-")
}
