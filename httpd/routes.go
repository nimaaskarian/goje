package httpd

import (
	"embed"
	"io"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nimaaskarian/goje/timer"
)

//go:embed webgui-preact/dist/*
var embed_fs embed.FS

func (d *Daemon) JsonRoutes() {
	d.router.GET("/api/timer", func(c *gin.Context) {
		c.JSON(http.StatusOK, d.Timer)
	})
	d.router.POST("/api/timer/nextmode", func(c *gin.Context) {
		d.Timer.SwitchNextMode()
		c.JSON(http.StatusOK, d.Timer)
	})
	d.router.POST("/api/timer/pause", func(c *gin.Context) {
		d.Timer.Pause()
		c.JSON(http.StatusOK, d.Timer)
	})
	d.router.POST("/api/timer/reset", func(c *gin.Context) {
		d.Timer.Reset()
		c.JSON(http.StatusOK, d.Timer)
	})
	d.router.POST("/api/timer/prevmode", func(c *gin.Context) {
		d.Timer.SwitchPrevMode()
		c.JSON(http.StatusOK, d.Timer)
	})
	d.router.POST("/api/timer", func(c *gin.Context) {
		prev_mode := d.Timer.Mode
		if err := c.BindJSON(d.Timer); err == nil {
			if prev_mode != d.Timer.Mode {
				d.Timer.Reset()
			}
			d.UpdateClients(timer.OnChangeEvent(d.Timer))
			c.JSON(http.StatusOK, d.Timer)
		}
	})
	d.router.GET("/api/timer/stream", func(c *gin.Context) {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		c.Header("Content-Type", "text/event-stream")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")
		client := make(chan Event, 1)
		client <- timer.OnChangeEvent(d.Timer)
		d.lastId++
		id := d.lastId
		d.Clients.Store(id, client)
		defer func() {
			d.Clients.Delete(id)
			close(client)
		}()
		c.Stream(func(w io.Writer) bool {
			if event, ok := <-client; ok {
				c.SSEvent(event.Name(), event.Payload())
				return true
			}
			return false
		})
	})
}

func (d *Daemon) WebguiRoutes(custom_css_file string) {
	static, _ := fs.Sub(embed_fs, "webgui-preact/dist/assets")
	d.router.StaticFS("/assets", http.FS(static))
	if custom_css_file != "" {
		d.router.StaticFile("/custom.css", custom_css_file)
	}
	data, _ := embed_fs.ReadFile("webgui-preact/dist/index.html")
	d.router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html;  charset=utf-8", data)
	})
}
