package ws

import (
	"crawlab/core/codec"
	"crawlab/core/codec/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"sync"
	"time"
)

const defaultMultipartMemory = 32 << 20 // 32 MB

type Engine struct {
	RouterGroup
	options  *Options
	Upgrader *websocket.Upgrader
	hub      *hub
	codec    codec.Marshaller
	// Value of 'maxMemory' param that is given to http.Request's ParseMultipartForm
	// method call.
	MaxMultipartMemory int64
	allNoRoute         EventHandlersChain
	allNoMethod        EventHandlersChain
	noRoute            EventHandlersChain
	noMethod           EventHandlersChain
	trees              methodTrees
	secureJsonPrefix   string
	connectHandler     SessionHandlerFunc
	pongHandler        SessionHandlerFunc
	closeHandler       CloseHandlerFunc
	disconnectHandler  SessionHandlerFunc
	errorHandler       ErrorHandlerFunc
	pool               sync.Pool
}

func (e *Engine) messageHandler(ctx *Context) {
	message, err := ctx.Message()
	if err != nil {
		ctx.Abort()
		e.errorHandler(ctx, err)
		return
	}
	eventName := message.GetEvent()
	//// Find root of the tree for the given HTTP method
	t := e.trees
	for i, tl := 0, len(t); i < tl; i++ {
		if t[i].method != ctx.msgType.String() {
			continue
		}
		root := t[i].root
		//	// Find route in tree
		handlers, params := root.getValue(eventName, ctx.Params)
		if handlers != nil {
			ctx.handlers = handlers
			ctx.Params = params
			ctx.Next()
			return
		}

	}
	ctx.handlers = e.allNoRoute
	//serveError(c, http.StatusNotFound, default404Body)
}

func (e *Engine) OnPong(handler SessionHandlerFunc) {
	if e.pongHandler != nil {
		_, _ = fmt.Fprintln(os.Stderr, "pongHandler not nil,force replace it, make sure it not bug")
	}
	e.pongHandler = handler
}
func (e *Engine) OnClose(handler CloseHandlerFunc) {
	if e.closeHandler != nil {
		_, _ = fmt.Fprintln(os.Stderr, "closeHandler not nil,force replace it, make sure it not bug")
	}
	e.closeHandler = handler
}
func (e *Engine) OnConnect(handler SessionHandlerFunc) {
	if e.connectHandler != nil {
		_, _ = fmt.Fprintln(os.Stderr, "connectHandler not nil,force replace it, make sure it not bug")
	}
	e.connectHandler = handler
}

func (e *Engine) OnDisconnect(handler SessionHandlerFunc) {
	if e.connectHandler != nil {
		_, _ = fmt.Fprintln(os.Stderr, "disconnectHandler not nil,force replace it, make sure it not bug")
	}
	e.disconnectHandler = handler
}
func (e *Engine) ServeWebsocket(w http.ResponseWriter, req *http.Request, keys map[string]interface{}) error {
	if e.hub.closed() {
		return errors.New("engine instance is closed")
	}

	conn, err := e.Upgrader.Upgrade(w, req, w.Header())

	if err != nil {
		return err
	}
	pool,err :=ants.NewPool(1000, func(opts *ants.Options) {
		opts.ExpiryDuration = 10 *time.Second
	})
	if err != nil {
		return err
	}
	session := &Session{
		Request: req,
		Keys:    keys,
		conn:    conn,
		output:  make(chan *envelope, e.options.MessageBufferSize),
		engine:  e,
		open:    true,
		rwmutex: &sync.RWMutex{},
		coroutinePool: pool,
	}
	e.hub.register <- session
	if e.connectHandler != nil {
		e.connectHandler(session)
	}
	if session.open {
		go session.writePump()

		session.readPump()
		if !e.hub.closed() {
			e.hub.unregister <- session
		}
		session.close()
	}
	if e.disconnectHandler != nil {
		e.disconnectHandler(session)
	}

	return nil
}

var _ WRouter = &Engine{}

func New(opts... EngineOption) (*Engine,error) {

	options := EngineOptions{
		WsOptions: Options{
			WriteWait:         10 * time.Second,
			PongWait:          60 * time.Second,
			PingPeriod:        (60 * time.Second * 9) / 10,
			MaxMessageSize:    512,
			MessageBufferSize: 256,
		},
		Upgrader:&websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		Marshaller:        &json.Marshaler{},

	}
	for _, o := range opts {
		o(&options)
	}
	hub,err := newHub()
	if err!=nil{
		return nil,err
	}
	go hub.run()

	engine := &Engine{
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "",
			root:     true,
		},
		options:   &options.WsOptions  ,
		Upgrader:           options.Upgrader,
		MaxMultipartMemory: defaultMultipartMemory,
		trees:              make(methodTrees, 0, 9),
		secureJsonPrefix:   "while(1);",
		hub:                hub,
		errorHandler:       defaultErrorHandle,
		codec:              options.Marshaller ,
	}
	engine.RouterGroup.engine = engine
	engine.pool.New = func() interface{} {
		return engine.allocateContext()
	}
	return engine,nil
}
func (e *Engine) allocateContext() *Context {
	return &Context{engine: e}
}

func (e *Engine) addRoute(method Method, path string, handlers EventHandlersChain) {

	root := e.trees.get(string(method))
	if root == nil {
		root = new(node)
		e.trees = append(e.trees, methodTree{method: string(method), root: root})
	}
	root.addRoute(path, handlers)
}

func defaultErrorHandle(ctx *Context, err error) {
	fmt.Println(err)
}
