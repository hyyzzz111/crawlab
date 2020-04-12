package ws

type envelope struct {
	t      int
	msg    []byte
	filter filterFunc
}
