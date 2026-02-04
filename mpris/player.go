// thanks to https://github.com/natsukagami/mpd-mpris/
// code of the module "mpris" is mostly MIT licensed. places changed by me are
// BSD 2-clause as other parts of this code

package mpris

import (
	"log/slog"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"

	"github.com/nimaaskarian/goje/timer"
)

func newProp(value any, cb func(*prop.Change) *dbus.Error) *prop.Prop {
	return &prop.Prop{
		Value:    value,
		Writable: true,
		Emit:     prop.EmitTrue,
		Callback: cb,
	}
}

func (p *Player) OnShuffle(c *prop.Change) *dbus.Error {
	return nil
}

func (p *Player) OnVolume(c *prop.Change) *dbus.Error {
	return nil
}

func (p *Player) setProp(iface, name string, value dbus.Variant) {
	if err := p.Instance.props.Set(iface, name, value); err != nil {
		slog.Error("setting prop failed", "interface", iface, "name", name, "err", err)
	}
}

type Player struct {
	*Instance
	props map[string]*prop.Prop
}

type TimeInUs int64

// UsFromDuration returns the type from a time.Duration
func UsFromDuration(t time.Duration) TimeInUs {
	return TimeInUs(t / time.Microsecond)
}

func (t TimeInUs) Duration() time.Duration {
	return time.Duration(t) * time.Microsecond
}

type LoopStatus = string

// Defined LoopStatuses
const (
	LoopStatusNone     LoopStatus = "None"
	LoopStatusTrack    LoopStatus = "Track"
	LoopStatusPlaylist LoopStatus = "Playlist"
)

type PlaybackStatus string

const (
	PlaybackStatusPlaying PlaybackStatus = "Playing"
	PlaybackStatusPaused  PlaybackStatus = "Paused"
	PlaybackStatusStopped PlaybackStatus = "Stopped"
)

// https://specifications.freedesktop.org/mpris/latest/Player_Interface.html#Playback_Status
func PlaybackStatusFromPomodoroTimer(pt *timer.PomodoroTimer) PlaybackStatus {
	if pt.State.Paused {
		// if intance pauses after long break finish, and timer is in an
		// init-paused state, then the timer is stopped
		if pt.Config.Paused && pt.State.Mode == timer.Pomodoro && pt.State.Duration == pt.Config.Duration[timer.Pomodoro] {
			return PlaybackStatusStopped
		}
		return PlaybackStatusPaused
	}
	return PlaybackStatusPlaying
}

func LoopStatusFromPomodoroTimer(pt *timer.PomodoroTimer) LoopStatus {
	if pt.Config.Paused {
		return LoopStatusNone
	}
	return LoopStatusPlaylist
}

// OnLoopStatus handles LoopStatus change.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Property:LoopStatus
func (p *Player) OnLoopStatus(c *prop.Change) *dbus.Error {
	loop := LoopStatus(c.Value.(string))
	switch loop {
	case LoopStatusNone:
		p.pt.Config.Paused = true
	case LoopStatusPlaylist:
		p.pt.Config.Paused = false
	case LoopStatusTrack:
		p.pt.Config.Paused = false
	}
	return nil
}

// Next skips to the next track in the tracklist.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Next
func (p *Player) Next() *dbus.Error {
	slog.Info("next request recieved from mpris")
	p.pt.SwitchNextMode()
	return nil
}

// Previous skips to the previous track in the tracklist.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Previous
func (p *Player) Previous() *dbus.Error {
	slog.Info("prev request recieved from mpris")
	p.pt.SwitchPrevMode()
	return nil
}

// Pause pauses playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Pause
func (p *Player) Pause() *dbus.Error {
	slog.Info("pause request recieved from mpris")
	p.pt.Pause(true)
	return nil
}

// Pause starts or resumes playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Play
func (p *Player) Play() *dbus.Error {
	slog.Info("play request recieved from mpris")
	p.pt.Pause(false)
	return nil
}

// PlayPause toggles playback.
// If playback is already paused, resumes playback.
// If playback is stopped, starts playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:PlayPause
func (p *Player) PlayPause() *dbus.Error {
	slog.Info("play-pause request recieved from mpris")
	p.pt.TogglePause()
	return nil
}

// Stop stops playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Stop
func (p *Player) Stop() *dbus.Error {
	slog.Info("stop request recieved from mpris")
	p.pt.Init()
	p.pt.Pause(true)
	return nil
}

// Seek seeks forward in the current track by the specified number of microseconds.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Seek
func (p *Player) Seek(x TimeInUs) *dbus.Error {
	slog.Info("seek recieved from mpris", "duration", x.Duration())
	p.pt.SeekAdd(x.Duration())
	return nil
}

type TrackID string

// SetPosition sets the current track position in microseconds.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:SetPosition
func (p *Player) SetPosition(o TrackID, x TimeInUs) *dbus.Error {
	slog.Info("set-position recieved from mpris", "duration", x.Duration())
	p.pt.SeekTo(x.Duration())
	return nil
}

// Emit the Seeked DBus signal.
func (p *Player) Seeked(x TimeInUs) *dbus.Error {
	return dbus.MakeFailedError(p.dbus.Emit("/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.Player.Seeked", x))
}
