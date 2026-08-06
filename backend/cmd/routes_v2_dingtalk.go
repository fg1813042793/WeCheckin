package main

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	dingtalkh5 "wecheckin/backend/internal/handler/dingtalkh5"
	dingtalkh5mw "wecheckin/backend/internal/middleware/dingtalk_h5"
)

func registerV2DingTalkH5Routes(h *server.Hertz) {
	handler := dingtalkh5.NewHandler()
	group := h.Group("/api/v2/dingtalk/h5")
	group.GET("/public-config", handler.Auth.PublicConfig)
	group.POST("/login", handler.Auth.Login)
	group.POST("/sso-login", handler.Auth.SSOLogin)
	group.POST("/bind-self", handler.Auth.BindSelf)

	authed := group.Group("", dingtalkh5mw.DingTalkH5Auth())
	authed.POST("/logout", handler.Auth.Logout)
	authed.POST("/account/avatar", handler.Account.UploadAvatar)
	authed.PATCH("/account/profile", handler.Account.UpdateProfile)
	authed.PATCH("/account/password", handler.Account.ChangePassword)

	auth := group.Group("", dingtalkh5mw.DingTalkH5Auth(), dingtalkh5mw.DingTalkH5Perm())
	auth.GET("/bootstrap", handler.Bootstrap.Bootstrap)
	auth.GET("/workbench", handler.Bootstrap.Workbench)
	auth.GET("/reviews", handler.Review.ListReviews)
	auth.POST("/reviews", handler.Review.CreateReview)
	auth.GET("/reviews/export", handler.Review.ExportReviews)
	auth.GET("/reviews/:id", handler.Review.ReviewDetail)
	auth.DELETE("/reviews/:id", handler.Review.DeleteReview)
	auth.POST("/reviews/:id/save-self", withBodyOrFormParam("id", "id", handler.Review.SaveSelf))
	auth.POST("/reviews/:id/submit-self", withBodyOrFormParam("id", "id", handler.Review.SubmitSelf))
	auth.POST("/reviews/:id/submit-manager", withBodyOrFormParam("id", "id", handler.Review.SubmitManager))
	auth.POST("/reviews/:id/submit-hrbp", withBodyOrFormParam("id", "id", handler.Review.SubmitHRBP))
	auth.POST("/reviews/:id/confirm-result", withBodyOrFormParam("id", "id", handler.Review.ConfirmResult))
	auth.POST("/reviews/:id/dispute-result", withBodyOrFormParam("id", "id", handler.Review.DisputeResult))
	auth.POST("/reviews/:id/withdraw", withBodyOrFormParam("id", "id", handler.Review.Withdraw))
	auth.POST("/reviews/:id/return-employee", withBodyOrFormParam("id", "id", handler.Review.ReturnEmployee))
	auth.POST("/reviews/:id/return-manager", withBodyOrFormParam("id", "id", handler.Review.ReturnManager))
	auth.POST("/reviews/:id/return-hrbp", withBodyOrFormParam("id", "id", handler.Review.ReturnHRBP))
	auth.POST("/reviews/:id/finalize", withBodyOrFormParam("id", "id", handler.Review.Finalize))
	auth.GET("/users", handler.User.ListUsers)
	auth.POST("/users", handler.User.CreateUser)
	auth.PUT("/users/:id", withBodyOrFormParam("id", "id", handler.User.UpdateUser))
	auth.DELETE("/users/:id", handler.User.DeleteUser)
	auth.GET("/template", handler.Template.Template)
	auth.PUT("/template", handler.Template.SaveTemplate)
}
