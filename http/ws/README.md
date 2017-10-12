# Websockets

This package provides utilities for handling web socket clients, and registering listeners for application events.  It builds on top of the `gorilla/websocket` package.

Example usage:

```go
package main

import (
	"github.com/globalprofessionalsearch/go-tools/http/ws"
)

func main () {
	clientPool := ws.NewClientPool()

	// create an http handler for upgrading connection, wrap it in middlewares if you want for any
	// auth handling
	httpHandler := ws.NewUpgradeHandler(func(client *ws.Client, req *http.Request) {
		// the pool will automatically listen for a disconnect event
		clientPool.Register(client)

		// handle connection events: notify all clients in pool of this user connecting/disconnecting
		client.On(ws.DISCONNECTED, func() {
			user, _ := c.Context().Get('user').(appUser)
			data := struct{Id string}{user.Id}
			clientPool.Broadcast("participants.disconnected", data)
		})
		client.On(ws.CONNECTED, func() {
			client.Send("app.connected", nil)
			user, _ := c.Context().Get('user').(appUser)
			data := struct{Id string}{user.Id}
			clientPool.Broadcast("participants.connected", data)
		})
		client.OnError(func(err error) {
			// TODO: maybe encoding a message failed or something...?
		})

		// register app event handlers
		client.Handle("chats.msg", incomingMsgHandler)
		client.Handle("chats.msg-private", incomingPrivateMsgHandler)
		client.Handle("participants.info", getParticipantInfoHandler)
		client.Handle("app.signout", signoutHandler)
	})

	http.ListenAndServe(":80", httpHandler)

	// listen for interrupt, disconnect connected clients
}

func incomingChatMsgHandler(client *ws.Client, msg ws.Message) {
	
}

```