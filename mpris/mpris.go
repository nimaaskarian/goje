// https://github.com/emersion/go-mpris/blob/master/mpris.go
// the repo is unmaintained and archived so instead of go get direct copying is used
// https://github.com/natsukagami/mpd-mpris/blob/master/player.go

package mpris

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
	"github.com/nimaaskarian/goje/timer"
)

type Instance struct {
	dbus  *dbus.Conn
	pt *timer.PomodoroTimer

	// interface implementations
	root   *MediaPlayer2
	player *Player
	props *prop.Properties

	Name string

	displayName string
}

// NewInstance creates a new instance that takes care of the specified mpd.
func NewInstance(pt *timer.PomodoroTimer) (ins *Instance, err error) {
	ins = &Instance{
		pt: pt,
		Name: fmt.Sprintf("org.mpris.MediaPlayer2.goje.instance%d", os.Getpid()),
		displayName: "Goje",
	}

	if ins.dbus, err = dbus.SessionBus(); err != nil {
		return nil, err
	}

	ins.root = &MediaPlayer2{Instance: ins}
	ins.player = &Player{Instance: ins}
	playStatus := PlaybackStatusFromPomodoroTimer(pt)
	loopStatus := LoopStatusFromPomodoroTimer(pt)
	ins.player.props = map[string]*prop.Prop{
		"PlaybackStatus": newProp(playStatus, nil),
		"LoopStatus":     newProp(loopStatus, ins.player.OnLoopStatus),
		"Rate":           newProp(1.0, notImplemented),
		"Shuffle":        newProp(false, notImplemented),
		"Metadata":       newProp("Goje", nil),
		"Volume":         newProp(1.0, notImplemented),
		"Position": {
			Value:    UsFromDuration(pt.State.Duration),
			Writable: true,
			Emit:     prop.EmitFalse,
			Callback: nil,
		},
		"MinimumRate":   newProp(1.0, nil),
		"MaximumRate":   newProp(1.0, nil),
		"CanGoNext":     newProp(true, nil),
		"CanGoPrevious": newProp(true, nil),
		"CanPlay":       newProp(true, nil),
		"CanPause":      newProp(true, nil),
		"CanSeek":       newProp(true, nil),
		"CanControl":    newProp(true, nil),
	}

	ins.props, err = prop.Export(ins.dbus, "/org/mpris/MediaPlayer2", map[string]map[string]*prop.Prop{
		"org.mpris.MediaPlayer2":        ins.root.properties(),
		"org.mpris.MediaPlayer2.Player": ins.player.props,
	})
	return
}

func notImplemented(c *prop.Change) *dbus.Error {
	return dbus.MakeFailedError(errors.New("Not implemented"))
}

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

// Start starts the instance. Blocking, so you should fire and forget ;)
func (ins *Instance) Start(ctx context.Context) error {
	ins.dbus.Export(ins.root, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2")
	ins.dbus.Export(ins.player, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.Player")
	ins.dbus.Export(introspect.NewIntrospectable(ins.IntrospectNode()), "/org/mpris/MediaPlayer2", "org.freedesktop.DBus.Introspectable")

	ins.pt.Config.OnChange.Append(func(pt *timer.PomodoroTimer) {
		go ins.player.setProp("org.mpris.MediaPlayer2.Player", "Position", dbus.MakeVariant(UsFromDuration(pt.State.Duration)))
	})
	reply, err := ins.dbus.RequestName(ins.Name, dbus.NameFlagReplaceExisting)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		return err
	}
	return nil
}


func (p *Player) setProp(iface, name string, value dbus.Variant) {
	if err := p.Instance.props.Set(iface, name, value); err != nil {
		log.Printf("Setting %s %s failed: %+v\n", iface, name, err)
	}
}

func (i *Instance) IntrospectNode() *introspect.Node {
	return &introspect.Node{
		Name: i.Name,
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			introspect.Interface{
				Name: "org.mpris.MediaPlayer2",
				Properties: []introspect.Property{
					introspect.Property{
						Name:   "CanQuit",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "CanRaise",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "HasTrackList",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "Identity",
						Type:   "s",
						Access: "read",
					},
					introspect.Property{
						Name:   "SupportedUriSchemes",
						Type:   "as",
						Access: "read",
					},
					introspect.Property{
						Name:   "SupportedMimeTypes",
						Type:   "as",
						Access: "read",
					},
					{
						Name:   "Fullscreen",
						Type:   "b",
						Access: "read",
					},
					{
						Name:   "CanSetFullscreen",
						Type:   "b",
						Access: "read",
					},
					{
						Name:   "DesktopEntry",
						Type:   "s",
						Access: "read",
					},
				},
				Methods: []introspect.Method{
					introspect.Method{
						Name: "Raise",
					},
					introspect.Method{
						Name: "Quit",
					},
				},
			},
			introspect.Interface{
				Name: "org.mpris.MediaPlayer2.Player",
				Properties: []introspect.Property{
					introspect.Property{
						Name:   "PlaybackStatus",
						Type:   "s",
						Access: "read",
					},
					introspect.Property{
						Name:   "LoopStatus",
						Type:   "s",
						Access: "readwrite",
					},
					introspect.Property{
						Name:   "Rate",
						Type:   "d",
						Access: "readwrite",
					},
					introspect.Property{
						Name:   "Shuffle",
						Type:   "b",
						Access: "readwrite",
					},
					introspect.Property{
						Name:   "Metadata",
						Type:   "a{sv}",
						Access: "read",
					},
					introspect.Property{
						Name:   "Volume",
						Type:   "d",
						Access: "readwrite",
					},
					introspect.Property{
						Name:   "Position",
						Type:   "x",
						Access: "read",
					},
					introspect.Property{
						Name:   "MinimumRate",
						Type:   "d",
						Access: "read",
					},
					introspect.Property{
						Name:   "MaximumRate",
						Type:   "d",
						Access: "read",
					},
					introspect.Property{
						Name:   "CanGoNext",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "CanGoPrevious",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "CanPlay",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "CanSeek",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "CanControl",
						Type:   "b",
						Access: "read",
					},
					{
						Name:   "CanPause",
						Type:   "b",
						Access: "read",
					},
				},
				Signals: []introspect.Signal{
					introspect.Signal{
						Name: "Seeked",
						Args: []introspect.Arg{
							introspect.Arg{
								Name: "Position",
								Type: "x",
							},
						},
					},
				},
				Methods: []introspect.Method{
					introspect.Method{
						Name: "Next",
					},
					introspect.Method{
						Name: "Previous",
					},
					introspect.Method{
						Name: "Pause",
					},
					introspect.Method{
						Name: "PlayPause",
					},
					introspect.Method{
						Name: "Stop",
					},
					introspect.Method{
						Name: "Play",
					},
					introspect.Method{
						Name: "Seek",
						Args: []introspect.Arg{
							introspect.Arg{
								Name:      "Offset",
								Type:      "x",
								Direction: "in",
							},
						},
					},
					introspect.Method{
						Name: "SetPosition",
						Args: []introspect.Arg{
							introspect.Arg{
								Name:      "TrackId",
								Type:      "o",
								Direction: "in",
							},
							introspect.Arg{
								Name:      "Position",
								Type:      "x",
								Direction: "in",
							},
						},
					},
				},
			},
			introspect.Interface{
				Name: "org.freedesktop.DBus.Properties",
				Signals: []introspect.Signal{
					introspect.Signal{
						Name: "PropertiesChanged",
						Args: []introspect.Arg{
							introspect.Arg{
								Name: "interface_name",
								Type: "s",
							},
							introspect.Arg{
								Name: "changed_properties",
								Type: "a{sv}",
							},
						},
					},
				},
				Methods: []introspect.Method{
					introspect.Method{
						Name: "Get",
						Args: []introspect.Arg{
							introspect.Arg{
								Name:      "interface_name",
								Type:      "s",
								Direction: "in",
							},
							introspect.Arg{
								Name:      "property_name",
								Type:      "s",
								Direction: "in",
							},
							introspect.Arg{
								Name:      "value",
								Type:      "v",
								Direction: "out",
							},
						},
					},
					introspect.Method{
						Name: "GetAll",
						Args: []introspect.Arg{
							introspect.Arg{
								Name:      "interface_name",
								Type:      "s",
								Direction: "in",
							},
							introspect.Arg{
								Name:      "properties",
								Type:      "a{sv}",
								Direction: "out",
							},
						},
					},
					introspect.Method{
						Name: "Set",
						Args: []introspect.Arg{
							introspect.Arg{
								Name:      "interface_name",
								Type:      "s",
								Direction: "in",
							},
							introspect.Arg{
								Name:      "property_name",
								Type:      "s",
								Direction: "out",
							},
							introspect.Arg{
								Name:      "value",
								Type:      "v",
								Direction: "in",
							},
						},
					},
				},
			},
			// TODO: This interface is not fully implemented.
			// introspect.Interface{
			// 	Name: "org.mpris.MediaPlayer2.TrackList",

			// },
		},
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
