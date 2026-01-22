package timer

import (
	"time"
)

type TimerConfig struct {
	Sessions    uint                    `mapstructure:"sessions,omitempty"`
	Duration    [MODE_MAX]time.Duration `mapstructure:"duration"`
	OnModeEnd   TimerConfigEvent        `json:"-"`
	OnModeStart TimerConfigEvent        `json:"-"`
	OnChange    TimerConfigEvent        `json:"-"`
	// event for clients of an outbound server, to push sets to server. all change
	// events should run this first, if failed then run other events
	OnSet           TimerConfigEvent `json:"-"`
	OnPause         TimerConfigEvent `json:"-"`
	OnQuit          TimerConfigEvent `json:"-"`
	OnInit          TimerConfigEvent `json:"-"`
	Paused          bool             `mapstructure:"paused,omitempty"`
	DurationPerTick time.Duration    `mapstructure:"duration-per-tick"`
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

// OnEventOnce functions run only on the next event (only once)
type TimerConfigEvent struct {
	OnEvent     []func(*PomodoroTimer)
	OnEventOnce []func(*PomodoroTimer)
}

func (e *TimerConfigEvent) Append(handler func(*PomodoroTimer)) {
	e.OnEvent = append(e.OnEvent, handler)
}

// append function to be run only on the next event
func (e *TimerConfigEvent) AppendOnce(handler func(*PomodoroTimer)) {
	e.OnEventOnce = append(e.OnEventOnce, handler)
}

// this is non-blocking (goroutine). it iterates through all the events and goroutines them.
func (e *TimerConfigEvent) Run(t *PomodoroTimer) (ran bool) {
	for _, handler := range append(e.OnEvent, e.OnEventOnce...) {
		go handler(t)
		ran = true
	}
	e.OnEventOnce = nil
	return ran
}

// blocking version of Run(). uses no goroutines
func (e *TimerConfigEvent) RunSync(t *PomodoroTimer) (ran bool) {
	for _, handler := range append(e.OnEvent, e.OnEventOnce...) {
		handler(t)
		ran = true
	}
	e.OnEventOnce = nil
	return ran
}
