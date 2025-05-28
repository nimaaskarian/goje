package httpd

import (
	"embed"
	"fmt"
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
	d.router.POST("/api/timer", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
		c.BindJSON(d.Timer)
		c.JSON(http.StatusOK, d.Timer)
	})
	d.router.GET("/api/timer/stream", func(c *gin.Context) {
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

func (d *Daemon) WebguiRoutes() {
	static, _ := fs.Sub(embed_fs, "webgui-preact/dist/assets")
	d.router.StaticFS("/assets", http.FS(static))
	d.router.GET("/", func (c *gin.Context) {
    data, _ := embed_fs.ReadFile("webgui-preact/dist/index.html")
    c.Data(http.StatusOK, "text/html;  charset=utf-8", data)
  })
}
