package json

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nimaaskarian/tom/timer"
)

type Daemon struct {
	router *gin.Engine
	Timer  *timer.Timer
}

func (d *Daemon) Init() {
	d.router = gin.Default()
	d.router.GET("/api/timer", func(c *gin.Context) {
		c.JSON(http.StatusOK, d.Timer)
	})
	d.router.POST("/api/timer", func(c *gin.Context) {
    c.BindJSON(d.Timer)
	})
}

func (s *Daemon) Run(address string) error {
	return s.router.Run(address)
}
