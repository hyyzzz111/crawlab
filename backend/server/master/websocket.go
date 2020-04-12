package master

import (
	"crawlab/app/master/config"
	"crawlab/pkg/core/ws"
	"crawlab/server/common"
	"github.com/gin-gonic/gin"
)

type websocketServer struct {
	engine *ws.Engine
}

func (w *websocketServer) installWebsocketService(g *gin.Engine, config *config.Websocket) {
	if config.Enable {
		g.Any(config.Path, func(context *gin.Context) {
			err := w.engine.ServeWebsocket(context.Writer, context.Request, context.Keys)
			if err != nil {
				_ = context.AbortWithError(400, err)
			}
		})
	}
}
func (w *websocketServer) RegisterRaw(event common.WebsocketEvent, method ws.Method, handlers ws.EventHandlersChain) {
	switch method {
	case ws.Text:
		w.engine.Text(event.String(), handlers...)
	case ws.Binary:
		w.engine.Binary(event.String(), handlers...)
	}
}
func (w *websocketServer) Register(event common.WebsocketEvent, handlers ws.EventHandlersChain) {
	w.RegisterRaw(event, ws.Text, handlers)
}
func newWebsocketServer(opts ...ws.EngineOption) (*websocketServer, error) {
	engine, err := ws.New(opts...)
	if err != nil {
		return nil, err
	}
	return &websocketServer{engine: engine}, nil
}
