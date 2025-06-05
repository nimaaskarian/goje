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

type Timer struct {
	Config           *TimerConfig
	Duration         time.Duration
	Mode             TimerMode
	FinishedSessions uint
	Paused           bool
}

func (t *Timer) Reset() {
	t.SeekTo(t.Config.Duration[t.Mode])
	t.Config.OnModeStart.Run(t)
}

func (t *Timer) SetConfig(config *TimerConfig) {
	t.Config = config
}

func (t *Timer) Init() {
	t.Mode = Pomodoro
	t.FinishedSessions = 0
	t.Paused = t.Config.Paused
	t.SeekTo(t.Config.Duration[t.Mode])
	if t.Paused {
		t.Config.OnPause.Run(t)
		t.Config.OnPause.AppendOnce(func(t *Timer) {
			if !t.Paused {
				t.Config.OnModeStart.Run(t)
			}
		})
	} else {
		t.Config.OnModeStart.Run(t)
	}
}

func (t *Timer) Pause() {
	t.Paused = !t.Paused
	t.Config.OnPause.Run(t)
}

func (t *Timer) SeekTo(duration time.Duration) {
	t.Duration = duration
	t.Config.OnChange.Run(t)
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
	if t.Duration <= 0 {
		t.Config.OnModeEnd.Run(t)
		t.SwitchNextMode()
	}
	time.Sleep(t.Config.DurationPerTick)
	if t.Paused {
		return
	}
	t.Duration -= t.Config.DurationPerTick
	t.Config.OnChange.Run(t)
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
		t.Init()
		return
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
