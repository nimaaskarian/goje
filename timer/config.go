package timer

import (
	"time"
)

type TimerConfig struct {
	Sessions        uint
	Duration        [MODE_MAX]time.Duration
	OnModeEnd       TimerConfigEvent `json:"-"`
	OnModeStart     TimerConfigEvent `json:"-"`
	OnChange        TimerConfigEvent `json:"-"`
	OnPause         TimerConfigEvent `json:"-"`
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

type TimerConfigEvent struct {
	OnEvent     []func(*Timer)
  // list of functions to be run only on the next event (only once)
	OnEventOnce []func(*Timer)
}

func (e *TimerConfigEvent) Append(handler func(*Timer)) {
  e.OnEvent= append(e.OnEvent, handler)
}


// append function to be run only on the next event
func (e *TimerConfigEvent) AppendOnce(handler func(*Timer)) {
  e.OnEventOnce = append(e.OnEventOnce, handler)
}

func (e *TimerConfigEvent) Run(t *Timer) {
	for _, handler := range append(e.OnEvent, e.OnEventOnce...) {
		handler(t)
	}
	e.OnEventOnce = nil
}

