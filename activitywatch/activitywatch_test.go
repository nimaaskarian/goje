package activitywatch

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	aw_go "github.com/nimaaskarian/aw-go"
	"github.com/nimaaskarian/goje/timer"
)

func TestWatcherDuration(t *testing.T) {
	tomato := timer.Timer{}
	tomato.Config = &timer.DefaultConfig
	tomato.Config.Duration = [timer.MODE_MAX]time.Duration{
		time.Second * 2,
		time.Second,
		time.Second * 3,
	}
  tomato.Config.Sessions = 2
	client := aw_go.ActivityWatchClient{
		Config: aw_go.ActivityWatchClientConfig{
			Protocol: "http",
			Hostname: "127.0.0.1",
			Port:     "5612",
		},
		ClientName: "goje-pomodoro-watcher",
	}
	client.Init()
	watcher := Watcher{
		client:    client,
		bucket_id: "aw-watcher-goje_" + client.ClientHostname,
	}

	watcher.AddEventWatchers(tomato.Config)
	router := gin.Default()
	router.POST("/api/0/buckets/:bucket_id/events", func(c *gin.Context) {
		event := aw_go.Event{}
		if err := c.BindJSON(&event); err != nil {
			t.Fatal(err)
		}
    mode := tomato.Mode
    if time.Duration(event.Duration) != tomato.Config.Duration[mode] {
			t.Fatal("duration mismatch", time.Duration(event.Duration), tomato.Config.Duration[mode], mode)
    }
	})
	go router.Run(client.Config.Hostname + ":" + client.Config.Port)
	tomato.Init()
	go tomato.Loop()
	time.Sleep(time.Second * 9)
}
