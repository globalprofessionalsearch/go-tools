# Websockets

This package provides utilities for handling web socket clients, and registering listeners for application events.  It builds on top of the `gorilla/websocket` package.

Example usage:

```go
package main

import (
	"github.com/globalprofessionalsearch/go-tools/ws"
	"github.com/globalprofessionalsearch/go-tools/ws/cluster"
)

func main () {
	channel := cluster.Channel("my-channel", redis)

	// create an http handler for upgrading connection, wrap it in middlewares if you want for any
	// auth handling
	httpHandler := ws.NewUpgradeHandler(func(client *ws.Client, req *http.Request) {
		user, _ := req.Context().Value('user').(appUser)

		// the pool will automatically listen for a disconnect event
		channel.RegisterClient("<userID>", client) // subscribes to Broadcast, and enables Send to client
		channel.UnregisterClient("<userID>")
		channel.Send("<anotherID>", "event.name", MsgPayload) // send to specific cnx in channel
		channel.Broadcast("event.name", MsgPayload) // send to all clients in channel
		channel.BroadcastLocal("event.name", MsgPayload) // send to clients directly connected to this channel (on this server)
		channel.Disconnect() // disconnect all clients

		client.Broadcast("event.name", MsgPayload) // send to all other clients in channels
		client.Send("event.name", MsgPayload) // send a message directly to client
		client.Disconnect() // disconnect client, and remove from all channels

		// handle connection events: notify all clients in pool of this user connecting/disconnecting
		client.On(ws.DISCONNECTED, func() {
			data := struct{Id string}{user.Id}
			channel.Broadcast("participants.disconnected", data)
		})
		client.On(ws.CONNECTED, func() {
			client.Send("app.connected", nil)
			data := struct{Id string}{user.Id}
			client.Broadcast("participants.connected", data)
		})
		client.OnError(func(err error) {
			// TODO: maybe encoding a message failed or something...?
			// custom error type to pass the failed message if relevant?
		})

		// register app event handlers
		client.HandleMessage("chats.msg", incomingMsgHandler)
		client.HandleMessage("chats.msg-private", incomingPrivateMsgHandler)
		client.HandleMessage("participants.info", getParticipantInfoHandler)
		client.HandleMessage("app.signout", signoutHandler)

		go client.Start()  // right?
	})

	http.ListenAndServe(":80", httpHandler)
	channel.Broadcast("app.reconnect")
	// listen for interrupt, disconnect connected clients
}

func incomingChatMsgHandler(client *ws.Client, msg ws.Message) {
	var payload map[string]interface{}
	if err := msg.Unmarshal(&payload); err != nil {
		
	}
}

```
