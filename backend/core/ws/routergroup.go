package ws
//go:generate enumer -type Method -linecomment
const (
	Binary Method = iota
	Text
	Empty
)
type Method int


type WRouter interface {
	WRoutes
	Group(string, ...EventHandlerFunc) *RouterGroup
	OnConnect(SessionHandlerFunc)
	OnDisconnect(SessionHandlerFunc)
	OnPong(SessionHandlerFunc)
	OnClose(CloseHandlerFunc)
}
type WRoutes interface {
	Use(...EventHandlerFunc) WRoutes

	Handle(Method, string, ...EventHandlerFunc) WRoutes
	Any(string, ...EventHandlerFunc) WRoutes
	Binary(string, ...EventHandlerFunc) WRoutes
	Text(string, ...EventHandlerFunc) WRoutes
	//HEAD(string, ...EventHandlerFunc) WRoutes

	//StaticFile(string, string) WRoutes
	//Static(string, string) WRoutes
	//StaticFS(string, http.FileSystem) WRoutes
}
type RouterGroup struct {
	Handlers EventHandlersChain
	basePath string
	engine   *Engine
	root     bool
}

func (group RouterGroup) Use(middleware ...EventHandlerFunc) WRoutes {
	group.Handlers = append(group.Handlers, middleware...)
	return group.returnObj()
}
func (group *RouterGroup) Group(relativePath string, handlers ...EventHandlerFunc) *RouterGroup {
	return &RouterGroup{
		Handlers: group.combineHandlers(handlers),
		basePath: group.calculateAbsolutePath(relativePath),
		engine:   group.engine,
	}
}
func (group RouterGroup) Handle(method Method, name string, handlers ...EventHandlerFunc) WRoutes {
	return group.handle(method, name, handlers)
}

func (group RouterGroup) Any(event string, handlers ...EventHandlerFunc) WRoutes {
	group.Binary(event, handlers...)
	group.Text(event, handlers...)
	return group.returnObj()

}

func (group RouterGroup) Binary(event string, handlers ...EventHandlerFunc) WRoutes {
	return group.handle(Binary, event, handlers)
}

func (group RouterGroup) Text(event string, handlers ...EventHandlerFunc) WRoutes {
	return group.handle(Text, event, handlers)
}

var _ WRoutes = &RouterGroup{}

func (group *RouterGroup) returnObj() WRoutes {
	if group.root {
		return group.engine
	}
	return group
}
func (group *RouterGroup) combineHandlers(handlers EventHandlersChain) EventHandlersChain {
	finalSize := len(group.Handlers) + len(handlers)
	if finalSize >= int(abortIndex) {
		panic("too many handlers")
	}
	mergedHandlers := make(EventHandlersChain, finalSize)
	copy(mergedHandlers, group.Handlers)
	copy(mergedHandlers[len(group.Handlers):], handlers)
	return mergedHandlers
}
func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(group.basePath, relativePath)
}
func (group *RouterGroup) handle(method Method, eventName string, handlers EventHandlersChain) WRoutes {
	absolutePath := group.calculateAbsolutePath(eventName)
	handlers = group.combineHandlers(handlers)
	group.engine.addRoute(method, absolutePath, handlers)
	return group.returnObj()
}
