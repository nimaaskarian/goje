package timer

type Event struct {
	name    string
	payload any
}

func (e Event) Payload() any {
  return e.payload
}

func (e Event) Name() string {
  return e.name
}

func OnModeEndEvent(payload any) (e Event) {
  e.name = "end"
  e.payload = payload
  return
}

func OnModeStartEvent(payload any) (e Event) {
  e.name = "start"
  e.payload = payload
  return
}

func OnChangeEvent(payload any) (e Event) {
  e.name = "change"
  e.payload = payload
  return
}

func OnPauseEvent(payload any) (e Event) {
  e.name = "pause"
  e.payload = payload
  return
}


