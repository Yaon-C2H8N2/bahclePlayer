package controllers

import (
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
)

type EventSubPool struct {
	pool                map[string]*EventSub
	defaultWebSocketUrl string
}

func GetEventSubPool(defaultWebSocketUrl string) *EventSubPool {
	return &EventSubPool{
		pool:                make(map[string]*EventSub),
		defaultWebSocketUrl: defaultWebSocketUrl,
	}
}

func (esp *EventSubPool) AddEventSub(apiWrapper *ApiWrapper, user models.Users) error {
	es, err := GetEventSub(apiWrapper, user, esp.defaultWebSocketUrl)
	if err != nil {
		tokenResponse, err := RefreshUserToken(user.RefreshToken)
		if err != nil {
			fmt.Printf("Error refreshing token for user %s: %s\n", user.Username, err)
			return err
		}
		fmt.Printf("Refreshing token for user %s\n", user.Username)

		oldUser := user
		user, err = models.AddOrUpdateUser(user, *tokenResponse)
		if err != nil {
			fmt.Printf("Error updating user %s: %s\n", oldUser.Username, err)
			return err
		}
		es, err = GetEventSub(apiWrapper, user, esp.defaultWebSocketUrl)
		if err != nil {
			fmt.Printf("Error getting EventSub for user %s: %s\n", user.Username, err)
			return err
		}
	}

	es.onStarted = func(this *EventSub) {
		this.DropAllSubscriptions()
		this.InitSubscriptions()
	}
	es.onRefresh = func(this *EventSub, reconnectUrl string) {
		esp.refreshEventSub(this, reconnectUrl)
	}
	es.onError = func(this *EventSub, err error) {
		//todo : error logic
	}
	es.Start()

	esp.pool[user.TwitchId] = es
	return nil
}

func (esp *EventSubPool) UpdateUser(user models.Users) error {
	eventSub, exists := esp.pool[user.TwitchId]
	if !exists {
		return fmt.Errorf("eventSub not found for user: %s", user.TwitchId)
	}
	err := eventSub.SetUser(user)
	if err != nil {
		return fmt.Errorf("error updating user %s: %s", user.Username, err)
	}

	//esp.refreshEventSub(eventSub, esp.defaultWebSocketUrl) // Uncomment to force refresh the EventSub along the user update

	return nil
}

func (esp *EventSubPool) refreshEventSub(oldEventSub *EventSub, reconnectUrl string) {
	if oldEventSub == nil {
		return
	}
	fmt.Printf("Refreshing EventSub for user %s\n", oldEventSub.user.Username)
	newEventSub, err := GetEventSub(oldEventSub.apiWrapper, oldEventSub.user, reconnectUrl)
	if err != nil {
		fmt.Printf("Error refreshing EventSub for user %s: %s\n", oldEventSub.user.Username, err)
		return
	}
	newEventSub.onStarted = func(this *EventSub) {
		oldEventSub.DropAllSubscriptions()
		this.InitSubscriptions()
		oldEventSub.Stop()
		esp.pool[oldEventSub.user.TwitchId] = newEventSub
	}
	newEventSub.onRefresh = oldEventSub.onRefresh
	newEventSub.onError = oldEventSub.onError
	newEventSub.Start()
}
