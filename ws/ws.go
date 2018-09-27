// Package websockets provides convenience mechanisms for handling messages
// on clients
package websockets

import (
	"context"
	"encoding/json"
	"sync"
)

type ClientEvent int

const (
	DISCONNECTED ClientEvent = iota
	CONNECTED
	ERROR
)

type Client struct {
	//client *websocket.Client
	mtx         *sync.RWMutex
	ctx         *context.Context
	listeners   map[ClientEvent]func()
	errListener func(error)
	handlers    map[string]MessageHandler
}

func (c *Client) trigger()                                   {}
func (c *Client) triggerError()                              {}
func (c *Client) On(evt ClientEvent, listener func(Message)) {}
func (c *Client) OnError(func(error))                        {}
func (c *Client) SendMessage(Message)                        {}

type ClientPool struct {
	// methods for registering, unregistering clients, broadcasting to group
	mtx *sync.RWMutex
	ctx *context.Context
}

type Message struct {
	Name string
	Data *json.RawMessage
	ctx  *context.Context
}

func (m Message) Context() *context.Context {
	return m.ctx
}

func (m Message) Unmarshal(target interface{}) error {
	return json.Unmarshal(*m.Data, target)
}

type MessageHandler func(*Client, Message)
