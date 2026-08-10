package client

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	admindict "wecheckin/backend/internal/handler/admin/dict"
	adminuser "wecheckin/backend/internal/handler/admin/user"
	clientpassport "wecheckin/backend/internal/handler/client/passport"
	clientsurvey "wecheckin/backend/internal/handler/client/survey"
	publicgeo "wecheckin/backend/internal/handler/public/geo"
	publichome "wecheckin/backend/internal/handler/public/home"
	"wecheckin/backend/pkg/response"
)

func registerPublicRoutes(h *server.Hertz) {
	hm := publichome.NewHomeHandler()
	pp := clientpassport.NewPassportHandler()
	geo := publicgeo.NewGeoHandler()
	aUser := adminuser.NewAdminUserHandler()
	aDict := admindict.NewAdminDictHandler()
	cSurvey := clientsurvey.NewClientSurveyHandler()

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
