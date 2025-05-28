package json

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nimaaskarian/tom/timer"
)

type Daemon struct {
	router        *gin.Engine
	Timer         *timer.Timer
	TimerJsonChan chan string
}

func (d *Daemon) Init() {
	d.router = gin.Default()
	d.router.GET("/api/timer", func(c *gin.Context) {
		fmt.Println("SENDING Access-Control-Allow-Origin *. TURN OFF IN PRODUCTION")
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, d.Timer)
	})
	d.router.POST("/api/timer", func(c *gin.Context) {
		fmt.Println("SENDING Access-Control-Allow-Origin *. TURN OFF IN PRODUCTION")
		c.Header("Access-Control-Allow-Origin", "*")
		c.BindJSON(d.Timer)
		c.JSON(http.StatusOK, d.Timer)
	})
	d.router.GET("/api/timer/stream", func(c *gin.Context) {
		fmt.Println("SENDING Access-Control-Allow-Origin *. TURN OFF IN PRODUCTION")
		c.Header("Access-Control-Allow-Origin", "*")

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")
		c.Stream(func(w io.Writer) bool {
			if timer, ok := <-d.TimerJsonChan; ok {
        fmt.Println(timer)
				c.SSEvent("timer", timer)
				return true
			}
			return false
		})
	})
}

func (s *Daemon) Run(address string) error {
	s.router.Use(CORSMiddleware())
	return s.router.Run(address)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
