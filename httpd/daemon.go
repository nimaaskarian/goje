package httpd

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/nimaaskarian/goje/timer"
)

type ClientsMap *sync.Map

type Daemon struct {
	router       *gin.Engine
	Timer        *timer.Timer
	lastId       uint
	ClosingIds   chan uint
	Clients      *sync.Map
}

func (d *Daemon) UpdateAllChangeEvent(timer *timer.Timer) {
  d.UpdateClients(d.ChangeEvent())
}

func (d *Daemon) SetupEndStartEvents() {
  d.Timer.Config.OnModeStart = append(d.Timer.Config.OnModeStart, func(t *timer.Timer) {
    d.UpdateClients(timer.OnModeStartEvent(t))
  })
  d.Timer.Config.OnModeEnd = append(d.Timer.Config.OnModeEnd, func(t *timer.Timer) {
    d.UpdateClients(timer.OnChangeEvent(t))
  })
}

func (d *Daemon) ChangeEvent() timer.Event {
	return timer.Event{
		Name:    "change",
		Payload: d.Timer,
	}
}

func (d *Daemon) UpdateClients(e timer.Event) {
  d.Clients.Range(func(id any, value any) (closed bool) {
    client := value.(chan timer.Event)
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
	// gin.SetMode(gin.ReleaseMode)
	d.router = gin.Default()
	d.router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	})
}

func (d *Daemon) Run(address string) error {
	return d.router.Run(address)
}
