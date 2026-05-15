package cmd

import (
	"encoding/base64"
	"log/slog"
	"net/http"
	"strings"

	"github.com/nimaaskarian/goje/timer"
	"github.com/nimaaskarian/goje/utils"
)

func ntfySetup(config *AppConfig) {
	slog.Info("setting up ntfy", "address", config.NtfyAddress)
	config.NtfyAddress = utils.FixHttpAddress(config.NtfyAddress)
	config.NtfyClickUrl = utils.FixHttpAddress(config.NtfyClickUrl)
	config.Timer.OnInit.Append((func(pt *timer.PomodoroTimer) {
		if req, err := ntfyRequest(config, "Timer init!", "tomato,arrow_forward"); err == nil {
			if _, err := http.DefaultClient.Do(req); err != nil {
				slog.Error("Failed to send ntfy request", "err", err)
			}
		}
	}))
	config.Timer.OnModeStart.Append((func(pt *timer.PomodoroTimer) {
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
		if req, err := ntfyRequest(config, msg, tags); err == nil {
			if _, err := http.DefaultClient.Do(req); err != nil {
				slog.Error("Failed to send ntfy request", "err", err)
			}
		}
	}))
	config.Timer.OnPause.Append(func(pt *timer.PomodoroTimer) {
		var msg, tags string
		if pt.State.Paused {
			msg = "Timer paused!"
			tags = "pause_button"
		} else {
			msg = "Timer unpaused!"
			tags = "arrow_forward"
		}
		if req, err := ntfyRequest(config, msg, tags); err == nil {
			if _, err := http.DefaultClient.Do(req); err != nil {
				slog.Error("Failed to send ntfy request", "err", err)
			}
		}
	})
	if config.Timer.Paused {
		config.Timer.OnModeEnd.Append((func(pt *timer.PomodoroTimer) {
			if pt.State.Mode == 2 {
				if req, err := ntfyRequest(config, "Long break ended!", "tomato"); err == nil {
					if _, err := http.DefaultClient.Do(req); err != nil {
						slog.Error("Failed to send ntfy request", "err", err)
					}
				}
			}
		}))
	}
}

// helper function that creates a request
func ntfyRequest(config *AppConfig, content, tags string) (*http.Request, error) {
	req, err := http.NewRequest("POST", config.NtfyAddress, strings.NewReader(content))
	if err != nil {
		slog.Error("Failed to create ntfy request", "err", err)
		return nil, err
	}
	if config.NtfyAuth != "" {
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(config.NtfyAuth)))
	}
	if config.NtfyClickUrl != "" {
		req.Header.Set("Click", config.NtfyClickUrl)
	}
	req.Header.Set("Tags", tags)
	return req, nil
}
