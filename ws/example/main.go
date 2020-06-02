package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mayur-tolexo/group-chat/ws"
)

//User model
type User struct {
}

//GetGroup group name i.e. item_id
func (User) GetGroup(vars map[string]string) string {
	fmt.Println("vars", vars)
	return vars["group_id"]
}

//SetUser user details
func (User) SetUser(u *ws.User) (status int, err error) {
	status = http.StatusOK
	return
}

//GetInitData group data initial data
func (User) GetInitData(c *ws.Client) (data interface{}, status int, err error) {
	data = map[string]interface{}{"hello": "world"}
	status = http.StatusOK
	return
}

//SetData group data
func (User) SetData(c *ws.Client, data interface{}) (d, ackd interface{}, status int, err error) {
	d = fmt.Sprintf("%v", data)
	status = http.StatusOK
	fmt.Println(c.GetGroup(), c.GetUser())
	return
}

func main() {
	hub := ws.NewHub(User{})
	go hub.Run()

	r := mux.NewRouter()
	r.HandleFunc("/group/{group_id}", hub.HandleWebSocket)
	http.Handle("/", r)

	host := "localhost:5000"
	fmt.Println("serving at ", host)
	err := http.ListenAndServe(host, nil)
	if err != nil {
		log.Fatal(err)
	}
}
