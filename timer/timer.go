package timer

import (
	"log/slog"
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
	if !t.Config.OnSet.Run(t) {
		t.Config.OnChange.Run(t)
	}
	if t.Paused {
		t.Config.OnPause.OnEventOnce = []func(*Timer){func(t *Timer) {
			if !t.Paused && !t.Config.OnSet.Run(t) {
				t.Config.OnModeStart.Run(t)
			}
		}}
	} else if !t.Config.OnSet.Run(t){
		t.Config.OnModeStart.Run(t)
	}
}

func (t *Timer) Init() {
	t.Mode = Pomodoro
	t.FinishedSessions = 0
	t.Paused = t.Config.Paused
	t.Reset()
}

func (t *Timer) Pause() {
	t.Paused = !t.Paused
	if !t.Config.OnSet.Run(t) {
		t.Config.OnPause.Run(t)
	}
}

func (t *Timer) SeekTo(duration time.Duration) {
	t.Duration = duration
	if !t.Config.OnSet.Run(t) {
		t.Config.OnChange.Run(t)
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
	if t.Duration <= 0 {
		// timer before executing OnModeRun, so SwitchNextMode wouldn't
		// change the timer reference during the call.
		t_copy := *t
		t.Config.OnModeEnd.Run(&t_copy)
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
	slog.Info("timer loop started")
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
		if t.FinishedSessions > 0 {
			t.FinishedSessions--
		}
		t.Mode = Pomodoro
	case ShortBreak:
		if t.FinishedSessions > 0 {
			t.FinishedSessions--
		}
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
