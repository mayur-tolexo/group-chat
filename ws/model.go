package ws

import (
	"github.com/gorilla/websocket"
)

//constants
const (
	KindConnected  = "connected"
	KindUserJoined = "joined"
	KindUserLeft   = "left"
	KindStroke     = "stroke"
	KindMaster     = "master"
	KindUser       = "users"
	KindAck        = "ack"
	KindError      = "error"
)

//perm constants
const (
	View = "view"
	Edit = "edit"
)

//Event type
type Event interface {
	GetGroup(vars map[string]string) string
	SetUser(u *User) (int, error)
	GetInitData(c *Client) (interface{}, int, error)
	SetData(c *Client, data interface{}) (d, ackd interface{}, status int, err error)
}

//Mods function
type Mods func([]byte, *Client)

//User created
type User struct {
	ID     int               `json:"id"`
	UserID int               `json:"user_id"`
	Color  string            `json:"color,omitempty"`
	Data   interface{}       `json:"data,omitempty"`
	Vars   map[string]string `json:"-"`
}

//Stroke data strokes
type Stroke struct {
	Kind       string      `json:"kind"`
	ID         int         `json:"id,omitempty"`
	MsgID      int         `json:"msg_id,omitempty"`
	Group      string      `json:"group,omitempty"`
	Permission string      `json:"permission"`
	Data       interface{} `json:"data,omitempty"`
	Info       interface{} `json:"info,omitempty"`
	Error      string      `json:"error,omitempty"`
	Msg        string      `json:"msg,omitempty"`
	Trace      string      `json:"trace,omitempty"`
	Status     int         `json:"status"`
}

//SwitchMaster master data
type SwitchMaster struct {
	Kind       string `json:"kind"`
	ID         int    `json:"id"`
	Group      string `json:"group"`
	Permission string `json:"permission"`
	Users      []User `json:"users"`
}

//Ack message
type Ack struct {
	Kind       string      `json:"kind"`
	ID         int         `json:"id"`
	MsgID      int         `json:"msg_id,omitempty"`
	Group      string      `json:"group"`
	Permission string      `json:"permission"`
	Data       interface{} `json:"data,omitempty"`
}

//Connected new user
type Connected struct {
	Color  string `json:"color"`
	Users  []User `json:"users"`
	Stroke `json:",flatten"`
}

//Hub model
type Hub struct {
	event      Event
	clients    map[string][]*Client
	master     map[string]*Client
	nextID     int
	register   chan *Client
	unregister chan *Client
	mods       []func(n Mods) Mods
}

//Client model
type Client struct {
	hub      *Hub
	id       int
	group    string
	socket   *websocket.Conn
	outbound chan []byte
	user     *User
}

//UserJoined user joined model
type UserJoined struct {
	Kind  string `json:"kind"`
	Group string `json:"group"`
	ID    int    `json:"id"`
	User
}

//UserLeft user left model
type UserLeft struct {
	Kind  string `json:"kind"`
	Group string `json:"group"`
	ID    int    `json:"id"`
}
