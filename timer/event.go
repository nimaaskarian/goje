package timer

type Event struct {
	Name    string
	Payload any
}

func OnModeEndEvent(payload any) (e Event) {
  e.Name = "end"
  e.Payload = payload
  return
}

func OnModeStartEvent(payload any) (e Event) {
  e.Name = "start"
  e.Payload = payload
  return
}

func OnChangeEvent(payload any) (e Event) {
  e.Name = "change"
  e.Payload = payload
  return
}

func OnPauseEvent(payload any) (e Event) {
  e.Name = "pause"
  e.Payload = payload
  return
}


