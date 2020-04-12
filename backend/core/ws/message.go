package ws

type Message interface {
	SetEvent(name string)
	GetEvent() string
	SetBody(body interface{})
	GetBody() interface{}
	SetMeta(k string,value string)
	GetMeta(k string) map[string]string
	reset()
}
