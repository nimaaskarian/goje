package httpd
import (
	"github.com/nimaaskarian/tom/timer"
	"github.com/gin-gonic/gin"
)

type Daemon struct {
	router        *gin.Engine
	Timer         *timer.Timer
	TimerJsonChan chan string
}

func (d *Daemon) Init() {
	d.router = gin.Default()
}

func (s *Daemon) Run(address string) error {
	return s.router.Run(address)
}
