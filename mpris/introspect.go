package mpris

import (
	"github.com/godbus/dbus/v5/introspect"
)

func (i *Instance) IntrospectNode() *introspect.Node {
	return &introspect.Node{
		Name: i.name,
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			{
				Name: "org.mpris.MediaPlayer2",
				Properties: []introspect.Property{
					{
						Name:   "CanQuit",
						Type:   "b",
						Access: "read",
					},
					{
						Name:   "CanRaise",
						Type:   "b",
						Access: "read",
					},
					{
						Name:   "HasTrackList",
						Type:   "b",
						Access: "read",
					},
					{
						Name:   "Identity",
						Type:   "s",
						Access: "read",
					},
					{
						Name:   "SupportedUriSchemes",
						Type:   "as",
						Access: "read",
					},
					{
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
					{
						Name: "Raise",
					},
					{
						Name: "Quit",
					},
				},
			},
			{
				Name: "org.mpris.MediaPlayer2.Player",
				Properties: []introspect.Property{
					{
						Name:   "PlaybackStatus",
						Type:   "s",
						Access: "read",
					},
					{
						Name:   "LoopStatus",
						Type:   "s",
						Access: "readwrite",
					},
					{
						Name:   "Rate",
						Type:   "d",
						Access: "readwrite",
					},
					{
						Name:   "Shuffle",
						Type:   "b",
						Access: "readwrite",
					},
					{
						Name:   "Metadata",
						Type:   "a{sv}",
						Access: "read",
					},
					{
						Name:   "Volume",
						Type:   "d",
						Access: "readwrite",
					},
					{
						Name:   "Position",
						Type:   "x",
						Access: "read",
					},
					{
						Name:   "MinimumRate",
						Type:   "d",
						Access: "read",
					},
					{
						Name:   "MaximumRate",
						Type:   "d",
						Access: "read",
					},
					{
						Name:   "CanGoNext",
						Type:   "b",
						Access: "read",
					},
					{
						Name:   "CanGoPrevious",
						Type:   "b",
						Access: "read",
					},
					{
						Name:   "CanPlay",
						Type:   "b",
						Access: "read",
					},
					{
						Name:   "CanSeek",
						Type:   "b",
						Access: "read",
					},
					{
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
					{
						Name: "Seeked",
						Args: []introspect.Arg{
							{
								Name: "Position",
								Type: "x",
							},
						},
					},
				},
				Methods: []introspect.Method{
					{
						Name: "Next",
					},
					{
						Name: "Previous",
					},
					{
						Name: "Pause",
					},
					{
						Name: "PlayPause",
					},
					{
						Name: "Stop",
					},
					{
						Name: "Play",
					},
					{
						Name: "Seek",
						Args: []introspect.Arg{
							{
								Name:      "Offset",
								Type:      "x",
								Direction: "in",
							},
						},
					},
					{
						Name: "SetPosition",
						Args: []introspect.Arg{
							{
								Name:      "TrackId",
								Type:      "o",
								Direction: "in",
							},
							{
								Name:      "Position",
								Type:      "x",
								Direction: "in",
							},
						},
					},
				},
			},
			{
				Name: "org.freedesktop.DBus.Properties",
				Signals: []introspect.Signal{
					{
						Name: "PropertiesChanged",
						Args: []introspect.Arg{
							{
								Name: "interface_name",
								Type: "s",
							},
							{
								Name: "changed_properties",
								Type: "a{sv}",
							},
						},
					},
				},
				Methods: []introspect.Method{
					{
						Name: "Get",
						Args: []introspect.Arg{
							{
								Name:      "interface_name",
								Type:      "s",
								Direction: "in",
							},
							{
								Name:      "property_name",
								Type:      "s",
								Direction: "in",
							},
							{
								Name:      "value",
								Type:      "v",
								Direction: "out",
							},
						},
					},
					{
						Name: "GetAll",
						Args: []introspect.Arg{
							{
								Name:      "interface_name",
								Type:      "s",
								Direction: "in",
							},
							{
								Name:      "properties",
								Type:      "a{sv}",
								Direction: "out",
							},
						},
					},
					{
						Name: "Set",
						Args: []introspect.Arg{
							{
								Name:      "interface_name",
								Type:      "s",
								Direction: "in",
							},
							{
								Name:      "property_name",
								Type:      "s",
								Direction: "out",
							},
							{
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
