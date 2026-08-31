package swagger

import "wecheckin/backend/pkg/response"

var _ response.Resp

// @Tags API v2-公开接口
// @Summary 查询 /api/v2/home
// @Success 200 {object} response.Resp
// @Router /api/v2/home [get]
func swaggerV2HomeGet0() {}

// @Tags API v2-公开接口
// @Summary 查询 /api/v2/home/setup
// @Success 200 {object} response.Resp
// @Router /api/v2/home/setup [get]
func swaggerV2HomeSetupGet1() {}

// @Tags API v2-公开接口
// @Summary 查询 /api/v2/user-form-fields
// @Success 200 {object} response.Resp
// @Router /api/v2/user-form-fields [get]
func swaggerV2UserFormFieldsGet2() {}

// @Tags API v2-公开接口
// @Summary 提交 /api/v2/auth/login
// @Success 200 {object} response.Resp
// @Router /api/v2/auth/login [post]
func swaggerV2AuthLoginPost3() {}

// @Tags API v2-公开接口
// @Summary 提交 /api/v2/auth/password-login
// @Success 200 {object} response.Resp
// @Router /api/v2/auth/password-login [post]
func swaggerV2AuthPasswordLoginPost4() {}

// @Tags API v2-公开接口
// @Summary 提交 /api/v2/auth/register
// @Success 200 {object} response.Resp
// @Router /api/v2/auth/register [post]
func swaggerV2AuthRegisterPost5() {}

// @Tags API v2-公开接口
// @Summary 查询 /api/v2/geo/reverse
// @Success 200 {object} response.Resp
// @Router /api/v2/geo/reverse [get]
func swaggerV2GeoReverseGet6() {}

// @Tags API v2-公开接口
// @Summary 查询 /api/v2/dict/types
// @Success 200 {object} response.Resp
// @Router /api/v2/dict/types [get]
func swaggerV2DictTypesGet7() {}

// @Tags API v2-公开接口
// @Summary 查询 /api/v2/dict/items
// @Success 200 {object} response.Resp
// @Router /api/v2/dict/items [get]
func swaggerV2DictItemsGet8() {}

// @Tags API v2-公开接口
// @Summary 查询 /api/v2/events
// @Success 200 {object} response.Resp
// @Router /api/v2/events [get]
func swaggerV2EventsGet9() {}

// @Tags API v2-公开接口
// @Summary 查询 /api/v2/events/{id}
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/events/{id} [get]
func swaggerV2EventsIdGet10() {}

// @Tags API v2-公开接口
// @Summary 查询 /api/v2/surveys
// @Success 200 {object} response.Resp
// @Router /api/v2/surveys [get]
func swaggerV2SurveysGet11() {}

// @Tags API v2-公开接口
// @Summary 查询 /api/v2/surveys/{id}
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/surveys/{id} [get]
func swaggerV2SurveysIdGet12() {}

// @Tags API v2-公开接口
// @Summary 提交 /api/v2/surveys/{id}/responses
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/surveys/{id}/responses [post]
func swaggerV2SurveysIdResponsesPost13() {}

// @Tags API v2-公开接口
// @Summary 提交 /api/v2/survey/apply
// @Success 200 {object} response.Resp
// @Router /api/v2/survey/apply [post]
func swaggerV2SurveyApplyPost14() {}

// @Tags API v2-公开接口
// @Summary 提交 /api/v2/survey/validate
// @Success 200 {object} response.Resp
// @Router /api/v2/survey/validate [post]
func swaggerV2SurveyValidatePost15() {}

// @Tags API v2-公开接口
// @Summary 查询 /api/v2/exams
// @Success 200 {object} response.Resp
// @Router /api/v2/exams [get]
func swaggerV2ExamsGet16() {}

// @Tags API v2-公开接口
// @Summary 查询 /api/v2/exams/{id}
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/exams/{id} [get]
func swaggerV2ExamsIdGet17() {}

// @Tags API v2-公开接口
// @Summary 提交 /api/v2/exams/{id}/submissions
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/exams/{id}/submissions [post]
func swaggerV2ExamsIdSubmissionsPost18() {}

// @Tags API v2-公开接口
// @Summary 提交 /api/v2/exams/{id}/validation
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/exams/{id}/validation [post]
func swaggerV2ExamsIdValidationPost19() {}

// @Tags API v2-公开接口
// @Summary 查询 /api/v2/exam-results
// @Success 200 {object} response.Resp
// @Router /api/v2/exam-results [get]
func swaggerV2ExamResultsGet20() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/me/bootstrap
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/bootstrap [get]
func swaggerV2MeBootstrapGet201() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/me
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me [get]
func swaggerV2MeGet21() {}

// @Tags API v2-客户端
// @Summary 更新 /api/v2/me
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me [put]
func swaggerV2MePut22() {}

// @Tags API v2-客户端
// @Summary 提交 /api/v2/me/phone
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/phone [post]
func swaggerV2MePhonePost23() {}

// @Tags API v2-客户端
// @Summary 提交 /api/v2/me/logout
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/logout [post]
func swaggerV2MeLogoutPost24() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/me/favorites
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/favorites [get]
func swaggerV2MeFavoritesGet25() {}

// @Tags API v2-客户端
// @Summary 提交 /api/v2/me/favorites
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/favorites [post]
func swaggerV2MeFavoritesPost26() {}

// @Tags API v2-客户端
// @Summary 删除 /api/v2/me/favorites/{oid}
// @Security ClientToken
// @Param oid path string true "oid"
// @Success 200 {object} response.Resp
// @Router /api/v2/me/favorites/{oid} [delete]
func swaggerV2MeFavoritesOidDelete27() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/me/favorites/check
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/favorites/check [get]
func swaggerV2MeFavoritesCheckGet28() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/me/enrollments
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/enrollments [get]
func swaggerV2MeEnrollmentsGet29() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/me/enrollment-users
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/enrollment-users [get]
func swaggerV2MeEnrollmentUsersGet30() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/me/enrollment-records
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/enrollment-records [get]
func swaggerV2MeEnrollmentRecordsGet31() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/me/enrollment-calendar
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/enrollment-calendar [get]
func swaggerV2MeEnrollmentCalendarGet32() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/me/enrollment-day-records
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/enrollment-day-records [get]
func swaggerV2MeEnrollmentDayRecordsGet33() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/me/events
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/events [get]
func swaggerV2MeEventsGet34() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/me/event-roles
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/event-roles [get]
func swaggerV2MeEventRolesGet35() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/me/managed-events
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/managed-events [get]
func swaggerV2MeManagedEventsGet36() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/me/survey-responses
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/survey-responses [get]
func swaggerV2MeSurveyResponsesGet37() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/me/survey-responses/{id}
// @Security ClientToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/me/survey-responses/{id} [get]
func swaggerV2MeSurveyResponsesIdGet38() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/me/exam-records
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/exam-records [get]
func swaggerV2MeExamRecordsGet39() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/news
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/news [get]
func swaggerV2NewsGet40() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/news/categories
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/news/categories [get]
func swaggerV2NewsCategoriesGet41() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/news/{id}
// @Security ClientToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/news/{id} [get]
func swaggerV2NewsIdGet42() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/enrollments
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/enrollments [get]
func swaggerV2EnrollmentsGet43() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/enrollments/{id}
// @Security ClientToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/enrollments/{id} [get]
func swaggerV2EnrollmentsIdGet44() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/enrollments/{id}/join-days
// @Security ClientToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/enrollments/{id}/join-days [get]
func swaggerV2EnrollmentsIdJoinDaysGet45() {}

// @Tags API v2-客户端
// @Summary 提交 /api/v2/enrollments/{id}/joins
// @Security ClientToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/enrollments/{id}/joins [post]
func swaggerV2EnrollmentsIdJoinsPost46() {}

// @Tags API v2-客户端
// @Summary 提交 /api/v2/enrollments/{id}/submissions
// @Security ClientToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/enrollments/{id}/submissions [post]
func swaggerV2EnrollmentsIdSubmissionsPost47() {}

// @Tags API v2-客户端
// @Summary 提交 /api/v2/events/{id}/participants
// @Security ClientToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/events/{id}/participants [post]
func swaggerV2EventsIdParticipantsPost48() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/events/{id}/participants
// @Security ClientToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/events/{id}/participants [get]
func swaggerV2EventsIdParticipantsGet49() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/events/{id}/dynamics
// @Security ClientToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/events/{id}/dynamics [get]
func swaggerV2EventsIdDynamicsGet50() {}

// @Tags API v2-客户端
// @Summary 提交 /api/v2/events/{id}/dynamics
// @Security ClientToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/events/{id}/dynamics [post]
func swaggerV2EventsIdDynamicsPost51() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/events/{id}/scores
// @Security ClientToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/events/{id}/scores [get]
func swaggerV2EventsIdScoresGet52() {}

// @Tags API v2-客户端
// @Summary 提交 /api/v2/events/{id}/scores
// @Security ClientToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/events/{id}/scores [post]
func swaggerV2EventsIdScoresPost53() {}

// @Tags API v2-客户端
// @Summary 提交 /api/v2/exams/{id}/start
// @Security ClientToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/exams/{id}/start [post]
func swaggerV2ExamsIdStartPost54() {}

// @Tags API v2-客户端
// @Summary 查询 /api/v2/exam-records/{id}
// @Security ClientToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/exam-records/{id} [get]
func swaggerV2ExamRecordsIdGet55() {}

// @Tags API v2-客户端
// @Summary 更新 /api/v2/exam-records/{id}/answers
// @Security ClientToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/exam-records/{id}/answers [put]
func swaggerV2ExamRecordsIdAnswersPut56() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/auth/login
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/auth/login [post]
func swaggerV2AdminAuthLoginPost57() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/home
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/home [get]
func swaggerV2AdminHomeGet58() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/home/recommendations
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/home/recommendations [delete]
func swaggerV2AdminHomeRecommendationsDelete59() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/managers
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/managers [get]
func swaggerV2AdminManagersGet60() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/managers
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/managers [post]
func swaggerV2AdminManagersPost61() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/managers/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/managers/{id} [get]
func swaggerV2AdminManagersIdGet62() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/managers/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/managers/{id} [put]
func swaggerV2AdminManagersIdPut63() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/managers/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/managers/{id} [delete]
func swaggerV2AdminManagersIdDelete64() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/managers
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/managers [delete]
func swaggerV2AdminManagersDelete65() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/managers/{id}/status
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/managers/{id}/status [patch]
func swaggerV2AdminManagersIdStatusPatch66() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/managers/{id}/password
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/managers/{id}/password [patch]
func swaggerV2AdminManagersIdPasswordPatch67() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/admin-sessions
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/admin-sessions [get]
func swaggerV2AdminAdminSessionsGet68() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/admin-sessions/{id}/force-offline
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/admin-sessions/{id}/force-offline [post]
func swaggerV2AdminAdminSessionsIdForceOfflinePost69() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/admin-sessions/batch-force-offline
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/admin-sessions/batch-force-offline [post]
func swaggerV2AdminAdminSessionsBatchForceOfflinePost70() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/auth/logout
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/auth/logout [post]
func swaggerV2AdminAuthLogoutPost71() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/logs
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/logs [get]
func swaggerV2AdminLogsGet72() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/logs
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/logs [delete]
func swaggerV2AdminLogsDelete73() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/settings
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/settings [put]
func swaggerV2AdminSettingsPut74() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/settings/content
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/settings/content [put]
func swaggerV2AdminSettingsContentPut75() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/settings/mini-qr
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/settings/mini-qr [get]
func swaggerV2AdminSettingsMiniQrGet76() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/settings/debug-token
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/settings/debug-token [get]
func swaggerV2AdminSettingsDebugTokenGet77() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/dingtalk/settings
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/settings [get]
func swaggerV2AdminDingTalkSettingsGet() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/dingtalk/settings
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/settings [put]
func swaggerV2AdminDingTalkSettingsPut() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/dingtalk/settings/notification-test
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/settings/notification-test [post]
func swaggerV2AdminDingTalkSettingsNotificationTestPost() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/dingtalk/user-bindings
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/user-bindings [get]
func swaggerV2AdminDingTalkUserBindingsGet() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/dingtalk/user-bindings
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/user-bindings [post]
func swaggerV2AdminDingTalkUserBindingsPost() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/dingtalk/user-bindings/{id}/status
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/user-bindings/{id}/status [patch]
func swaggerV2AdminDingTalkUserBindingsIDStatusPatch() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/dingtalk/user-bindings/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/user-bindings/{id} [delete]
func swaggerV2AdminDingTalkUserBindingsIDDelete() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/dingtalk/perf-reviews
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/perf-reviews [get]
func swaggerV2AdminDingTalkPerfReviewsGet() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/dingtalk/perf-reviews/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/perf-reviews/{id} [get]
func swaggerV2AdminDingTalkPerfReviewsIDGet() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/dingtalk/perf-reviews/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/perf-reviews/{id} [delete]
func swaggerV2AdminDingTalkPerfReviewsIDDelete() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/dingtalk/perf-histories
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/perf-histories [get]
func swaggerV2AdminDingTalkPerfHistoriesGet() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/users
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users [get]
func swaggerV2AdminUsersGet78() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/users
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users [post]
func swaggerV2AdminUsersPost79() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/users/by-openid/{openid}
// @Security AdminToken
// @Param openid path string true "openid"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/by-openid/{openid} [get]
func swaggerV2AdminUsersByOpenidGetCompat() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/users/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/{id} [get]
func swaggerV2AdminUsersIdGet80() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/users/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/{id} [put]
func swaggerV2AdminUsersIdPut81() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/users/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/{id} [delete]
func swaggerV2AdminUsersIdDelete82() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/users
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users [delete]
func swaggerV2AdminUsersDelete83() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/users/{id}/status
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/{id}/status [patch]
func swaggerV2AdminUsersIdStatusPatch84() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/users/{id}/password
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/{id}/password [patch]
func swaggerV2AdminUsersIdPasswordPatch85() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/users/form-fields
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/form-fields [get]
func swaggerV2AdminUsersFormFieldsGet86() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/users/form-fields
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/form-fields [put]
func swaggerV2AdminUsersFormFieldsPut87() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/users/data
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/data [get]
func swaggerV2AdminUsersDataGet88() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/users/data/export
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/data/export [get]
func swaggerV2AdminUsersDataExportGet89() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/users/data/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/data/{id} [delete]
func swaggerV2AdminUsersDataIdDelete90() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/user-sessions
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/user-sessions [get]
func swaggerV2AdminUserSessionsGet91() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/user-sessions/{id}/force-offline
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/user-sessions/{id}/force-offline [post]
func swaggerV2AdminUserSessionsIdForceOfflinePost92() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/user-sessions/batch-force-offline
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/user-sessions/batch-force-offline [post]
func swaggerV2AdminUserSessionsBatchForceOfflinePost93() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/news
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news [get]
func swaggerV2AdminNewsGet94() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/news
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news [post]
func swaggerV2AdminNewsPost95() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/news/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news/{id} [get]
func swaggerV2AdminNewsIdGet96() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/news/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news/{id} [put]
func swaggerV2AdminNewsIdPut97() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/news/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news/{id} [delete]
func swaggerV2AdminNewsIdDelete98() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/news
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news [delete]
func swaggerV2AdminNewsDelete99() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/news/{id}/status
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news/{id}/status [patch]
func swaggerV2AdminNewsIdStatusPatch100() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/news/{id}/recommendation
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news/{id}/recommendation [patch]
func swaggerV2AdminNewsIdRecommendationPatch1001() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/news/{id}/sort
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news/{id}/sort [patch]
func swaggerV2AdminNewsIdSortPatch101() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/news/{id}/forms
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news/{id}/forms [patch]
func swaggerV2AdminNewsIdFormsPatch102() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/news/{id}/picture
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news/{id}/picture [patch]
func swaggerV2AdminNewsIdPicturePatch103() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/news/{id}/content
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news/{id}/content [patch]
func swaggerV2AdminNewsIdContentPatch104() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/enrollments
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments [get]
func swaggerV2AdminEnrollmentsGet105() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/enrollments
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments [post]
func swaggerV2AdminEnrollmentsPost106() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/enrollments/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id} [get]
func swaggerV2AdminEnrollmentsIdGet107() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/enrollments/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id} [put]
func swaggerV2AdminEnrollmentsIdPut108() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/enrollments/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id} [delete]
func swaggerV2AdminEnrollmentsIdDelete109() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/enrollments
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments [delete]
func swaggerV2AdminEnrollmentsDelete110() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/enrollments/{id}/status
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/status [patch]
func swaggerV2AdminEnrollmentsIdStatusPatch111() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/enrollments/{id}/sort
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/sort [patch]
func swaggerV2AdminEnrollmentsIdSortPatch112() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/enrollments/{id}/recommendation
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/recommendation [patch]
func swaggerV2AdminEnrollmentsIdRecommendationPatch113() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/enrollments/{id}/forms
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/forms [patch]
func swaggerV2AdminEnrollmentsIdFormsPatch114() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/enrollments/{id}/clear
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/clear [post]
func swaggerV2AdminEnrollmentsIdClearPost115() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/enrollments/{id}/joins
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/joins [get]
func swaggerV2AdminEnrollmentsIdJoinsGet116() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/enrollments/{id}/joins/{joinId}
// @Security AdminToken
// @Param id path int true "id"
// @Param joinId path int true "joinId"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/joins/{joinId} [delete]
func swaggerV2AdminEnrollmentsIdJoinsJoinIdDelete117() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/enrollments/{id}/joins
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/joins [delete]
func swaggerV2AdminEnrollmentsIdJoinsDelete118() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/enrollments/{id}/users
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/users [get]
func swaggerV2AdminEnrollmentsIdUsersGet119() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/enrollments/{id}/stats
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/stats [get]
func swaggerV2AdminEnrollmentsIdStatsGet120() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/enrollments/{id}/users/{userId}
// @Security AdminToken
// @Param id path int true "id"
// @Param userId path int true "userId"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/users/{userId} [delete]
func swaggerV2AdminEnrollmentsIdUsersUserIdDelete121() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/enrollments/{id}/users
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/users [delete]
func swaggerV2AdminEnrollmentsIdUsersDelete122() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/enrollments/{id}/users/{userId}/forms
// @Security AdminToken
// @Param id path int true "id"
// @Param userId path int true "userId"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/users/{userId}/forms [put]
func swaggerV2AdminEnrollmentsIdUsersUserIdFormsPut123() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/enrollments/{id}/export
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/export [get]
func swaggerV2AdminEnrollmentsIdExportGet124() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/enrollments/{id}/export
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/export [post]
func swaggerV2AdminEnrollmentsIdExportPost125() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/enrollments/{id}/export
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/export [delete]
func swaggerV2AdminEnrollmentsIdExportDelete126() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/events
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events [get]
func swaggerV2AdminEventsGet127() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/events
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events [post]
func swaggerV2AdminEventsPost128() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/events/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id} [get]
func swaggerV2AdminEventsIdGet129() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/events/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id} [put]
func swaggerV2AdminEventsIdPut130() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/events/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id} [delete]
func swaggerV2AdminEventsIdDelete131() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/events
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events [delete]
func swaggerV2AdminEventsDelete132() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/events/{id}/status
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/status [patch]
func swaggerV2AdminEventsIdStatusPatch133() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/events/{id}/recommendation
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/recommendation [patch]
func swaggerV2AdminEventsIdRecommendationPatch134() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/events/{id}/top
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/top [patch]
func swaggerV2AdminEventsIdTopPatch135() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/events/{id}/participants
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/participants [get]
func swaggerV2AdminEventsIdParticipantsGet136() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/events/{id}/participants/{participantId}
// @Security AdminToken
// @Param id path int true "id"
// @Param participantId path int true "participantId"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/participants/{participantId} [put]
func swaggerV2AdminEventsIdParticipantsParticipantIdPut137() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/events/{id}/participants/{participantId}
// @Security AdminToken
// @Param id path int true "id"
// @Param participantId path int true "participantId"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/participants/{participantId} [delete]
func swaggerV2AdminEventsIdParticipantsParticipantIdDelete138() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/events/{id}/participants
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/participants [delete]
func swaggerV2AdminEventsIdParticipantsDelete139() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/events/{id}/dynamics
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/dynamics [get]
func swaggerV2AdminEventsIdDynamicsGet140() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/events/{id}/dynamics
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/dynamics [post]
func swaggerV2AdminEventsIdDynamicsPost141() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/events/{id}/dynamics/{dynamicId}
// @Security AdminToken
// @Param id path int true "id"
// @Param dynamicId path int true "dynamicId"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/dynamics/{dynamicId} [put]
func swaggerV2AdminEventsIdDynamicsDynamicIdPut142() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/events/{id}/dynamics/{dynamicId}
// @Security AdminToken
// @Param id path int true "id"
// @Param dynamicId path int true "dynamicId"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/dynamics/{dynamicId} [delete]
func swaggerV2AdminEventsIdDynamicsDynamicIdDelete143() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/events/{id}/dynamics
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/dynamics [delete]
func swaggerV2AdminEventsIdDynamicsDelete144() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/events/{id}/scores
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/scores [get]
func swaggerV2AdminEventsIdScoresGet145() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/events/{id}/scores
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/scores [post]
func swaggerV2AdminEventsIdScoresPost1451() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/events/{id}/scores/{scoreId}
// @Security AdminToken
// @Param id path int true "id"
// @Param scoreId path int true "scoreId"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/scores/{scoreId} [put]
func swaggerV2AdminEventsIdScoresScoreIdPut146() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/event-dept-users
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/event-dept-users [get]
func swaggerV2AdminEventDeptUsersGet147() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/dict/types
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dict/types [get]
func swaggerV2AdminDictTypesGet148() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/dict/items
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dict/items [get]
func swaggerV2AdminDictItemsGet149() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/dict/items
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dict/items [post]
func swaggerV2AdminDictItemsPost150() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/dict/items/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dict/items/{id} [put]
func swaggerV2AdminDictItemsIdPut151() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/dict/items/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dict/items/{id} [delete]
func swaggerV2AdminDictItemsIdDelete152() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/dict/types/{typeCode}/items
// @Security AdminToken
// @Param typeCode path string true "typeCode"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dict/types/{typeCode}/items [delete]
func swaggerV2AdminDictTypesTypeCodeItemsDelete153() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/dict/types/{typeCode}
// @Security AdminToken
// @Param typeCode path string true "typeCode"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dict/types/{typeCode} [patch]
func swaggerV2AdminDictTypesTypeCodePatch154() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/departments/tree
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/departments/tree [get]
func swaggerV2AdminDepartmentsTreeGet155() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/departments
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/departments [post]
func swaggerV2AdminDepartmentsPost156() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/departments/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/departments/{id} [put]
func swaggerV2AdminDepartmentsIdPut157() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/departments/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/departments/{id} [delete]
func swaggerV2AdminDepartmentsIdDelete158() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/positions
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/positions [get]
func swaggerV2AdminPositionsGet() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/positions
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/positions [post]
func swaggerV2AdminPositionsPost() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/positions/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/positions/{id} [put]
func swaggerV2AdminPositionsIdPut() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/positions/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/positions/{id} [delete]
func swaggerV2AdminPositionsIdDelete() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/roles
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/roles [get]
func swaggerV2AdminRolesGet159() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/roles
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/roles [post]
func swaggerV2AdminRolesPost160() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/roles/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/roles/{id} [put]
func swaggerV2AdminRolesIdPut161() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/roles/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/roles/{id} [delete]
func swaggerV2AdminRolesIdDelete162() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/roles
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/roles [delete]
func swaggerV2AdminRolesDelete163() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/roles/application-permissions
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/roles/application-permissions [get]
func swaggerV2AdminRolesApplicationPermissionsGet164() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/permissions/tree
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/permissions/tree [get]
func swaggerV2AdminPermissionsTreeGet165() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/permissions
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/permissions [get]
func swaggerV2AdminPermissionsGet166() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/permissions
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/permissions [post]
func swaggerV2AdminPermissionsPost167() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/permissions/{key}
// @Security AdminToken
// @Param key path string true "权限编码"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/permissions/{key} [put]
func swaggerV2AdminPermissionsKeyPut168() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/permissions/{key}
// @Security AdminToken
// @Param key path string true "权限编码"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/permissions/{key} [delete]
func swaggerV2AdminPermissionsKeyDelete169() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/me/menus
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/me/menus [get]
func swaggerV2AdminMeMenusGet169() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/me/perms
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/me/perms [get]
func swaggerV2AdminMePermsGet170() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/me/password
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/me/password [patch]
func swaggerV2AdminMePasswordPatch1701() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/survey-types
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-types [get]
func swaggerV2AdminSurveyTypesGet171() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/survey-schema/parse
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-schema/parse [post]
func swaggerV2AdminSurveySchemaParsePost172() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/survey-expressions/evaluate
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-expressions/evaluate [post]
func swaggerV2AdminSurveyExpressionsEvaluatePost173() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/survey-report/enroll
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-report/enroll [get]
func swaggerV2AdminSurveyReportEnrollGet174() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/survey-report/enroll/export
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-report/enroll/export [get]
func swaggerV2AdminSurveyReportEnrollExportGet175() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/survey-report/event
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-report/event [get]
func swaggerV2AdminSurveyReportEventGet176() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/survey-report/event/export
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-report/event/export [get]
func swaggerV2AdminSurveyReportEventExportGet177() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/survey-report/survey
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-report/survey [get]
func swaggerV2AdminSurveyReportSurveyGet178() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/survey-report/survey/export
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-report/survey/export [get]
func swaggerV2AdminSurveyReportSurveyExportGet179() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/surveys
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys [get]
func swaggerV2AdminSurveysGet180() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/surveys
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys [post]
func swaggerV2AdminSurveysPost181() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/surveys/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id} [get]
func swaggerV2AdminSurveysIdGet182() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/surveys/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id} [put]
func swaggerV2AdminSurveysIdPut183() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/surveys/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id} [delete]
func swaggerV2AdminSurveysIdDelete184() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/surveys/{id}/status
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id}/status [patch]
func swaggerV2AdminSurveysIdStatusPatch185() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/surveys/{id}/copy
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id}/copy [post]
func swaggerV2AdminSurveysIdCopyPost186() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/surveys/{id}/statistics
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id}/statistics [get]
func swaggerV2AdminSurveysIdStatisticsGet187() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/surveys/{id}/responses
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id}/responses [get]
func swaggerV2AdminSurveysIdResponsesGet188() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/surveys/{id}/responses/export
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id}/responses/export [get]
func swaggerV2AdminSurveysIdResponsesExportGet189() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/survey-responses/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-responses/{id} [get]
func swaggerV2AdminSurveyResponsesIdGet190() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/survey-responses/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-responses/{id} [delete]
func swaggerV2AdminSurveyResponsesIdDelete191() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/survey-responses
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-responses [delete]
func swaggerV2AdminSurveyResponsesDelete192() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/surveys/{id}/channels
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id}/channels [get]
func swaggerV2AdminSurveysIdChannelsGet193() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/surveys/{id}/channels
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id}/channels [post]
func swaggerV2AdminSurveysIdChannelsPost194() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/survey-channels/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-channels/{id} [delete]
func swaggerV2AdminSurveyChannelsIdDelete195() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/survey-resources
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-resources [post]
func swaggerV2AdminSurveyResourcesPost196() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/surveys/{id}/resources
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id}/resources [get]
func swaggerV2AdminSurveysIdResourcesGet197() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/survey-resources/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-resources/{id} [delete]
func swaggerV2AdminSurveyResourcesIdDelete198() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/survey-question-bank
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-question-bank [get]
func swaggerV2AdminSurveyQuestionBankGet199() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/survey-question-bank
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-question-bank [post]
func swaggerV2AdminSurveyQuestionBankPost200() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/survey-question-bank/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-question-bank/{id} [put]
func swaggerV2AdminSurveyQuestionBankIdPut201() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/survey-question-bank/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-question-bank/{id} [delete]
func swaggerV2AdminSurveyQuestionBankIdDelete202() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/survey-question-bank/categories
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-question-bank/categories [get]
func swaggerV2AdminSurveyQuestionBankCategoriesGet203() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/survey-notifications
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-notifications [get]
func swaggerV2AdminSurveyNotificationsGet204() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/survey-notifications/{id}/read
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-notifications/{id}/read [patch]
func swaggerV2AdminSurveyNotificationsIdReadPatch205() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/survey-notifications/unread-count
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-notifications/unread-count [get]
func swaggerV2AdminSurveyNotificationsUnreadCountGet206() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/survey-template-presets
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-template-presets [get]
func swaggerV2AdminSurveyTemplatePresetsGet207() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/survey-template-presets
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-template-presets [put]
func swaggerV2AdminSurveyTemplatePresetsPut208() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/exams
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams [get]
func swaggerV2AdminExamsGet209() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/exams
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams [post]
func swaggerV2AdminExamsPost210() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/exams/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id} [get]
func swaggerV2AdminExamsIdGet211() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/exams/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id} [put]
func swaggerV2AdminExamsIdPut212() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/exams/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id} [delete]
func swaggerV2AdminExamsIdDelete213() {}

// @Tags API v2-后台管理
// @Summary 变更 /api/v2/admin/exams/{id}/status
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id}/status [patch]
func swaggerV2AdminExamsIdStatusPatch214() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/exams/{id}/records
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id}/records [get]
func swaggerV2AdminExamsIdRecordsGet215() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/exams/{id}/records/{recordId}
// @Security AdminToken
// @Param id path int true "id"
// @Param recordId path int true "recordId"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id}/records/{recordId} [get]
func swaggerV2AdminExamsIdRecordsRecordIdGet216() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/exams/{id}/records/{recordId}
// @Security AdminToken
// @Param id path int true "id"
// @Param recordId path int true "recordId"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id}/records/{recordId} [delete]
func swaggerV2AdminExamsIdRecordsRecordIdDelete217() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/exams/{id}/records
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id}/records [delete]
func swaggerV2AdminExamsIdRecordsDelete218() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/exams/{id}/statistics
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id}/statistics [get]
func swaggerV2AdminExamsIdStatisticsGet219() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/exam-resources
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exam-resources [post]
func swaggerV2AdminExamResourcesPost220() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/exams/{id}/resources
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id}/resources [get]
func swaggerV2AdminExamsIdResourcesGet221() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/exam-resources/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exam-resources/{id} [delete]
func swaggerV2AdminExamResourcesIdDelete222() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/exam-question-bank
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exam-question-bank [get]
func swaggerV2AdminExamQuestionBankGet223() {}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/exam-question-bank
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exam-question-bank [post]
func swaggerV2AdminExamQuestionBankPost224() {}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/exam-question-bank/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exam-question-bank/{id} [put]
func swaggerV2AdminExamQuestionBankIdPut225() {}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/exam-question-bank/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exam-question-bank/{id} [delete]
func swaggerV2AdminExamQuestionBankIdDelete226() {}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/exam-question-bank/categories
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exam-question-bank/categories [get]
func swaggerV2AdminExamQuestionBankCategoriesGet227() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询工作流定义列表
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-definitions [get]
func swaggerV2AdminWorkflowDefinitionsGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 新建工作流定义
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-definitions [post]
func swaggerV2AdminWorkflowDefinitionsPost() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询工作流定义详情
// @Security AdminToken
// @Param id path int true "流程定义 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-definitions/{id} [get]
func swaggerV2AdminWorkflowDefinitionsIDGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 更新工作流定义
// @Security AdminToken
// @Param id path int true "流程定义 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-definitions/{id} [put]
func swaggerV2AdminWorkflowDefinitionsIDPut() {}

// @Tags API v2-后台管理-工作流
// @Summary 删除工作流定义
// @Security AdminToken
// @Param id path int true "流程定义 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-definitions/{id} [delete]
func swaggerV2AdminWorkflowDefinitionsIDDelete() {}

// @Tags API v2-后台管理-工作流
// @Summary 校验工作流定义
// @Security AdminToken
// @Param id path int true "流程定义 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-definitions/{id}/validate [post]
func swaggerV2AdminWorkflowDefinitionsIDValidatePost() {}

// @Tags API v2-后台管理-工作流
// @Summary 发布工作流定义
// @Security AdminToken
// @Param id path int true "流程定义 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-definitions/{id}/publish [post]
func swaggerV2AdminWorkflowDefinitionsIDPublishPost() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询工作流定义版本
// @Security AdminToken
// @Param id path int true "流程定义 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-definitions/{id}/versions [get]
func swaggerV2AdminWorkflowDefinitionsIDVersionsGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询组织审批身份列表
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-org-approver-identities [get]
func swaggerV2AdminWorkflowOrgApproverIdentitiesGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询组织审批身份人员配置
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-org-approver-assignments [get]
func swaggerV2AdminWorkflowOrgApproverAssignmentsGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 保存组织审批身份人员配置
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-org-approver-assignments [put]
func swaggerV2AdminWorkflowOrgApproverAssignmentsPut() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询工作流实例列表
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-instances [get]
func swaggerV2AdminWorkflowInstancesGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 启动工作流实例
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-instances [post]
func swaggerV2AdminWorkflowInstancesPost() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询工作流实例详情
// @Security AdminToken
// @Param id path string true "流程实例 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-instances/{id} [get]
func swaggerV2AdminWorkflowInstancesIDGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 取消运行中的工作流实例
// @Security AdminToken
// @Param id path string true "流程实例 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-instances/{id}/cancel [post]
func swaggerV2AdminWorkflowInstancesIDCancelPost() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询工作流任务列表
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-tasks [get]
func swaggerV2AdminWorkflowTasksGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 完成工作流任务
// @Security AdminToken
// @Param id path string true "流程任务 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-tasks/{id}/complete [post]
func swaggerV2AdminWorkflowTasksIDCompletePost() {}

// @Tags API v2-客户端-OA流程
// @Summary 查询已发布的 OA 流程
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/workflows/definitions [get]
func swaggerV2WorkflowsDefinitionsGet() {}

// @Tags API v2-客户端-OA流程
// @Summary 查询 OA 流程和表单 Schema
// @Security ClientToken
// @Param id path int true "流程定义 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/workflows/definitions/{id} [get]
func swaggerV2WorkflowsDefinitionsIDGet() {}

// @Tags API v2-客户端-OA流程
// @Summary 发起 OA 流程
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/workflows/instances [post]
func swaggerV2WorkflowsInstancesPost() {}

// @Tags API v2-客户端-OA流程
// @Summary 查询我的 OA 流程申请
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/workflows/instances [get]
func swaggerV2WorkflowsInstancesGet() {}

// @Tags API v2-客户端-OA流程
// @Summary 查询我发起或参与的 OA 流程详情
// @Security ClientToken
// @Param id path string true "流程实例 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/workflows/instances/{id} [get]
func swaggerV2WorkflowsInstancesIDGet() {}

// @Tags API v2-客户端-OA流程
// @Summary 撤回未处理的 OA 流程申请
// @Security ClientToken
// @Param id path string true "流程实例 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/workflows/instances/{id}/withdraw [post]
func swaggerV2WorkflowsInstancesIDWithdrawPost() {}

// @Tags API v2-客户端-OA流程
// @Summary 查询我的 OA 流程任务
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/workflows/tasks [get]
func swaggerV2WorkflowsTasksGet() {}

// @Tags API v2-客户端-OA流程
// @Summary 审批或拒绝我的 OA 流程任务
// @Security ClientToken
// @Param id path string true "流程任务 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/workflows/tasks/{id}/complete [post]
func swaggerV2WorkflowsTasksIDCompletePost() {}
