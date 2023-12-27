package server

import (
	"fmt"
	"log"
)

var NextID int

type Hub struct {
	clients    map[*Client]int
	quit       chan struct{}
	Broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		register:   make(chan *Client),
		quit:       make(chan struct{}),
		unregister: make(chan *Client),
		clients:    make(map[*Client]int),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case <-h.quit:
			log.Println("hub: quit")
			return
		case client := <-h.register:
			NextID++
			fmt.Println("현재ID: ", NextID)
			h.clients[client] = NextID
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				log.Printf("client %d disconnected.\n", h.clients[client])
				delete(h.clients, client)
				close(client.Send)
			}
		case message := <-h.Broadcast:
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}
	}
}
