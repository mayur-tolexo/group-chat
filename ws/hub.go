package ws

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/mayur-tolexo/flaw"
	"github.com/mayur-tolexo/group-chat/ally"
	"github.com/tidwall/gjson"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

//NewHub will create new hub
func NewHub(event Event) *Hub {
	return &Hub{
		event:      event,
		nextID:     1,
		clients:    make(map[string][]*Client),
		master:     make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

//Run will run hub
func (hub *Hub) Run() {
	for {
		select {
		case client := <-hub.register:
			hub.onConnect(client)
		case client := <-hub.unregister:
			hub.onDisconnect(client)
		}
	}
}

//broadcast will braodcast
func (hub *Hub) broadcast(group string, message interface{}, ignore *Client) (
	err error) {
	for _, c := range hub.clients[group] {
		if c != ignore {
			if err = hub.send(message, c); err != nil {
				break
			}
		}
	}
	return
}

//send will data
func (hub *Hub) send(message interface{}, client *Client) (err error) {
	var data map[string]interface{}
	data, err = ally.StructTagValue(message, "json")
	if hub.master[client.group] == client {
		data["permission"] = Edit
	} else {
		data["permission"] = View
	}
	val, _ := json.Marshal(data)
	client.outbound <- val
	return
}

//HandleWebSocket will handle web socket request
func (hub *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {

	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "could not upgrade", http.StatusInternalServerError)
		return
	}
	vars := mux.Vars(r)
	group := hub.event.GetGroup(vars)

	client := newClient(hub, group, socket)
	client.user.Vars = vars
	go client.write()

	var status int
	if status, err = hub.event.SetUser(client.user); err != nil {
		hub.send(NewError(status, err), client)
	} else {
		hub.register <- client
		go client.read()
	}
}

//onConnect will connect
func (hub *Hub) onConnect(client *Client) {
	log.Println("client connected: ", client.group, client.socket.RemoteAddr())

	// Make new client
	client.id = hub.nextID
	hub.nextID++
	client.user.ID = client.id
	clients := hub.clients[client.group]

	color := generateColor()
	for _, c := range clients {
		if c.user.UserID == client.user.UserID {
			color = c.user.Color
			break
		}
	}
	client.user.Color = color
	clients = append(clients, client)
	hub.clients[client.group] = clients
	if hub.master[client.group] == nil {
		hub.master[client.group] = client
	}

	users := hub.getUsers(client.group, true)
	if data, status, err := hub.event.GetInitData(client); err == nil {
		// Notify that a user joined
		hub.send(NewConnected(client, users, data), client)
		hub.broadcast(client.group, NewUserJoined(client), client)
	} else {
		hub.send(NewError(status, err), client)
	}
}

//onDisconnect msg disconnect
func (hub *Hub) onDisconnect(client *Client) {
	log.Println("client disconnected: ", client.group, client.socket.RemoteAddr())

	client.close()

	// Find index of client
	i := -1
	group := client.group
	for j, c := range hub.clients[group] {
		if c.id == client.id {
			i = j
			break
		}
	}
	// Delete client from list
	copy(hub.clients[group][i:], hub.clients[group][i+1:])
	hub.clients[group][len(hub.clients[group])-1] = nil
	hub.clients[group] = hub.clients[group][:len(hub.clients[group])-1]

	if hub.master[group] == client {
		if len(hub.clients[group]) > 0 {
			hub.master[group] = hub.clients[group][0]
		} else {
			hub.master[group] = nil
		}
	}

	hub.broadcast(group, NewUserLeft(client), nil)
}

//onMessage msg receive
func (hub *Hub) onMessage(data []byte, client *Client) {

	fn := Mods(func(data []byte, client *Client) {
		kind := gjson.GetBytes(data, "kind").String()
		switch kind {
		case KindMaster:
			hub.master[client.group] = client
			hub.sendUserMsg(data, client, true)
		case KindUser:
			hub.sendUserMsg(data, client, true)
		case KindConnected:
			hub.sendUserMsg(data, client, false)
		default:
			hub.sendStrokMsg(data, client)
		}
	})

	for _, m := range hub.mods {
		fn = m(fn)
	}
	fn(data, client)
}

func (hub *Hub) sendUserMsg(data []byte, client *Client, unq bool) {
	var msg SwitchMaster
	msg.ID = client.id
	msg.Kind = KindUser
	msg.Group = client.group
	msg.Users = hub.getUsers(client.group, unq)
	hub.broadcast(client.group, msg, nil)
}

func (hub *Hub) sendStrokMsg(data []byte, client *Client) {
	var (
		msg    *Stroke
		ackMsg interface{}
		status int
	)
	if err := json.Unmarshal(data, &msg); err != nil {
		hub.sendBadReq(client, err)
	} else {
		msg.ID = client.id
		msg.Kind = KindStroke
		msg.Group = client.group
		msg.Status = http.StatusOK
		if hub.master[client.group] == client {
			if msg.Data, ackMsg, status, err = hub.event.SetData(client, msg.Data); err != nil {
				hub.send(NewError(status, err), client)
			}
		} else {
			msg.Data = nil
		}

		if err == nil {
			if err = hub.broadcast(client.group, msg, client); err == nil {
				hub.send(NewAck(msg, ackMsg), client)
			} else {
				hub.sendBadReq(client, err)
			}
		}
	}
}

//sendBadReq will send bad request
func (hub *Hub) sendBadReq(client *Client, err error) {
	err = flaw.BadReqError(err)
	hub.send(NewError(http.StatusBadRequest, err), client)
}

//generateColor rendom color
func generateColor() string {
	c := colorful.Hsv(rand.Float64()*360.0, 0.8, 0.8)
	return c.Hex()
}

func (hub *Hub) getUsers(group string, unq bool) (users []User) {
	users = make([]User, 0)
	visited := make(map[int]struct{})
	for _, c := range hub.clients[group] {
		if _, exists := visited[c.user.UserID]; !exists || !unq {
			visited[c.user.UserID] = struct{}{}
			users = append(users, *c.user)
		}
	}
	return
}

//AddMod middlewares
func (hub *Hub) AddMod(m ...func(Mods) Mods) {
	hub.mods = append(hub.mods, m...)
	return
}
