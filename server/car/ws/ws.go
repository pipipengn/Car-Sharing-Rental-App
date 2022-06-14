package ws

import (
	"context"
	mq "coolcar/car/rabbitmq/mq_interface"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net/http"
)

type Options struct {
	Logger        *zap.Logger
	Upgrader      *websocket.Upgrader
	CarSubscriber mq.CarSubscriber
}

func NewHandler(o *Options) gin.HandlerFunc {
	return func(c *gin.Context) {
		ws, err := o.Upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			o.Logger.Warn("cannot upgrade", zap.Error(err))
			return
		}
		defer func(ws *websocket.Conn) {
			_ = ws.Close()
		}(ws)

		msgs, cleanFunc, err := o.CarSubscriber.Subscribe(context.Background())
		defer cleanFunc()
		if err != nil {
			o.Logger.Error("cannot subscribe", zap.Error(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		done := make(chan struct{})
		go func(done chan struct{}) {
			for {
				if _, _, err = ws.ReadMessage(); err != nil {
					if !websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
						o.Logger.Warn("unexpected read error: %v\n", zap.Error(err))
					} else {
						o.Logger.Warn("normal error: %v\n", zap.Error(err))
					}
					done <- struct{}{}
					break
				}
			}
		}(done)

		for {
			select {
			case msg := <-msgs:
				err := ws.WriteJSON(msg)
				if err != nil {
					o.Logger.Warn("cannot write json", zap.Error(err))
				}
			case <-done:
				return
			}
		}
	}
}
