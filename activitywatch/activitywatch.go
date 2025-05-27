package activitywatch

import (
	"time"

	"github.com/nimaaskarian/aw-go"
	"github.com/nimaaskarian/tom/timer"
)

const BUCKET_ID = "aw-watcher-tom_nimas-pc"
const EVENT_TYPE = "pomodoro_status"

func SetupTimerConfig(config * timer.TimerConfig) {
	client := aw_go.ActivityWatchClient{
		Config: aw_go.ActivityWatchClientConfig{
			Protocol: "http",
			Hostname: "127.0.0.1",
			Port:     "5600",
		},
		ClientName: "tom-pomodoro-watcher",
	}
	client.Init()
  client.CreateBucket(BUCKET_ID, EVENT_TYPE)
	for mode := range timer.MODE_MAX {
		config.OnModeEnd[mode] = func(t *timer.Timer) {
			duration := config.Duration[mode]
			start := time.Now().Add(-duration).UTC()
      mode_string := mode.String()
			event := aw_go.Event{
				Duration:  aw_go.SecondsDuration(duration),
				Timestamp: aw_go.IsoTime(start),
				Data: map[string]any{
					"status": mode_string,
					"title": mode_string,
				},
			}
			client.InsertEvent(BUCKET_ID, event)
		}
	}
}
