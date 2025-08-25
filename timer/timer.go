package timer

import (
	"context"
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

type PomdoroTimerState struct {
	Duration         time.Duration
	Mode             TimerMode
	FinishedSessions uint
	Paused           bool
}

func (state *PomdoroTimerState) IsZero() bool {
	return !state.Paused &&
		state.Mode == Pomodoro &&
		state.Duration == 0 &&
		state.FinishedSessions == 0
}

type PomodoroTimer struct {
	Config *TimerConfig
	State  PomdoroTimerState
}

func (t *PomodoroTimer) Reset() {
	slog.Info("timer reseted. new time is %s", t.Config.Duration[t.State.Mode].String())
	t.SeekTo(t.Config.Duration[t.State.Mode])
	if !t.Config.OnSet.Run(t) {
		t.Config.OnChange.Run(t)
	}
	if t.State.Paused {
		t.Config.OnPause.OnEventOnce = []func(*PomodoroTimer){func(t *PomodoroTimer) {
			if !t.State.Paused && !t.Config.OnSet.Run(t) {
				t.Config.OnModeStart.Run(t)
			}
		}}
	} else if !t.Config.OnSet.Run(t) {
		t.Config.OnModeStart.Run(t)
	}
}

func (t *PomodoroTimer) Init() {
	t.State.Mode = Pomodoro
	t.State.FinishedSessions = 0
	t.State.Paused = t.Config.Paused
	t.Reset()
}

func (t *PomodoroTimer) Pause() {
	t.State.Paused = !t.State.Paused
	if !t.Config.OnSet.Run(t) {
		t.Config.OnPause.Run(t)
	}
}

func (t *PomodoroTimer) SeekTo(duration time.Duration) {
	t.State.Duration = duration
	if !t.Config.OnSet.Run(t) {
		t.Config.OnChange.Run(t)
	}
}

func (t *PomodoroTimer) SeekAdd(duration time.Duration) {
	new_duration := t.State.Duration + duration
	if new_duration < 0 {
		t.SeekTo(0)
	} else {
		t.SeekTo(new_duration)
	}
}

func (t *PomodoroTimer) beforeTick() {
	if t.State.Duration <= 0 {
		// timer before executing OnModeRun, so SwitchNextMode wouldn't
		// change the timer reference during the call.
		t_copy := *t
		t.Config.OnModeEnd.Run(&t_copy)
		t.SwitchNextMode()
	}
}

func (t *PomodoroTimer) tick() {
	if t.State.Paused {
		return
	}
	t.State.Duration -= t.Config.DurationPerTick
	t.Config.OnChange.Run(t)
}

// Halts the current thread for ever. Use in a go routine.
func (t *PomodoroTimer) Loop(ctx context.Context) {
	slog.Info("timer loop started")
	timer := time.NewTimer(t.Config.DurationPerTick)
	for {
		t.beforeTick()
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			t.tick()
			timer.Reset(t.Config.DurationPerTick)
		}
	}
}

func (t *PomodoroTimer) SwitchNextMode() {
	switch t.State.Mode {
	case Pomodoro:
		t.State.FinishedSessions++
		if t.State.FinishedSessions >= t.Config.Sessions {
			t.State.Mode = LongBreak
		} else {
			t.State.Mode = ShortBreak
		}
	case LongBreak:
		t.Init()
		return
	case ShortBreak:
		t.State.Mode = Pomodoro
	}
	t.Reset()
}

func (t *PomodoroTimer) SwitchPrevMode() {
	switch t.State.Mode {
	case Pomodoro:
		if t.State.FinishedSessions == 0 {
			t.State.FinishedSessions = t.Config.Sessions
			t.State.Mode = LongBreak
		} else {
			t.State.Mode = ShortBreak
		}
	case LongBreak:
		if t.State.FinishedSessions > 0 {
			t.State.FinishedSessions--
		}
		t.State.Mode = Pomodoro
	case ShortBreak:
		if t.State.FinishedSessions > 0 {
			t.State.FinishedSessions--
		}
		t.State.Mode = Pomodoro
	}
	t.Reset()
}

func (t *PomodoroTimer) String() string {
	rounded := t.State.Duration.Round(time.Second)
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
