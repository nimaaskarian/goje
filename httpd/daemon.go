package httpd

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/nimaaskarian/goje/timer"
)

type Daemon struct {
	engine     *gin.Engine
	Timer      *timer.PomodoroTimer
	lastId     uint
	ClosingIds chan uint
	Clients    *sync.Map
}

func (d *Daemon) SetupEvents() {
	d.Timer.Config.OnChange.Append(func(t *timer.PomodoroTimer) {
		d.BroadcastToSSEClients(ChangeEvent(t))
	})
	d.Timer.Config.OnModeStart.Append(func(t *timer.PomodoroTimer) {
		d.BroadcastToSSEClients(NewEvent(t, "start"))
	})
	d.Timer.Config.OnModeEnd.Append(func(t *timer.PomodoroTimer) {
		d.BroadcastToSSEClients(NewEvent(t, "end"))
	})
	d.Timer.Config.OnPause.Append(func(t *timer.PomodoroTimer) {
		d.BroadcastToSSEClients(NewEvent(t, "pause"))
	})
}

func (d *Daemon) BroadcastToSSEClients(e Event) {
	d.Clients.Range(func(id any, value any) bool {
		client := value.(chan Event)
		client <- e
		return true
	})
}

func (d *Daemon) Init() {
	gin.SetMode(gin.ReleaseMode)
	d.engine = gin.Default()
	d.engine.Use(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Writer.Header().Set("Cache-Control", "public, max-age=31536000")
		}
	}, gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{"/api"})))
}

func (d *Daemon) Run(address, certfile, keyfile string, ctx context.Context) {
	httpServer := &http.Server{
		Addr:    address,
		Handler: d.engine.Handler(),
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	if certfile != "" && keyfile != "" {
		slog.Info("running https daemon", "address", address, "certfile", certfile, "keyfile", keyfile)
		go func() {
			if err := httpServer.ListenAndServeTLS(certfile, keyfile); err != nil && err != http.ErrServerClosed {
				slog.Error("https server failed", "err", err)
			}
		}()
	} else {
		slog.Info("running http daemon", "address", address)
		go func() {
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("http server failed", "err", err)
			}
		}()
	}

	<-ctx.Done()
	d.BroadcastToSSEClients(Event{Name: "restart"})
	slog.Info("shutting http server down...")
	ctx = context.Background()
	httpServer.Shutdown(ctx)
}
