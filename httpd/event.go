package httpd

type Event struct {
	Name    string
	Payload any
}

func NewEvent(payload any, name string) (e Event) {
	e.Name = name
	e.Payload = payload
	return
}

func ChangeEvent(payload any) Event {
	return NewEvent(payload, "change")
}
