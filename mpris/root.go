package mpris

import (
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
	"github.com/nimaaskarian/goje/utils"
)

type MediaPlayer2 struct {
	*Instance;
	webguiAddress string;
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


// Raise brings the media player's user interface to the front using any appropriate mechanism available.
// For goje this can mean opening goje's url

// https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Method:Raise
func (m *MediaPlayer2) Raise() *dbus.Error { 
	if m.webguiAddress != "" {
		utils.OpenURL(utils.FixHttpAddress(m.webguiAddress))
	}
	return nil
}

// Quit causes the media player to stop running.
// But for goje, it's not up to the client to end its existence. Hence this function does nothing.
//
// https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Method:Quit
func (m *MediaPlayer2) Quit() *dbus.Error { return nil }
