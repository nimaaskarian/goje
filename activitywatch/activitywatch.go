package activitywatch

import (
	"time"

	"github.com/nimaaskarian/aw-go"
	"github.com/nimaaskarian/goje/timer"
)

const EVENT_TYPE = "pomodoro_status"

type Data struct {
	paused_start time.Time
	started      time.Time
	client       aw_go.ActivityWatchClient
	bucket_id    string
}

func (d *Data) Init() {
	d.client = aw_go.ActivityWatchClient{
		Config: aw_go.ActivityWatchClientConfig{
			Protocol: "http",
			Hostname: "127.0.0.1",
			Port:     "5600",
		},
		ClientName: "goje-pomodoro-watcher",
	}
	d.client.Init()
	d.bucket_id = "aw-watcher-goje_" + d.client.ClientHostname
	d.client.CreateBucket(d.bucket_id, EVENT_TYPE)
}

func (d *Data) pushCurrentMode(t *timer.Timer, now time.Time) {
	duration := now.UTC().Sub(d.started)
	mode_string := t.Mode.String()
	event := aw_go.Event{
		Duration:  aw_go.SecondsDuration(duration),
		Timestamp: aw_go.IsoTime(d.started),
		Data: map[string]any{
			"status": mode_string,
			"title":  mode_string,
		},
	}
	d.client.InsertEvent(d.bucket_id, event)
}

func (d *Data) AddEventWatchers(config *timer.TimerConfig) {
	config.OnModeStart = append(config.OnModeStart, func(t *timer.Timer) {
		d.started = time.Now().UTC()
	})
	config.OnModeEnd = append(config.OnModeEnd, func(t *timer.Timer) {
		d.pushCurrentMode(t, time.Now().UTC())
	})
	config.OnPause = append(config.OnPause, func(t *timer.Timer) {
		if t.Paused {
      now := time.Now().UTC()
			d.paused_start = now
			d.pushCurrentMode(t, now)
		} else {
			now := time.Now().UTC()
			d.started = now
			duration := now.Sub(d.paused_start)
			event := aw_go.Event{
				Duration:  aw_go.SecondsDuration(duration),
				Timestamp: aw_go.IsoTime(d.paused_start),
				Data: map[string]any{
					"status": "Paused",
					"title":  "Paused",
				},
			}
			d.client.InsertEvent(d.bucket_id, event)
		}
	})
}
