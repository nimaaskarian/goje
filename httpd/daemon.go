package httpd

import (
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/gzip"
	"github.com/nimaaskarian/goje/timer"
)

type Daemon struct {
	router     *gin.Engine
	Timer      *timer.Timer
	lastId     uint
	ClosingIds chan uint
	Clients    *sync.Map
}

func (d *Daemon) SetupEvents() {
	d.Timer.Config.OnChange.Append(func (t *timer.Timer) {
		d.BroadcastToSSEClients(ChangeEvent(t))
	})
	d.Timer.Config.OnModeStart.Append(func(t *timer.Timer) {
		d.BroadcastToSSEClients(NewEvent(t, "start"))
	})
	d.Timer.Config.OnModeEnd.Append(func(t *timer.Timer) {
		d.BroadcastToSSEClients(NewEvent(t, "end"))
	})
	d.Timer.Config.OnPause.Append(func(t *timer.Timer) {
		d.BroadcastToSSEClients(NewEvent(t, "pause"))
	})
}

func (d *Daemon) BroadcastToSSEClients(e Event) {
	d.Clients.Range(func(id any, value any) (closed bool) {
		client := value.(chan Event)
		defer func() {
			if recover() != nil {
				closed = false
			}
		}()
		client <- e
		return true
	})
}

func (d *Daemon) Init() {
	gin.SetMode(gin.ReleaseMode)
	d.router = gin.Default()
	d.router.Use(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Writer.Header().Set("Cache-Control", "public, max-age=31536000")
		}
	}, gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{"/api"})))
}

func (d *Daemon) Run(address string) error {
	return d.router.Run(address)
}
