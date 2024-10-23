package twitch

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

type EventSub struct {
	onEvent   func(event any)
	sessionId string
}

func (es *EventSub) Start() {
	es.listenToMessages()
}

func (es *EventSub) OnEvent(callback func(event any)) {
	es.onEvent = callback
}

//func SubscribeToMessageEvents() {
//	twitchUrl, _ := url.Parse("https://api.twitch.tv/helix/eventsub/subscriptions")
//
//	/*
//		curl -X POST 'https://api.twitch.tv/helix/eventsub/subscriptions' \
//		-H 'Authorization: Bearer 2gbdx6oar67tqtcmt49t3wpcgycthx' \
//		-H 'Client-Id: wbmytr93xzw8zbg0p1izqyzzc5mbiz' \
//		-H 'Content-Type: application/json' \
//		-d '{"
//		    type": "user.update",
//		    "version": "1",
//		    "condition": {
//		        "user_id": "1234"
//		    },
//		    "transport": {
//		        "method": "websocket",
//		        "session_id": "AQoQexAWVYKSTIu4ec_2VAxyuhAB"
//		    }
//		}'
//	*/
//
//	if err != nil {
//		panic(err)
//	}
//}

func (es *EventSub) listenToMessages() {
	//https://github.com/gorilla/websocket/blob/main/examples/echo/client.go
	conn, _, err := websocket.DefaultDialer.Dial("wss://eventsub.wss.twitch.tv/ws", nil)

	if err != nil {
		panic(err)
	}
	defer conn.Close()

	go func() {
		for {
			_, messageBytes, err := conn.ReadMessage()
			if err != nil {
				log.Printf("err: %s", messageBytes)
				panic(err)
			}

			var message = &BaseMessage{}
			err = json.Unmarshal(messageBytes, message)
			if err != nil {
				log.Printf("err: %s", messageBytes)
				panic(err)
			}

			switch message.Metadata.MessageType {
			case "welcome":
				var welcomeMessage = &WelcomeMessage{}
				err = json.Unmarshal(messageBytes, welcomeMessage)
				if err != nil {
					log.Printf("err: %s", messageBytes)
					panic(err)
				}
				log.Printf("Welcome message: %s", welcomeMessage.Payload.Session.Id)

				es.sessionId = welcomeMessage.Payload.Session.Id

			case "notification":
				var notificationMessage = &NotificationMessage{}
				err = json.Unmarshal(messageBytes, notificationMessage)
				if err != nil {
					log.Printf("err: %s", messageBytes)
					panic(err)
				}
				es.onEvent(notificationMessage.Payload.Event)
			}
		}
	}()
}
