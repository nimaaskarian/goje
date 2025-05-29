package httpd

import (
	"embed"
	"io"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed webgui-preact/dist/*
var embed_fs embed.FS

func (d *Daemon) JsonRoutes() {
	d.router.GET("/api/timer", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
		c.JSON(http.StatusOK, d.Timer)
	})
	d.router.POST("/api/timer/nextmode", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
    d.Timer.SwitchNextMode()
    d.UpdateClients(d.TimerEvent())
    c.JSON(http.StatusOK, d.Timer)
	})
	d.router.POST("/api/timer/prevmode", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
    d.Timer.SwitchPrevMode()
    d.UpdateClients(d.TimerEvent())
    c.JSON(http.StatusOK, d.Timer)
	})
	d.router.POST("/api/timer", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
    if err := c.BindJSON(d.Timer); err == nil {
      d.UpdateClients(d.TimerEvent())
      c.JSON(http.StatusOK, d.Timer)
    }
	})
	d.router.GET("/api/timer/stream", SSEMiddleware, func(c *gin.Context) {
		client := make(chan Event, 1)
		client <- Event{Payload: d.Timer, Name: "timer"}
		d.lastId++
		id := d.lastId
		d.Clients[id] = client
		defer delete(d.Clients, id)
		c.Stream(func(w io.Writer) bool {
			if event, ok := <-client; ok {
				c.SSEvent(event.Name, event.Payload)
				return true
			}
			return false
		})
	})
}

func SSEMiddleware(c *gin.Context) {
  c.Header("Content-Type", "text/event-stream")
  c.Header("Cache-Control", "no-cache")
  c.Header("Connection", "keep-alive")
  c.Header("Transfer-Encoding", "chunked")
}

func (d *Daemon) WebguiRoutes() {
	static, _ := fs.Sub(embed_fs, "webgui-preact/dist/assets")
	d.router.StaticFS("/assets", http.FS(static))
  data, _ := embed_fs.ReadFile("webgui-preact/dist/index.html")
	d.router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html;  charset=utf-8", data)
	})
}
