package activitywatch

import (
	"fmt"
	"time"

	"github.com/nimaaskarian/aw-go"
	"github.com/nimaaskarian/goje/timer"
)

const EVENT_TYPE = "pomodoro_status"

func SetupTimerConfig(config * timer.TimerConfig) {
	client := aw_go.ActivityWatchClient{
		Config: aw_go.ActivityWatchClientConfig{
			Protocol: "http",
			Hostname: "127.0.0.1",
			Port:     "5600",
		},
		ClientName: "goje-pomodoro-watcher",
	}
	client.Init()
  bucket_id := fmt.Sprintf("aw-watcher-goje_%s", client.ClientHostname)
  client.CreateBucket(bucket_id, EVENT_TYPE)
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
			client.InsertEvent(bucket_id, event)
		}
	}
}
