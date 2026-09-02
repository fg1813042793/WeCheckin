package dingtalkh5

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	dingtalkh5 "wecheckin/backend/internal/handler/dingtalkh5"
	dingtalkh5mw "wecheckin/backend/internal/middleware/dingtalk_h5"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowinfra "wecheckin/backend/internal/modules/workflow/infrastructure"
	workflowhttp "wecheckin/backend/internal/modules/workflow/transport/httpclient"
	"wecheckin/backend/internal/routes/v2/routeparam"
	"wecheckin/backend/pkg/database"
)

func Register(h *server.Hertz) {
	handler := dingtalkh5.NewHandler()
	db := database.GetDB()
	workflowStore := workflowinfra.NewGormStore(db)
	workflowNotificationRepository := workflowinfra.NewGormNotificationRepository(db)
	workflowNotificationDispatcher := workflowapp.NewNotificationDispatcher(
		workflowNotificationRepository,
		workflowinfra.NewDingTalkNotificationChannel(db, nil),
	)
	workflowService := workflowapp.NewServiceWithNotifications(
		workflowStore,
		workflowinfra.NewAssigneeResolver(db),
		workflowinfra.NewRandomIDGenerator(),
		workflowapp.DefaultLifecycleEventPublisher(),
		workflowNotificationDispatcher,
	)
	workflowHandler := workflowhttp.NewRuntimeHandler(workflowService)
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
	auth.POST("/reviews/:id/save-self", routeparam.WithBodyOrFormParam("id", "id", handler.Review.SaveSelf))
	auth.POST("/reviews/:id/submit-self", routeparam.WithBodyOrFormParam("id", "id", handler.Review.SubmitSelf))
	auth.POST("/reviews/:id/submit-manager", routeparam.WithBodyOrFormParam("id", "id", handler.Review.SubmitManager))
	auth.POST("/reviews/:id/submit-hrbp", routeparam.WithBodyOrFormParam("id", "id", handler.Review.SubmitHRBP))
	auth.POST("/reviews/:id/confirm-result", routeparam.WithBodyOrFormParam("id", "id", handler.Review.ConfirmResult))
	auth.POST("/reviews/:id/dispute-result", routeparam.WithBodyOrFormParam("id", "id", handler.Review.DisputeResult))
	auth.POST("/reviews/:id/withdraw", routeparam.WithBodyOrFormParam("id", "id", handler.Review.Withdraw))
	auth.POST("/reviews/:id/return-employee", routeparam.WithBodyOrFormParam("id", "id", handler.Review.ReturnEmployee))
	auth.POST("/reviews/:id/return-manager", routeparam.WithBodyOrFormParam("id", "id", handler.Review.ReturnManager))
	auth.POST("/reviews/:id/return-hrbp", routeparam.WithBodyOrFormParam("id", "id", handler.Review.ReturnHRBP))
	auth.POST("/reviews/:id/finalize", routeparam.WithBodyOrFormParam("id", "id", handler.Review.Finalize))
	auth.GET("/users", handler.User.ListUsers)
	auth.POST("/users", handler.User.CreateUser)
	auth.PUT("/users/:id", routeparam.WithBodyOrFormParam("id", "id", handler.User.UpdateUser))
	auth.DELETE("/users/:id", handler.User.DeleteUser)
	auth.GET("/template", handler.Template.Template)
	auth.PUT("/template", handler.Template.SaveTemplate)
	auth.GET("/workflows/definitions", workflowHandler.ListDefinitions)
	auth.GET("/workflows/definitions/:id", workflowHandler.GetDefinition)
	auth.GET("/workflows/drafts/:definitionId", workflowHandler.GetStartDraft)
	auth.PUT("/workflows/drafts/:definitionId", workflowHandler.SaveStartDraft)
	auth.POST("/workflows/instances", workflowHandler.StartInstance)
	auth.GET("/workflows/instances", workflowHandler.ListMyInstances)
	auth.GET("/workflows/instances/:id", workflowHandler.GetMyInstance)
	auth.POST("/workflows/instances/:id/withdraw", workflowHandler.WithdrawInstance)
	auth.GET("/workflows/tasks", workflowHandler.ListMyTasks)
	auth.POST("/workflows/tasks/:id/complete", workflowHandler.CompleteTask)
}
