package main

import (
	"github.com/gin-gonic/gin"
	"github.com/insoft-cloud/websocket-go/kafka"
	"github.com/insoft-cloud/websocket-go/server"
)

func main() {
	hub := server.NewHub()
	go kafka.InitKafkaConsumer(hub)
	go hub.Run()

	r := gin.Default()
	r.GET("/ws", server.WsHandler(hub))
	r.Run(":8080")
}
