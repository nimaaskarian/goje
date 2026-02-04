package mpris

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"

	"github.com/nimaaskarian/goje/timer"
)

type Instance struct {
	dbus *dbus.Conn
	pt   *timer.PomodoroTimer

	// interface implementations
	root   *MediaPlayer2
	player *Player
	props  *prop.Properties

	name string

	displayName string
}

type MetadataMap map[string]any

type InstanceOpts struct {
	NoInstance bool;
	WebguiAddress string;
}

func NewInstance(pt *timer.PomodoroTimer, opts *InstanceOpts) (ins *Instance, err error) {
	var name string
	if opts.NoInstance {
		name = "org.mpris.MediaPlayer2.goje"
	} else {
		name = fmt.Sprintf("org.mpris.MediaPlayer2.goje.instance%d", os.Getpid())
	}
	ins = &Instance{
		pt:          pt,
		name:        name,
		displayName: "Goje",
	}

	if ins.dbus, err = dbus.SessionBus(); err != nil {
		return nil, err
	}

	ins.root = &MediaPlayer2{Instance: ins, webguiAddress: opts.WebguiAddress}
	ins.player = &Player{Instance: ins}
	playStatus := PlaybackStatusFromPomodoroTimer(pt)
	loopStatus := LoopStatusFromPomodoroTimer(pt)
	ins.player.props = map[string]*prop.Prop{
		"PlaybackStatus": newProp(playStatus, nil),
		"LoopStatus":     newProp(loopStatus, ins.player.OnLoopStatus),
		"Rate":           newProp(1.0, notImplemented),
		"Shuffle":        newProp(false, ins.player.OnShuffle),
		"Metadata":       newProp(MapFromTimer(pt), nil),
		"Volume":         newProp(1.0, ins.player.OnVolume),
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

// DON'T PUT EMPTY STRING IN THE MAP. results in error
func MapFromTimer(pt *timer.PomodoroTimer) MetadataMap {
	return MetadataMap{
		"mpris:trackid": dbus.ObjectPath(fmt.Sprintf("/org/goje/Mode/%d", pt.State.Mode)),
		"mpris:length":  pt.State.Duration / time.Microsecond,
		"xesam:title":   pt.State.Mode.String(),
	}
}

func notImplemented(c *prop.Change) *dbus.Error {
	return dbus.MakeFailedError(errors.New("Not implemented"))
}

// Start starts the instance. Blocking, so you should fire and forget ;)
func (ins *Instance) Start(ctx context.Context) error {
	ins.dbus.Export(ins.root, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2")
	ins.dbus.Export(ins.player, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.Player")
	ins.dbus.Export(introspect.NewIntrospectable(ins.IntrospectNode()), "/org/mpris/MediaPlayer2", "org.freedesktop.DBus.Introspectable")

	reply, err := ins.dbus.RequestName(ins.name, dbus.NameFlagReplaceExisting)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		return err
	}
	ins.pt.Config.OnChange.Append(func(pt *timer.PomodoroTimer) {
		ins.player.setProp("org.mpris.MediaPlayer2.Player", "Position", dbus.MakeVariant(UsFromDuration(pt.State.Duration)))
		ins.player.setProp("org.mpris.MediaPlayer2.Player", "Metadata", dbus.MakeVariant(MapFromTimer(pt)))
	})
	return nil
}

func (ins *Instance) Close() error {
	return ins.dbus.Close()
}
