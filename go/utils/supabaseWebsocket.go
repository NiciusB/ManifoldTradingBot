package utils

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// https://supabase.com/docs/guides/realtime/protocol

var SUPABASE_ANON_KEY = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InB4aWRyZ2thdHVtbHZmcWF4Y2xsIiwicm9sZSI6ImFub24iLCJpYXQiOjE2Njg5OTUzOTgsImV4cCI6MTk4NDU3MTM5OH0.d_yYtASLzAoIIGdXUBIgRAGLBnNow7JG2SoaNMQ8ySg"
var url = "wss://pxidrgkatumlvfqaxcll.supabase.co/realtime/v1/websocket?apikey=" + SUPABASE_ANON_KEY + "&vsn=1.0.0"

type SupabaseEvent struct {
	Event   string                 `json:"event"`
	Topic   string                 `json:"topic"`
	Payload map[string]interface{} `json:"payload"`
	Ref     string                 `json:"ref"`
}

type SupabaseEventCallback func(SupabaseEvent)

var callbacks []SupabaseEventCallback

func AddSupabaseWebsocketEventListener(callback SupabaseEventCallback) {
	callbacks = append(callbacks, callback)
}

func SendSupabaseWebsocketMessage(msg string) error {
	websocketConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	err := websocketConn.WriteMessage(websocket.TextMessage, []byte(msg))

	return err
}

var websocketConn *websocket.Conn

func ConnectSupabaseWebsocket() {
	var err error
	websocketConn, _, err = websocket.DefaultDialer.Dial(url, nil)

	if err != nil {
		log.Fatal("Unable to connect to supabase websocket: ", err)
	}

	go func() {
		// Coroutine for receiving websocket messages
		for {
			messageType, message, err := websocketConn.ReadMessage()
			if err != nil {
				// We could try to gracefully recover from this in the future
				log.Fatal("Supabase websocket connection fatal error: read:", err, ". messageType: ", messageType)
			}

			//println(string(message))

			var event SupabaseEvent
			json.Unmarshal(message, &event)

			for _, callback := range callbacks {
				go callback(event)
			}
		}
	}()

	sendConnectionMessage()
	go sendHeartbeatMessages()
}

func sendConnectionMessage() {
	var msg = `{
"event": "phx_join",
"topic": "realtime:*",
"payload": {
	"config": {
		"broadcast": {
			"self": false
		},
		"presence": {
			"key": ""
		}
	}
},
"ref": null
}`

	err := SendSupabaseWebsocketMessage(msg)

	if err != nil {
		log.Println("sendConnectionMessage error:", err)
	}
}

func sendHeartbeatMessages() {
	for {
		time.Sleep(time.Second * 25)

		var msg = `{
"event": "heartbeat",
"topic": "phoenix",
"payload": {},
"ref": null
}`

		err := SendSupabaseWebsocketMessage(msg)

		if err != nil {
			log.Println("sendHeartbeatMessages error:", err)
		}
	}
}
