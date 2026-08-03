package main

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	dingtalkh5 "wecheckin/backend/internal/app/handler/client/dingtalkh5"
)

func registerV2DingTalkH5Routes(h *server.Hertz) {
	handler := dingtalkh5.NewHandler()
	group := h.Group("/api/v2/dingtalk/h5")
	group.GET("/public-config", handler.PublicConfig)
	group.POST("/login", handler.Login)
	group.POST("/sso-login", handler.SSOLogin)
	group.POST("/bind-self", handler.BindSelf)

	authed := group.Group("", handler.Auth())
	authed.POST("/logout", handler.Logout)
	authed.PATCH("/account/profile", handler.UpdateProfile)
	authed.PATCH("/account/password", handler.ChangePassword)

	auth := group.Group("", handler.Auth(), handler.ApiPerm())
	auth.GET("/bootstrap", handler.Bootstrap)
	auth.GET("/workbench", handler.Workbench)
	auth.GET("/reviews", handler.ListReviews)
	auth.POST("/reviews", handler.CreateReview)
	auth.GET("/reviews/export", handler.ExportReviews)
	auth.GET("/reviews/:id", handler.ReviewDetail)
	auth.DELETE("/reviews/:id", handler.DeleteReview)
	auth.POST("/reviews/:id/save-self", withBodyOrFormParam("id", "id", handler.SaveSelf))
	auth.POST("/reviews/:id/submit-self", withBodyOrFormParam("id", "id", handler.SubmitSelf))
	auth.POST("/reviews/:id/submit-manager", withBodyOrFormParam("id", "id", handler.SubmitManager))
	auth.POST("/reviews/:id/submit-hrbp", withBodyOrFormParam("id", "id", handler.SubmitHRBP))
	auth.POST("/reviews/:id/confirm-result", withBodyOrFormParam("id", "id", handler.ConfirmResult))
	auth.POST("/reviews/:id/dispute-result", withBodyOrFormParam("id", "id", handler.DisputeResult))
	auth.POST("/reviews/:id/withdraw", withBodyOrFormParam("id", "id", handler.Withdraw))
	auth.POST("/reviews/:id/return-employee", withBodyOrFormParam("id", "id", handler.ReturnEmployee))
	auth.POST("/reviews/:id/return-manager", withBodyOrFormParam("id", "id", handler.ReturnManager))
	auth.POST("/reviews/:id/return-hrbp", withBodyOrFormParam("id", "id", handler.ReturnHRBP))
	auth.POST("/reviews/:id/finalize", withBodyOrFormParam("id", "id", handler.Finalize))
	auth.GET("/users", handler.ListUsers)
	auth.POST("/users", handler.CreateUser)
	auth.PUT("/users/:id", withBodyOrFormParam("id", "id", handler.UpdateUser))
	auth.DELETE("/users/:id", handler.DeleteUser)
	auth.GET("/template", handler.Template)
	auth.PUT("/template", handler.SaveTemplate)
}
