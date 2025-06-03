package httpd

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/nimaaskarian/goje/timer"
)

type Daemon struct {
	router     *gin.Engine
	Timer      *timer.Timer
	lastId     uint
	ClosingIds chan uint
	Clients    *sync.Map
}

func (d *Daemon) UpdateAllChangeEvent(t *timer.Timer) {
	d.UpdateClients(timer.OnChangeEvent(t))
}

func (d *Daemon) SetupEndStartEvents() {
	d.Timer.Config.OnModeStart = append(d.Timer.Config.OnModeStart, func(t *timer.Timer) {
		d.UpdateClients(timer.OnModeStartEvent(t))
	})
	d.Timer.Config.OnModeEnd = append(d.Timer.Config.OnModeEnd, func(t *timer.Timer) {
		d.UpdateClients(timer.OnChangeEvent(t))
	})
	d.Timer.Config.OnPause = append(d.Timer.Config.OnPause, func(t *timer.Timer) {
		d.UpdateClients(timer.OnPauseEvent(t))
	})
}

type Event interface {
	Payload() any
	Name() string
}

func (d *Daemon) UpdateClients(e Event) {
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
		c.Writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	})
}

func (d *Daemon) Run(address string) error {
	return d.router.Run(address)
}
