package httpd

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/nimaaskarian/goje/timer"
)

type Event struct {
	Name    string
	Payload any
}

type ClientsMap *sync.Map

type Daemon struct {
	router       *gin.Engine
	Timer        *timer.Timer
	lastId       uint
	ClosingIds   chan uint
	Clients      *sync.Map
}

func (d *Daemon) TimerEvent() Event {
	return Event{
		Name:    "timer",
		Payload: d.Timer,
	}
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
	// gin.SetMode(gin.ReleaseMode)
	d.router = gin.Default()
	d.router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	})
}

func (s *Daemon) Run(address string) error {
	return s.router.Run(address)
}
