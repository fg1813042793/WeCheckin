package main

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"wecheckin-backend/backend/internal/app/handler"
	"wecheckin-backend/backend/pkg/response"
)

func registerPublicRoutes(h *server.Hertz) {
	hm := handler.NewHomeHandler()
	pp := handler.NewPassportHandler()
	geo := handler.NewGeoHandler()
	aUser := handler.NewAdminUserHandler()
	aDict := handler.NewAdminDictHandler()
	cSurvey := handler.NewClientSurveyHandler()

	h.GET("/test/test", func(ctx context.Context, c *app.RequestContext) {
		response.JSON(c, map[string]string{"msg": "ok"})
	})

	h.POST("/survey/apply", cSurvey.PublicApply)
	h.POST("/survey/validate", cSurvey.PublicValidate)

	h.GET("/home/setup_get", hm.GetSetup)
	h.GET("/home/list", hm.GetHomeList)
	h.GET("/user_form_fields", aUser.GetUserFormFields)

	h.POST("/passport/login", pp.Login)
	h.POST("/passport/login_pwd", pp.LoginByPwd)
	h.POST("/passport/register", pp.Register)
	h.GET("/geo/reverse", geo.ReverseGeocode)

	h.GET("/dict/types", aDict.GetDictTypes)
	h.GET("/dict/items", aDict.GetDictByType)
}
