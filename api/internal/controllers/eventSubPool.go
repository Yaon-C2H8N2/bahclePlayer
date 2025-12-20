package controllers

import (
	"fmt"

	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
)

type EventSubPool struct {
	pool                map[string]*EventSub
	defaultWebSocketUrl string
	apiWrapper          *ApiWrapper
}

func GetEventSubPool(defaultWebSocketUrl string, apiWrapper *ApiWrapper) *EventSubPool {
	return &EventSubPool{
		pool:                make(map[string]*EventSub),
		defaultWebSocketUrl: defaultWebSocketUrl,
		apiWrapper:          apiWrapper,
	}
}

func (esp *EventSubPool) AddEventSub(user models.Users) error {
	es, err := GetEventSub(esp.apiWrapper, user, esp.defaultWebSocketUrl)
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
		es, err = GetEventSub(esp.apiWrapper, user, esp.defaultWebSocketUrl)
		if err != nil {
			fmt.Printf("Error getting EventSub for user %s: %s\n", user.Username, err)
			return err
		}
	}

	var onStart = func(this *EventSub, eventData any) {
		fmt.Printf("EventSub started for user %s\n", this.user.Username)
	}
	var onRefresh = func(this *EventSub, eventData any) {
		fmt.Printf("EventSub refreshed for user %s\n", this.user.Username)
	}
	var onError = func(this *EventSub, eventData any) {
		err, _ := eventData.(error)
		fmt.Printf("EventSub error for user %s: %s\n", this.user.Username, err)
	}
	var onStop = func(this *EventSub, eventData any) {
		fmt.Printf("EventSub stopped for user %s\n", this.user.Username)
	}
	es.AddEventListener(EventListenerOnStarted, onStart)
	es.AddEventListener(EventListenerOnRefresh, onRefresh)
	es.AddEventListener(EventListenerOnError, onError)
	es.AddEventListener(EventListenerOnStopped, onStop)
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

	return nil
}
