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
		ws.On(ws.DISCONNECTED, func() {
			user, _ := c.Context().Get('user').(appUser)
			data := struct{Id string}{user.Id}
			clientPool.Broadcast("participant.disconnected", data)
		})
		ws.On(ws.CONNECTED, func() {
			ws.Send("app.connected", ws.NewMessage())
		})

		// register app event handlers
		ws.Handle("chats.msg", incomingMsgHandler)
		ws.Handle("chats.msg-private", incomingPrivateMsgHandler)
		ws.Handle("participants.info", getParticipantInfoHandler)
		ws.Handle("app.signout", signoutHandler)
	})

	http.ListenAndServe(":80", httpHandler)
}

func incomingChatMsgHandler(client *ws.Client, msg ws.Message) {
	
}

```