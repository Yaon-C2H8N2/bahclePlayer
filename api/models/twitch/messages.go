package twitch

import "time"

type BaseMessage struct {
	Metadata struct {
		MessageId        string    `json:"message_id"`
		MessageType      string    `json:"message_type"`
		MessageTimestamp time.Time `json:"message_timestamp"`
	} `json:"metadata"`
}

type WelcomeMessage struct {
	BaseMessage
	Payload struct {
		Session struct {
			Id                      string    `json:"id"`
			Status                  string    `json:"status"`
			ConnectedAt             time.Time `json:"connected_at"`
			KeepAliveTimeoutSeconds int       `json:"keep_alive_timeout_seconds"`
			ReconnectUrl            string    `json:"reconnect_url"`
		} `json:"session"`
	} `json:"payload"`
}

type NotificationMessage struct {
	BaseMessage
	Payload struct {
		Subscription struct {
		} `json:"subscription"`
		Event struct {
		} `json:"event"`
	} `json:"payload"`
}
