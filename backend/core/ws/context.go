package ws

import (
	"github.com/gin-gonic/gin/binding"
)

type Context struct {
	Params     Params
	handlers   EventHandlersChain
	index      int8
	session    *Session
	engine     *Engine
	data       []byte
	messageBag Message
	msgType    Method
	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[string]interface{}

	// Errors is a list of errors attached to all the handlers/middlewares who used this context.
	//Errors errorMsgs
}

func (c *Context) reset() {
	c.Params = c.Params[0:0]
	c.handlers = nil
	c.session = nil
	c.index = -1
	c.msgType = Empty
	c.data = c.data[0:0]
	c.Keys = nil
	c.messageBag.reset()
}
func (c *Context) Copy() *Context {
	var cp = *c
	cp.index = abortIndex
	cp.handlers = nil
	cp.Keys = map[string]interface{}{}
	for k, v := range c.Keys {
		cp.Keys[k] = v
	}
	return &cp
}

// Handler returns the main handler.
func (c *Context) Handler() EventHandlerFunc {
	return c.handlers.Last()
}

func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}
func (c *Context) Write(message interface{}) error {
	m, err := c.engine.codec.Marshal(message)
	if err != nil {
		return err
	}
	return c.session.Write(m)
}
func (c *Context) WriteBinary(message []byte) error {
	m, err := c.engine.codec.Marshal(message)
	if err != nil {
		return err
	}
	return c.session.WriteBinary(m)
}

// IsAborted returns true if the current context was aborted.
func (c *Context) IsAborted() bool {
	return c.index >= abortIndex
}
func (c *Context) Abort() {
	c.index = abortIndex
}
func (c *Context) Message() (msg Message, err error) {
	if c.messageBag.GetEvent() == "" {
		err = c.engine.codec.Unmarshal(c.data, &c.messageBag)
	}
	return c.messageBag, err
}

// ShouldBindJSON is a shortcut for c.ShouldBindWith(obj, binding.JSON).
//func (c *Context) ShouldBindJSON(obj interface{}) error {
//	return c.ShouldBindWith(obj, binding.JSON)
//}

// ShouldBindXML is a shortcut for c.ShouldBindWith(obj, binding.XML).
//func (c *Context) ShouldBindXML(obj interface{}) error {
//	return c.ShouldBindWith(obj, binding.XML)
//}

//// ShouldBindQuery is a shortcut for c.ShouldBindWith(obj, binding.Query).
//func (c *Context) ShouldBindQuery(obj interface{}) error {
//	return c.ShouldBindWith(obj, binding.Query)
//}

// ShouldBindYAML is a shortcut for c.ShouldBindWith(obj, binding.YAML).
//func (c *Context) ShouldBindYAML(obj interface{}) error {
//	return c.ShouldBindWith(obj, binding.YAML)
//}

// ShouldBindUri binds the passed struct pointer using the specified binding engine.
func (c *Context) ShouldBindUri(obj interface{}) error {
	m := make(map[string][]string)
	for _, v := range c.Params {
		m[v.Key] = []string{v.Value}
	}
	return binding.Uri.BindUri(m, obj)
}

// ShouldBindWith binds the passed struct pointer using the specified binding engine.
// See the binding package.
//func (c *Context) ShouldBindWith(obj interface{}, b binding.BindingBody) error {
//	return b.BindBody(c.messageBag.GetBody(), obj)
//}

// Set is used to store a new key/value pair exclusively for this context.
// It also lazy initializes  c.Keys if it was not used previously.
func (c *Context) Set(key string, value interface{}) {
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (c *Context) Get(key string) (value interface{}, exists bool) {
	value, exists = c.Keys[key]
	return
}
