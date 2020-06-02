package ws

import (
	"net/http"

	"github.com/mayur-tolexo/flaw"
)

//NewConnected client connection
func NewConnected(client *Client, users []User, data interface{}) *Connected {
	c := &Connected{Color: client.user.Color, Users: users}
	c.Kind = KindConnected
	c.ID = client.id
	c.Group = client.group
	c.Data = data
	c.Status = http.StatusOK
	return c
}

//NewUserJoined user joined
func NewUserJoined(client *Client) *UserJoined {
	uj := &UserJoined{
		Kind:  KindUserJoined,
		ID:    client.id,
		Group: client.group,
	}
	uj.User = *client.user
	return uj
}

//NewUserLeft user left
func NewUserLeft(client *Client) *UserLeft {
	return &UserLeft{
		Kind:  KindUserLeft,
		ID:    client.id,
		Group: client.group,
	}
}

//NewError will send error
func NewError(status int, err error) *Stroke {
	dMsg, trace := flaw.GetDebug(err)
	return &Stroke{
		Kind:   KindError,
		Status: status,
		Error:  dMsg,
		Trace:  trace,
		Msg:    flaw.GetMsg(err),
	}
}

//NewAck will send msg ack
func NewAck(msg *Stroke, ackData interface{}) *Ack {
	return &Ack{
		Kind:  KindAck,
		ID:    msg.ID,
		Group: msg.Group,
		MsgID: msg.MsgID,
		Data:  ackData,
	}
}
