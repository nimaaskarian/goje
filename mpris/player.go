// thanks to https://github.com/natsukagami/mpd-mpris/
// code of the module "mpris" is mostly MIT licensed. places changed by me are 
// BSD 2-clause as other parts of this code

package mpris

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
	"github.com/nimaaskarian/goje/timer"
)

type MediaPlayer2 struct {
	*Instance
}

func (m *MediaPlayer2) properties() map[string]*prop.Prop {

	return map[string]*prop.Prop{
		"CanQuit":      newProp(false, nil),         // https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Property:CanQuit
		"CanRaise":     newProp(false, nil),         // https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Property:CanRaise
		"HasTrackList": newProp(false, nil),         // https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Property:HasTrackList
		"Identity":     newProp(m.displayName, nil), // https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Property:Identity
		"DesktopEntry": newProp("goje", nil),        // doesn't actually exist

		"Fullscreen":       newProp(false, nil),
		"CanSetFullscreen": newProp(false, nil),

		// Empty because we can't add arbitary files in...
		"SupportedUriSchemes": newProp([]string{}, nil), // https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Property:SupportedUriSchemes
		"SupportedMimeTypes":  newProp([]string{}, nil), // https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Property:SupportedMimeTypes
	}
}

func newProp(value interface{}, cb func(*prop.Change) *dbus.Error) *prop.Prop {
	return &prop.Prop{
		Value:    value,
		Writable: true,
		Emit:     prop.EmitTrue,
		Callback: cb,
	}
}


func (p *Player) setProp(iface, name string, value dbus.Variant) {
	if err := p.Instance.props.Set(iface, name, value); err != nil {
		log.Printf("Setting %s %s failed: %+v\n", iface, name, err)
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
		return dbus.MakeFailedError(errors.New("not implemented"))
	}
	return nil
}

// Next skips to the next track in the tracklist.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Next
func (p *Player) Next() *dbus.Error {
	p.pt.SwitchNextMode()
	return nil
}

// Previous skips to the previous track in the tracklist.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Previous
func (p *Player) Previous() *dbus.Error {
	p.pt.SwitchPrevMode()
	return nil
}

// Pause pauses playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Pause
func (p *Player) Pause() *dbus.Error {
	p.pt.Pause(true)
	return nil
}

// Pause starts or resumes playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Play
func (p *Player) Play() *dbus.Error {
	p.pt.Pause(false)
	return nil
}

// PlayPause toggles playback.
// If playback is already paused, resumes playback.
// If playback is stopped, starts playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:PlayPause
func (p *Player) PlayPause() *dbus.Error {
	p.pt.TogglePause()
	return nil
}

// Stop stops playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Stop
func (p *Player) Stop() *dbus.Error {
	p.pt.Init()
	p.pt.Pause(true)
	return nil
}

// Seek seeks forward in the current track by the specified number of microseconds.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Seek
func (p *Player) Seek(x TimeInUs) *dbus.Error {
	p.pt.SeekAdd(x.Duration())
	return nil
}

type TrackID string
// SetPosition sets the current track position in microseconds.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:SetPosition
func (p *Player) SetPosition(o TrackID, x TimeInUs) *dbus.Error {
	p.pt.SeekTo(x.Duration())
	return nil
}
