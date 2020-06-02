package ws

import (
	"github.com/gorilla/websocket"
)

func newClient(hub *Hub, group string, socket *websocket.Conn) *Client {
	return &Client{
		hub:      hub,
		group:    group,
		user:     &User{},
		socket:   socket,
		outbound: make(chan []byte),
	}
}

//read will read from client
func (client *Client) read() {
	defer func() {
		client.hub.unregister <- client
	}()
	for {
		_, data, err := client.socket.ReadMessage()
		if err != nil {
			break
		}
		client.hub.onMessage(data, client)
	}
}

//write will write on client socket
func (client *Client) write() {
	for {
		select {
		case data, ok := <-client.outbound:
			if !ok {
				client.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			client.socket.WriteMessage(websocket.TextMessage, data)
		}
	}
}

//close will close client socket
func (client Client) close() {
	client.socket.Close()
	close(client.outbound)
}

//GetUser will return client user
func (client *Client) GetUser() User {
	return *client.user
}

//GetGroup will return client group
func (client *Client) GetGroup() string {
	return client.group
}
