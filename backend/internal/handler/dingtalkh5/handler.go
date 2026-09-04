package dingtalkh5

import (
	accounthandler "wecheckin/backend/internal/handler/dingtalkh5/account"
	authhandler "wecheckin/backend/internal/handler/dingtalkh5/auth"
	bootstraphandler "wecheckin/backend/internal/handler/dingtalkh5/bootstrap"
	notificationhandler "wecheckin/backend/internal/handler/dingtalkh5/notification"
	reviewhandler "wecheckin/backend/internal/handler/dingtalkh5/performance/review"
	templatehandler "wecheckin/backend/internal/handler/dingtalkh5/performance/template"
	userhandler "wecheckin/backend/internal/handler/dingtalkh5/performance/user"
	workflowattachmenthandler "wecheckin/backend/internal/handler/dingtalkh5/workflowattachment"
)

type Handler struct {
	Auth               *authhandler.Handler
	Account            *accounthandler.Handler
	Bootstrap          *bootstraphandler.Handler
	Notification       *notificationhandler.Handler
	Review             *reviewhandler.Handler
	Template           *templatehandler.Handler
	User               *userhandler.Handler
	WorkflowAttachment *workflowattachmenthandler.Handler
}

func NewHandler() *Handler {
	return &Handler{
		Auth:               authhandler.NewHandler(),
		Account:            accounthandler.NewHandler(),
		Bootstrap:          bootstraphandler.NewHandler(),
		Notification:       notificationhandler.NewHandler(),
		Review:             reviewhandler.NewHandler(),
		Template:           templatehandler.NewHandler(),
		User:               userhandler.NewHandler(),
		WorkflowAttachment: workflowattachmenthandler.NewHandler(),
	}
}
