package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var url = "wss://api.manifold.markets/ws"

type ManifoldSocketEvent struct {
	Type  string                 `json:"type"`
	Topic string                 `json:"topic"`
	Data  map[string]interface{} `json:"data"`
}

type ManifoldEventCallback func(ManifoldSocketEvent)

var callbacks []ManifoldEventCallback

func AddManifoldWebsocketEventListener(callback ManifoldEventCallback) {
	callbacks = append(callbacks, callback)
}

func SendManifoldApiWebsocketMessage(msg string) error {
	websocketConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	err := websocketConn.WriteMessage(websocket.TextMessage, []byte(msg))

	return err
}

var websocketConn *websocket.Conn

func ConnectManifoldApiWebsocket() {
	var err error
	websocketConn, _, err = websocket.DefaultDialer.Dial(url, nil)

	if err != nil {
		log.Fatal("Unable to connect to manifold api websocket: ", err)
	}

	go func() {
		// Coroutine for receiving websocket messages
		for {
			messageType, message, err := websocketConn.ReadMessage()
			if err != nil {
				// We could try to gracefully recover from this in the future
				log.Fatal("Manifold api websocket connection fatal error: read:", err, ". messageType: ", messageType)
			}

			//println(string(message))

			var event ManifoldSocketEvent
			json.Unmarshal(message, &event)

			for _, callback := range callbacks {
				go callback(event)
			}
		}
	}()

	sendConnectionMessage()
	go sendHeartbeatMessages()
}

var txid = 0

func sendConnectionMessage() {
	var msg = fmt.Sprintf(`{"type":"subscribe","txid":%d,"topics":["global/new-bet"]}`, txid)
	txid++

	err := SendManifoldApiWebsocketMessage(msg)

	if err != nil {
		log.Println("sendConnectionMessage error:", err)
	}
}

func sendHeartbeatMessages() {
	for {
		time.Sleep(time.Second * 30)

		var msg = fmt.Sprintf(`{"type":"ping","txid":%d}`, txid)
		txid++

		err := SendManifoldApiWebsocketMessage(msg)

		if err != nil {
			log.Println("sendHeartbeatMessages error:", err)
		}
	}
}
