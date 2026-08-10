package performance

import (
	accountsvc "wecheckin/backend/internal/service/dingtalkh5/account"
	authsvc "wecheckin/backend/internal/service/dingtalkh5/auth"
	bootstrapsvc "wecheckin/backend/internal/service/dingtalkh5/bootstrap"
	configsvc "wecheckin/backend/internal/service/dingtalkh5/config"
	domain "wecheckin/backend/internal/service/dingtalkh5/performance/review"
	reviewsvc "wecheckin/backend/internal/service/dingtalkh5/performance/review"
	templatesvc "wecheckin/backend/internal/service/dingtalkh5/performance/template"
	usersvc "wecheckin/backend/internal/service/dingtalkh5/performance/user"
)

const (
	ReviewStatusDraft           = domain.ReviewStatusDraft
	ReviewStatusManagerReview   = domain.ReviewStatusManagerReview
	ReviewStatusHRBPReview      = domain.ReviewStatusHRBPReview
	ReviewStatusEmployeeConfirm = domain.ReviewStatusEmployeeConfirm
	ReviewStatusHRFinal         = domain.ReviewStatusHRFinal
	ReviewStatusCompleted       = domain.ReviewStatusCompleted

	TemplateKeyDefault = templatesvc.TemplateKeyDefault

	DingTalkH5BindRequiredCode = authsvc.DingTalkH5BindRequiredCode
)

type (
	UserDTO                   = usersvc.UserDTO
	Objective                 = domain.Objective
	NextObjective             = templatesvc.NextObjective
	ValueScore                = domain.ValueScore
	HistoryDTO                = domain.HistoryDTO
	ReviewDTO                 = domain.ReviewDTO
	ReviewListResponse        = domain.ReviewListResponse
	CreateReviewFailure       = domain.CreateReviewFailure
	CreateReviewBatchResponse = domain.CreateReviewBatchResponse
	DingTalkH5Review          = domain.DingTalkH5Review
	GradeLevel                = templatesvc.GradeLevel
	ValueRubric               = templatesvc.ValueRubric
	ValueTemplate             = templatesvc.ValueTemplate
	TemplateDTO               = templatesvc.TemplateDTO
	LoginResponse             = authsvc.LoginResponse
	BootstrapResponse         = bootstrapsvc.BootstrapResponse
	DingTalkH5AppConfigDTO    = configsvc.DingTalkH5AppConfigDTO
	PublicConfigResponse      = configsvc.PublicConfigResponse
	WorkbenchStatsDTO         = bootstrapsvc.WorkbenchStatsDTO
	WorkbenchStatCardDTO      = bootstrapsvc.WorkbenchStatCardDTO
	AppMenuDTO                = bootstrapsvc.AppMenuDTO
	ReviewPayload             = domain.ReviewPayload
	UserPayload               = usersvc.UserPayload
	AccountProfilePayload     = accountsvc.AccountProfilePayload
	AccountProfileResponse    = accountsvc.AccountProfileResponse
	ReviewFilters             = domain.ReviewFilters

	DingTalkH5CorpConfig                = configsvc.DingTalkH5CorpConfig
	DingTalkUserIdentity                = configsvc.DingTalkUserIdentity
	DingTalkIdentityClient              = configsvc.DingTalkIdentityClient
	DingTalkWorkNotificationClient      = configsvc.DingTalkWorkNotificationClient
	DingTalkWorkNotificationPayload     = configsvc.DingTalkWorkNotificationPayload
	DingTalkH5BindRequiredResponse      = authsvc.DingTalkH5BindRequiredResponse
	DingTalkH5BindRequiredError         = authsvc.DingTalkH5BindRequiredError
	ExportResult                        = domain.ExportResult
	DingTalkH5NotificationDiagnosis     = configsvc.DingTalkH5NotificationDiagnosis
	DingTalkH5NotificationDiagnosisStep = configsvc.DingTalkH5NotificationDiagnosisStep
)

var (
	PublicConfigContext                       = authsvc.PublicConfigContext
	LoginContext                              = authsvc.LoginContext
	LogoutContext                             = authsvc.LogoutContext
	AuthenticateContext                       = authsvc.AuthenticateContext
	LoginByAuthCodeContext                    = authsvc.LoginByAuthCodeContext
	BindSelfContext                           = authsvc.BindSelfContext
	DingTalkH5BindRequiredData                = authsvc.DingTalkH5BindRequiredData
	SelfBindEnabledContext                    = authsvc.SelfBindEnabledContext
	BootstrapContext                          = bootstrapsvc.BootstrapContext
	DingTalkH5MenusForUserContext             = bootstrapsvc.DingTalkH5MenusForUserContext
	TemplateContext                           = templatesvc.TemplateContext
	ListReviewsContext                        = reviewsvc.ListReviewsContext
	GetReviewContext                          = reviewsvc.GetReviewContext
	CreateReviewContext                       = reviewsvc.CreateReviewContext
	CreateReviewsContext                      = reviewsvc.CreateReviewsContext
	SaveSelfContext                           = reviewsvc.SaveSelfContext
	SubmitSelfContext                         = reviewsvc.SubmitSelfContext
	SubmitManagerContext                      = reviewsvc.SubmitManagerContext
	SubmitHRBPContext                         = reviewsvc.SubmitHRBPContext
	ConfirmResultContext                      = reviewsvc.ConfirmResultContext
	DisputeResultContext                      = reviewsvc.DisputeResultContext
	FinalizeContext                           = reviewsvc.FinalizeContext
	WithdrawContext                           = reviewsvc.WithdrawContext
	ReturnEmployeeContext                     = reviewsvc.ReturnEmployeeContext
	ReturnManagerContext                      = reviewsvc.ReturnManagerContext
	ReturnHRBPContext                         = reviewsvc.ReturnHRBPContext
	DeleteReviewContext                       = reviewsvc.DeleteReviewContext
	ExportReviewsContext                      = reviewsvc.ExportReviewsContext
	ListUsersContext                          = usersvc.ListUsersContext
	CreateUserContext                         = usersvc.CreateUserContext
	UpdateUserContext                         = usersvc.UpdateUserContext
	DeleteUserContext                         = usersvc.DeleteUserContext
	LoadTemplateContext                       = templatesvc.LoadTemplateContext
	SaveTemplateContext                       = templatesvc.SaveTemplateContext
	EnsureSeedContext                         = templatesvc.EnsureSeedContext
	WorkbenchStatsContext                     = bootstrapsvc.WorkbenchStatsContext
	ListDingTalkH5CorpConfigsContext          = configsvc.ListDingTalkH5CorpConfigsContext
	SaveDingTalkH5CorpConfigsContext          = configsvc.SaveDingTalkH5CorpConfigsContext
	DingTalkH5NotificationEnabledContext      = configsvc.DingTalkH5NotificationEnabledContext
	DiagnoseDingTalkH5WorkNotificationContext = configsvc.DiagnoseDingTalkH5WorkNotificationContext
	ChangePasswordContext                     = accountsvc.ChangePasswordContext
	UpdateAccountProfileContext               = accountsvc.UpdateAccountProfileContext
	NormalizeUserID                           = accountsvc.NormalizeUserID
)
