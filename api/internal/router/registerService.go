package router

import (
	"reflect"

	"github.com/Yaon-C2H8N2/bahclePlayer/internal/context"
	"github.com/gin-gonic/gin"
)

type HandlerFunction func(c *gin.Context, appContext *context.AppContext)

func RegisterService(router *gin.Engine, appContext *context.AppContext, service any) {
	fields := reflect.TypeOf(service).Elem()
	for i := 0; i < fields.NumField(); i++ {
		field := fields.Field(i)

		routeTag := field.Tag.Get("method")
		pathTag := field.Tag.Get("path")

		if routeTag != "" && pathTag != "" {
			method := reflect.ValueOf(service).Elem().FieldByName(field.Name)
			if method.IsValid() && method.Kind() == reflect.Func {
				router.Handle(routeTag, pathTag, func(c *gin.Context) {
					method.Call([]reflect.Value{reflect.ValueOf(c), reflect.ValueOf(appContext)})
				})
			}
		}
	}
}
