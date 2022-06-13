package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"time"
)

func main() {
	r := gin.Default()
	r.GET("/ws", HandleWebsocket)
	_ = r.Run(":9000")
}

func HandleWebsocket(c *gin.Context) {
	u := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	ws, err := u.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("cannot upgrade: %v", err)
	}
	defer func(ws *websocket.Conn) {
		_ = ws.Close()
	}(ws)

	ch := make(chan struct{})
	go func(ch chan struct{}) {
		for {
			m := make(map[string]interface{})
			if err = ws.ReadJSON(&m); err != nil {
				if !websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					fmt.Printf("unexpected read error: %v\n", err)
				} else {
					fmt.Printf("normal error: %v\n", err)
				}
				ch <- struct{}{}
				break
			}
			fmt.Println(m)
		}
	}(ch)

	i := 0
	for {
		select {
		case <-time.After(time.Second):
		case <-ch:
			return
		}

		i++
		err := ws.WriteJSON(map[string]string{
			"hello": "websocket",
			"id":    strconv.Itoa(i),
		})
		if err != nil {
			fmt.Printf("err: %v", err)
		}

	}
}
