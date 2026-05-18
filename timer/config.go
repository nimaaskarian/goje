package timer

import (
	"time"
)

type TimerConfigHooks struct {
	OnModeEnd   TimerConfigHook `json:"-"`
	OnModeStart TimerConfigHook `json:"-"`
	OnChange    TimerConfigHook `json:"-"`
	// event for clients of an outbound server, to push sets to server. all change
	// events should run this first, if failed then run other events
	OnSet   TimerConfigHook `json:"-"`
	OnPause TimerConfigHook `json:"-"`
	OnQuit  TimerConfigHook `json:"-"`
	OnInit  TimerConfigHook `json:"-"`
}

type TimerConfig struct {
	Sessions        uint                    `mapstructure:"sessions,omitempty"`
	Duration        [MODE_MAX]time.Duration `mapstructure:"duration"`
	Hooks           TimerConfigHooks        `json:"-"`
	Paused          bool                    `mapstructure:"paused,omitempty"`
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
type TimerConfigHook struct {
	OnEvent     []func(*PomodoroTimer)
	OnEventOnce []func(*PomodoroTimer)
}

func (e *TimerConfigHook) Append(handler func(*PomodoroTimer)) {
	e.OnEvent = append(e.OnEvent, func(pt *PomodoroTimer) { go handler(pt) })
}

// this is non-blocking (goroutine). it iterates through all the events and goroutines them.
func (e *TimerConfigHook) Run(t *PomodoroTimer) (ran bool) {
	for _, handler := range append(e.OnEvent, e.OnEventOnce...) {
		go handler(t)
		ran = true
	}
	e.OnEventOnce = nil
	return ran
}

// blocking version of Run(). uses no goroutines
func (e *TimerConfigHook) RunSync(t *PomodoroTimer) (ran bool) {
	for _, handler := range append(e.OnEvent, e.OnEventOnce...) {
		handler(t)
		ran = true
	}
	e.OnEventOnce = nil
	return ran
}
