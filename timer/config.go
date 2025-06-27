package timer

import (
	"time"
)

type TimerConfig struct {
	Sessions        uint                    `mapstructure:"sessions"`
	Duration        [MODE_MAX]time.Duration `mapstructure:"duration"`
	OnModeEnd       TimerConfigEvent        `json:"-"`
	OnModeStart     TimerConfigEvent        `json:"-"`
	OnChange        TimerConfigEvent        `json:"-"`
	OnPause         TimerConfigEvent        `json:"-"`
	OnQuit          TimerConfigEvent        `json:"-"`
	Paused          bool                    `mapstructure:"paused"`
	DurationPerTick time.Duration           `mapstructure:"duration-per-tick"`
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
	OnEvent     []func(*Timer)
	OnEventOnce []func(*Timer)
}

func (e *TimerConfigEvent) Append(handler func(*Timer)) {
	e.OnEvent = append(e.OnEvent, handler)
}

// append function to be run only on the next event
func (e *TimerConfigEvent) AppendOnce(handler func(*Timer)) {
	e.OnEventOnce = append(e.OnEventOnce, handler)
}

// this is non-blocking. it iterates through all the events and goroutines them.
func (e *TimerConfigEvent) Run(t *Timer) {
	for _, handler := range append(e.OnEvent, e.OnEventOnce...) {
		go handler(t)
	}
	e.OnEventOnce = nil
}
