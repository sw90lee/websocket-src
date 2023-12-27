package server

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WsHandler(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)
		}
		client := &Client{Hub: hub, conn: conn, Send: make(chan []byte)}
		client.Hub.register <- client

		go client.WritePump()
		go client.ReadPump()
	}
}
