package timer

import (
	"context"
	"log"
	"log/slog"
	"strings"
	"time"
)

const VERSION = "v0.5.1"

type PomodoroTimerMode int

const (
	Pomodoro PomodoroTimerMode = iota
	ShortBreak
	LongBreak
	MODE_MAX
)

type PomodoroTimerState struct {
	Duration         time.Duration
	Mode             PomodoroTimerMode
	FinishedSessions uint
	Paused           bool
}

func (state *PomodoroTimerState) IsZero() bool {
	return !state.Paused &&
		state.Mode == Pomodoro &&
		state.Duration == 0 &&
		state.FinishedSessions == 0
}

type PomodoroTimer struct {
	Config *TimerConfig
	State  PomodoroTimerState
}

func (pt *PomodoroTimer) Reset() {
	slog.Info("timer reseted.", "new time", pt.Config.Duration[pt.State.Mode].String())
	pt.SeekTo(pt.Config.Duration[pt.State.Mode])
	if !pt.Config.OnSet.Run(pt) {
		pt.Config.OnChange.RunSync(pt)
	}
	if pt.State.Paused {
		pt.Config.OnPause.OnEventOnce = []func(*PomodoroTimer){func(pt *PomodoroTimer) {
			if !pt.State.Paused && !pt.Config.OnSet.Run(pt) {
				pt.Config.OnModeStart.Run(pt)
			}
		}}
	} else if !pt.Config.OnSet.Run(pt) {
		pt.Config.OnModeStart.Run(pt)
	}
}

func (pt *PomodoroTimer) Init() {
	pt.State.Mode = Pomodoro
	pt.State.FinishedSessions = 0
	pt.State.Paused = pt.Config.Paused
	pt.Reset()
	pt.Config.OnInit.Run(pt)
}

func (pt *PomodoroTimer) Pause(pauseValue bool) {
	pt.State.Paused = pauseValue
	if !pt.Config.OnSet.Run(pt) {
		pt.Config.OnPause.Run(pt)
	}
}

func (pt *PomodoroTimer) TogglePause() {
	pt.Pause(!pt.State.Paused)
}

func (pt *PomodoroTimer) SeekTo(duration time.Duration) {
	pt.State.Duration = duration
	if !pt.Config.OnSet.Run(pt) {
		pt.Config.OnChange.Run(pt)
	}
}

func (pt *PomodoroTimer) SeekAdd(duration time.Duration) {
	new_duration := pt.State.Duration + duration
	if new_duration < 0 {
		pt.SeekTo(0)
	} else {
		pt.SeekTo(new_duration)
	}
}

func (pt *PomodoroTimer) beforeTick() {
	if pt.State.Duration <= 0 {
		// timer before executing OnModeRun, so SwitchNextMode wouldn'pt
		// change the timer reference during the call.
		t_copy := *pt
		pt.Config.OnModeEnd.Run(&t_copy)
		pt.SwitchNextMode()
	}
}

func (pt *PomodoroTimer) tick() {
	if pt.State.Paused {
		return
	}
	pt.State.Duration -= pt.Config.DurationPerTick
	pt.Config.OnChange.Run(pt)
}

// Halts the current thread until ctx is Done. Use in a goroutine.
func (pt *PomodoroTimer) Loop(ctx context.Context) {
	slog.Info("timer loop started")
	timer := time.NewTimer(pt.Config.DurationPerTick)
	for {
		pt.beforeTick()
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			pt.tick()
			timer.Reset(pt.Config.DurationPerTick)
		}
	}
}

func (pt *PomodoroTimer) SwitchNextMode() {
	switch pt.State.Mode {
	case Pomodoro:
		pt.State.FinishedSessions++
		if pt.State.FinishedSessions >= pt.Config.Sessions {
			pt.State.Mode = LongBreak
		} else {
			pt.State.Mode = ShortBreak
		}
	case LongBreak:
		pt.Init()
		return
	case ShortBreak:
		pt.State.Mode = Pomodoro
	}
	pt.Reset()
}

func (pt *PomodoroTimer) SwitchPrevMode() {
	switch pt.State.Mode {
	case Pomodoro:
		if pt.State.FinishedSessions == 0 {
			pt.State.FinishedSessions = pt.Config.Sessions
			pt.State.Mode = LongBreak
		} else {
			pt.State.Mode = ShortBreak
		}
	case LongBreak:
		if pt.State.FinishedSessions > 0 {
			pt.State.FinishedSessions--
		}
		pt.State.Mode = Pomodoro
	case ShortBreak:
		if pt.State.FinishedSessions > 0 {
			pt.State.FinishedSessions--
		}
		pt.State.Mode = Pomodoro
	}
	pt.Reset()
}

func (pt *PomodoroTimer) String() string {
	rounded := pt.State.Duration.Round(time.Second)
	return rounded.String()
}

func (ptm PomodoroTimerMode) String() string {
	switch ptm {
	case Pomodoro:
		return "Pomodoro"
	case ShortBreak:
		return "Short Break"
	case LongBreak:
		return "Long Break"
	}
	log.Fatalf("Invalid pomodoro timer mode %d\n", ptm)
	return ""
}

func (ptm PomodoroTimerMode) SnakeCase() string {
	return strings.ReplaceAll(strings.ToLower(ptm.String()), " ", "_")
}

func (ptm PomodoroTimerMode) WormCase() string {
	return strings.ReplaceAll(strings.ToLower(ptm.String()), " ", "-")
}
