package ntfy

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/nimaaskarian/goje/timer"
	"github.com/nimaaskarian/goje/utils"
)

func Setup(ntfyAddress string, timerConfig timer.TimerConfig) {
	ntfyAddress = utils.FixHttpAddress(ntfyAddress)
	timerConfig.OnInit.Append((func(pt *timer.PomodoroTimer) {
		if req, err := ntfyRequest(ntfyAddress, "Timer init!", "tomato,arrow_forward"); err == nil {
			if _, err := http.DefaultClient.Do(req); err != nil {
				slog.Error("Failed to send ntfy request", "err", err)
			}
		}
	}))
	timerConfig.OnModeStart.Append((func(pt *timer.PomodoroTimer) {
		var msg, tags string
		switch pt.State.Mode {
		case 0:
			msg = "Pomodoro started!"
			tags = "tomato"
		case 1:
			msg = "Short break!"
			tags = "coffee"
		case 2:
			msg = "Long break!"
			tags = "tropical_drink"
		}
		if req, err := ntfyRequest(ntfyAddress, msg, tags); err == nil {
			if _, err := http.DefaultClient.Do(req); err != nil {
				slog.Error("Failed to send ntfy request", "err", err)
			}
		}
	}))
	timerConfig.OnPause.Append(func(pt *timer.PomodoroTimer) {
		var msg, tags string
		if pt.State.Paused {
			msg = "Timer paused!"
			tags = "pause_button"
		} else {
			msg = "Timer unpaused!"
			tags = "arrow_forward"
		}
		if req, err := ntfyRequest(ntfyAddress, msg, tags); err == nil {
			if _, err := http.DefaultClient.Do(req); err != nil {
				slog.Error("Failed to send ntfy request", "err", err)
			}
		}
	})
	if timerConfig.Paused {
		timerConfig.OnModeEnd.Append((func(pt *timer.PomodoroTimer) {
			if pt.State.Mode == 2 {
				if req, err := ntfyRequest(ntfyAddress, "Long break ended!", "tomato"); err == nil {
					if _, err := http.DefaultClient.Do(req); err != nil {
						slog.Error("Failed to send ntfy request", "err", err)
					}
				}
			}
		}))
	}
}

// helper function that creates a request
func ntfyRequest(address, content, tags string) (*http.Request, error) {
	req, err := http.NewRequest("POST", address, strings.NewReader(content))
	if err != nil {
		slog.Error("Failed to create ntfy request", "err", err)
		return nil, err
	}
	req.Header.Set("Tags", tags)
	return req, nil
}
