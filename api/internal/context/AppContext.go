package context

import (
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/controllers"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
)

type AppContext struct {
	ApiWrapper     *controllers.ApiWrapper
	EventSubPool   *controllers.EventSubPool
	PlayersManager *controllers.PlayersManager
	AppStatus      *models.AppStatus
}
