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
			Id        string `json:"id"`
			Status    string `json:"status"`
			Type      string `json:"type"`
			Version   string `json:"version"`
			Cost      int    `json:"cost"`
			CreatedAt string `json:"created_at"`
			Condition struct {
				BroadcasterUserId string `json:"broadcaster_user_id"`
				UserId            string `json:"user_id"`
			} `json:"condition"`
			Transport any `json:"transport"`
		} `json:"subscription"`
		Event any `json:"event"`
	} `json:"payload"`
}
