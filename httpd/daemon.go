package httpd

import (
	"github.com/gin-gonic/gin"
	"github.com/nimaaskarian/goje/timer"
)

type Event struct {
	Name    string
	Payload any
}

type ClientsMap map[uint]chan Event

type Daemon struct {
	router  *gin.Engine
	Timer   *timer.Timer
	lastId  uint
	Clients ClientsMap
}

func (d *Daemon) TimerEvent() Event {
  return Event {
    Name: "timer",
    Payload: d.Timer,
  }
}

func (d *Daemon) UpdateClients(e Event) {
  for _, client := range d.Clients {
    client <- e
  }
}

func (d *Daemon) Init() {
	d.router = gin.Default()
	d.router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	})
}

func (s *Daemon) Run(address string) error {
	return s.router.Run(address)
}
