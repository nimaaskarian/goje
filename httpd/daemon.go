package httpd

import (
	"github.com/gin-gonic/gin"
	"github.com/nimaaskarian/tom/timer"
)

type Event struct {
	Name    string
	Payload any
}

type Daemon struct {
	router  *gin.Engine
	Timer   *timer.Timer
	lastId  int
	Clients map[int]chan Event
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
