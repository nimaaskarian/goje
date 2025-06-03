package httpd

import (
	"embed"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/nimaaskarian/goje/timer"
)

//go:embed webgui-preact/dist/*
var embed_fs embed.FS

func (d *Daemon) JsonRoutes() {
	d.router.GET("/api/timer", NoCache, func(c *gin.Context) {
		c.JSON(http.StatusOK, d.Timer)
	})
	d.router.POST("/api/timer/nextmode", NoCache, func(c *gin.Context) {
		d.Timer.SwitchNextMode()
		c.JSON(http.StatusOK, d.Timer)
	})
	d.router.POST("/api/timer/pause", NoCache, func(c *gin.Context) {
		d.Timer.Pause()
		c.JSON(http.StatusOK, d.Timer)
	})
	d.router.POST("/api/timer/reset", NoCache, func(c *gin.Context) {
		d.Timer.Reset()
		c.JSON(http.StatusOK, d.Timer)
	})
	d.router.POST("/api/timer/prevmode", NoCache, func(c *gin.Context) {
		d.Timer.SwitchPrevMode()
		c.JSON(http.StatusOK, d.Timer)
	})
	d.router.POST("/api/timer", NoCache, func(c *gin.Context) {
		prev_mode := d.Timer.Mode
		if err := c.BindJSON(d.Timer); err == nil {
			if prev_mode != d.Timer.Mode {
				d.Timer.Reset()
			}
			d.UpdateClients(timer.OnChangeEvent(d.Timer))
			c.JSON(http.StatusOK, d.Timer)
		}
	})
	d.router.GET("/api/timer/stream", NoCache, func(c *gin.Context) {
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

func NoCache(c *gin.Context) {
	c.Header("Cache-Control", "no-cache")
}

func (d *Daemon) WebguiRoutes(custom_css_file string) {
	static, _ := fs.Sub(embed_fs, "webgui-preact/dist/assets")
	d.router.StaticFS("/assets", http.FS(static))
	if custom_css_file != "" {
		d.router.GET("/custom.css", NoCache, func(c *gin.Context) {
			data, _ := os.ReadFile(custom_css_file)
			c.Data(http.StatusOK, "text/css;  charset=utf-8", data)
		})
	}
	data, _ := embed_fs.ReadFile("webgui-preact/dist/index.html")
	d.router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html;  charset=utf-8", data)
	})
}
