// Package websockets provides convenience mechanisms for handling messages
// on clients
package websockets

import (
	"context"
	"encoding/json"
	"sync"
)

type CnxState int

const (
	DISCONNECTED CnxState = iota
	CONNECTED
)

type Client struct {
	state CnxState
	// methods for sending
	// methods for registering handlers
	ctx *context.Context
}

type Clients struct {
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

type Handler func(*Client, Message)
