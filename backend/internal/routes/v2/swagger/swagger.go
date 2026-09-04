package swagger

import (
	scheduledtaskapp "wecheckin/backend/internal/modules/scheduledtask/application"
	workflowservice "wecheckin/backend/internal/service/admin/workflow"
	"wecheckin/backend/pkg/response"
)

var _ response.Resp
var _ scheduledtaskapp.CreateTaskRequest
var _ workflowservice.PublishRequest

// @Tags API v2-公开接口-首页
// @Summary 查询 /api/v2/home
// @Param user_id query string false "用户 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/home [get]
func swaggerV2HomeGet0() {}

// @Tags API v2-公开接口-首页
// @Summary 查询 /api/v2/home/setup
// @Param key query string true "设置键名"
// @Success 200 {object} response.Resp
// @Router /api/v2/home/setup [get]
func swaggerV2HomeSetupGet1() {}

// @Tags API v2-公开接口-表单
// @Summary 查询 /api/v2/user-form-fields
// @Success 200 {object} response.Resp
// @Router /api/v2/user-form-fields [get]
func swaggerV2UserFormFieldsGet2() {}

// @Tags API v2-公开接口-认证
// @Summary 提交 /api/v2/auth/login
// @Accept application/x-www-form-urlencoded
// @Param user_id formData string true "用户ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/auth/login [post]
func swaggerV2AuthLoginPost3() {}

// @Tags API v2-公开接口-认证
// @Summary 提交 /api/v2/auth/password-login
// @Accept application/x-www-form-urlencoded
// @Param name formData string true "用户名/手机号"
// @Param pwd formData string true "密码"
// @Success 200 {object} response.Resp
// @Router /api/v2/auth/password-login [post]
func swaggerV2AuthPasswordLoginPost4() {}

// @Tags API v2-公开接口-认证
// @Summary 提交 /api/v2/auth/register
// @Accept application/x-www-form-urlencoded
// @Param user_id formData string true "用户ID"
// @Param name formData string true "姓名"
// @Param mobile formData string true "手机号"
// @Param pic formData string false "头像"
// @Param forms formData string false "表单数据（JSON）"
// @Success 200 {object} response.Resp
// @Router /api/v2/auth/register [post]
func swaggerV2AuthRegisterPost5() {}

// @Tags API v2-公开接口-地理编码
// @Summary 查询 /api/v2/geo/reverse
// @Param lat query string true "纬度"
// @Param lng query string true "经度"
// @Success 200 {object} response.Resp
// @Router /api/v2/geo/reverse [get]
func swaggerV2GeoReverseGet6() {}

// @Tags API v2-公开接口-字典
// @Summary 查询 /api/v2/dict/types
// @Success 200 {object} response.Resp
// @Router /api/v2/dict/types [get]
func swaggerV2DictTypesGet7() {}

// @Tags API v2-公开接口-字典
// @Summary 查询 /api/v2/dict/items
// @Param typeCode query string true "类型编码"
// @Success 200 {object} response.Resp
// @Router /api/v2/dict/items [get]
func swaggerV2DictItemsGet8() {}

// @Tags API v2-公开接口-赛事活动
// @Summary 查询 /api/v2/events
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param user_id query string false "用户ID"
// @Param keyword query string false "搜索关键词"
// @Param type query string false "活动类型"
// @Success 200 {object} response.Resp
// @Router /api/v2/events [get]
func swaggerV2EventsGet9() {}

// @Tags API v2-公开接口-赛事活动
// @Summary 查询 /api/v2/events/{id}
// @Param id path int true "id"
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/events/{id} [get]
func swaggerV2EventsIdGet10() {}

// @Tags API v2-公开接口-问卷
// @Summary 查询 /api/v2/surveys
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param keyword query string false "关键词"
// @Param category query string false "分类"
// @Param deviceId query string false "设备标识"
// @Success 200 {object} response.Resp
// @Router /api/v2/surveys [get]
func swaggerV2SurveysGet11() {}

// @Tags API v2-公开接口-问卷
// @Summary 查询 /api/v2/surveys/{id}
// @Param id path int true "id"
// @Param session query string false "会话标识"
// @Success 200 {object} response.Resp
// @Router /api/v2/surveys/{id} [get]
func swaggerV2SurveysIdGet12() {}

// @Tags API v2-公开接口-问卷
// @Summary 提交 /api/v2/surveys/{id}/responses
// @Accept application/json
// @Param id path int true "id"
// @Param body body PublicSurveySubmissionRequest true "问卷答案与会话信息"
// @Success 200 {object} response.Resp
// @Router /api/v2/surveys/{id}/responses [post]
func swaggerV2SurveysIdResponsesPost13() {}

// @Tags API v2-公开接口-问卷
// @Summary 提交 /api/v2/survey/apply
// @Accept application/json
// @Param body body PublicSurveyApplyRequest true "表单 Schema 与当前答案"
// @Success 200 {object} response.Resp
// @Router /api/v2/survey/apply [post]
func swaggerV2SurveyApplyPost14() {}

// @Tags API v2-公开接口-问卷
// @Summary 提交 /api/v2/survey/validate
// @Accept application/json
// @Param body body PublicSurveyValidateRequest true "问卷或表单校验数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/survey/validate [post]
func swaggerV2SurveyValidatePost15() {}

// @Tags API v2-公开接口-考试
// @Summary 查询 /api/v2/exams
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param keyword query string false "关键词"
// @Param deviceId query string false "设备标识"
// @Success 200 {object} response.Resp
// @Router /api/v2/exams [get]
func swaggerV2ExamsGet16() {}

// @Tags API v2-公开接口-考试
// @Summary 查询 /api/v2/exams/{id}
// @Param id path int true "id"
// @Param session query string false "会话标识"
// @Success 200 {object} response.Resp
// @Router /api/v2/exams/{id} [get]
func swaggerV2ExamsIdGet17() {}

// @Tags API v2-公开接口-考试
// @Summary 提交 /api/v2/exams/{id}/submissions
// @Accept application/json
// @Param id path int true "id"
// @Param body body PublicExamSubmissionRequest true "考试答案与会话信息"
// @Success 200 {object} response.Resp
// @Router /api/v2/exams/{id}/submissions [post]
func swaggerV2ExamsIdSubmissionsPost18() {}

// @Tags API v2-公开接口-考试
// @Summary 提交 /api/v2/exams/{id}/validation
// @Accept application/json
// @Param id path int true "id"
// @Param body body PublicExamValidationRequest true "待校验的考试答案"
// @Success 200 {object} response.Resp
// @Router /api/v2/exams/{id}/validation [post]
func swaggerV2ExamsIdValidationPost19() {}

// @Tags API v2-公开接口-考试
// @Summary 查询 /api/v2/exam-results
// @Param session query string false "会话标识"
// @Success 200 {object} response.Resp
// @Router /api/v2/exam-results [get]
func swaggerV2ExamResultsGet20() {}

// @Tags API v2-客户端-账户
// @Summary 查询 /api/v2/me/bootstrap
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/bootstrap [get]
func swaggerV2MeBootstrapGet201() {}

// @Tags API v2-客户端-账户
// @Summary 查询 /api/v2/me
// @Security ClientToken
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/me [get]
func swaggerV2MeGet21() {}

// @Tags API v2-客户端-账户
// @Summary 更新 /api/v2/me
// @Security ClientToken
// @Accept application/x-www-form-urlencoded
// @Param name formData string false "姓名"
// @Param mobile formData string false "手机号"
// @Param pic formData string false "头像"
// @Param user_id formData string false "用户ID"
// @Param forms formData string false "表单数据（JSON）"
// @Success 200 {object} response.Resp
// @Router /api/v2/me [put]
func swaggerV2MePut22() {}

// @Tags API v2-客户端-账户
// @Summary 提交 /api/v2/me/phone
// @Security ClientToken
// @Accept application/x-www-form-urlencoded
// @Param cloud_id formData string true "云ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/me/phone [post]
func swaggerV2MePhonePost23() {}

// @Tags API v2-客户端-账户
// @Summary 提交 /api/v2/me/logout
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/logout [post]
func swaggerV2MeLogoutPost24() {}

// @Tags API v2-客户端-收藏
// @Summary 查询 /api/v2/me/favorites
// @Security ClientToken
// @Param typ query string false "类型"
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/me/favorites [get]
func swaggerV2MeFavoritesGet25() {}

// @Tags API v2-客户端-收藏
// @Summary 提交 /api/v2/me/favorites
// @Security ClientToken
// @Accept application/x-www-form-urlencoded
// @Param title formData string true "标题"
// @Param oid formData string true "对象ID"
// @Param typ formData string true "类型"
// @Param path formData string false "路径"
// @Param user_id formData string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/me/favorites [post]
func swaggerV2MeFavoritesPost26() {}

// @Tags API v2-客户端-收藏
// @Summary 删除 /api/v2/me/favorites/{oid}
// @Security ClientToken
// @Accept application/x-www-form-urlencoded
// @Param oid path string true "oid"
// @Param user_id formData string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/me/favorites/{oid} [delete]
func swaggerV2MeFavoritesOidDelete27() {}

// @Tags API v2-客户端-收藏
// @Summary 查询 /api/v2/me/favorites/check
// @Security ClientToken
// @Param oid query string true "对象ID"
// @Param typ query string true "类型"
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/me/favorites/check [get]
func swaggerV2MeFavoritesCheckGet28() {}

// @Tags API v2-客户端-报名
// @Summary 查询 /api/v2/me/enrollments
// @Security ClientToken
// @Param user_id query string false "用户ID"
// @Param id query string false "ID"
// @Param enrollId query string false "报名项目 ID"
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
// @Success 200 {object} response.Resp
// @Router /api/v2/me/enrollments [get]
func swaggerV2MeEnrollmentsGet29() {}

// @Tags API v2-客户端-报名
// @Summary 查询 /api/v2/me/enrollment-users
// @Security ClientToken
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/me/enrollment-users [get]
func swaggerV2MeEnrollmentUsersGet30() {}

// @Tags API v2-客户端-报名
// @Summary 查询 /api/v2/me/enrollment-records
// @Security ClientToken
// @Param user_id query string false "用户ID"
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
// @Success 200 {object} response.Resp
// @Router /api/v2/me/enrollment-records [get]
func swaggerV2MeEnrollmentRecordsGet31() {}

// @Tags API v2-客户端-报名
// @Summary 查询 /api/v2/me/enrollment-calendar
// @Security ClientToken
// @Param user_id query string false "用户ID"
// @Param month query string false "年月 (2026-06)"
// @Success 200 {object} response.Resp
// @Router /api/v2/me/enrollment-calendar [get]
func swaggerV2MeEnrollmentCalendarGet32() {}

// @Tags API v2-客户端-报名
// @Summary 查询 /api/v2/me/enrollment-day-records
// @Security ClientToken
// @Param user_id query string false "用户ID"
// @Param day query string false "日期 (2026-06-01)"
// @Success 200 {object} response.Resp
// @Router /api/v2/me/enrollment-day-records [get]
func swaggerV2MeEnrollmentDayRecordsGet33() {}

// @Tags API v2-客户端-赛事活动
// @Summary 查询 /api/v2/me/events
// @Security ClientToken
// @Param user_id query string true "用户ID"
// @Param type query string false "活动类型"
// @Param status query string false "活动状态"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Resp
// @Router /api/v2/me/events [get]
func swaggerV2MeEventsGet34() {}

// @Tags API v2-客户端-赛事活动
// @Summary 查询 /api/v2/me/event-roles
// @Security ClientToken
// @Param user_id query string true "用户ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/me/event-roles [get]
func swaggerV2MeEventRolesGet35() {}

// @Tags API v2-客户端-赛事活动
// @Summary 查询 /api/v2/me/managed-events
// @Security ClientToken
// @Param user_id query string true "用户ID"
// @Param type query string false "活动类型"
// @Param status query string false "活动状态"
// @Param keyword query string false "搜索关键词"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Resp
// @Router /api/v2/me/managed-events [get]
func swaggerV2MeManagedEventsGet36() {}

// @Tags API v2-客户端-问卷
// @Summary 查询 /api/v2/me/survey-responses
// @Security ClientToken
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/me/survey-responses [get]
func swaggerV2MeSurveyResponsesGet37() {}

// @Tags API v2-客户端-问卷
// @Summary 查询 /api/v2/me/survey-responses/{id}
// @Security ClientToken
// @Param id path int true "id"
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/me/survey-responses/{id} [get]
func swaggerV2MeSurveyResponsesIdGet38() {}

// @Tags API v2-客户端-考试
// @Summary 查询 /api/v2/me/exam-records
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/me/exam-records [get]
func swaggerV2MeExamRecordsGet39() {}

// @Tags API v2-客户端-通知公告
// @Summary 查询 /api/v2/news
// @Security ClientToken
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param keyword query string false "关键词"
// @Param user_id query string false "用户 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/news [get]
func swaggerV2NewsGet40() {}

// @Tags API v2-客户端-通知公告
// @Summary 查询 /api/v2/news/categories
// @Security ClientToken
// @Success 200 {object} response.Resp
// @Router /api/v2/news/categories [get]
func swaggerV2NewsCategoriesGet41() {}

// @Tags API v2-客户端-通知公告
// @Summary 查询 /api/v2/news/{id}
// @Security ClientToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/news/{id} [get]
func swaggerV2NewsIdGet42() {}

// @Tags API v2-客户端-报名
// @Summary 查询 /api/v2/enrollments
// @Security ClientToken
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param user_id query string false "用户 ID"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Resp
// @Router /api/v2/enrollments [get]
func swaggerV2EnrollmentsGet43() {}

// @Tags API v2-客户端-报名
// @Summary 查询 /api/v2/enrollments/{id}
// @Security ClientToken
// @Param id path int true "id"
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/enrollments/{id} [get]
func swaggerV2EnrollmentsIdGet44() {}

// @Tags API v2-客户端-报名
// @Summary 查询 /api/v2/enrollments/{id}/join-days
// @Security ClientToken
// @Param id path int true "id"
// @Param enroll_id query string true "打卡ID"
// @Param day query string true "日期"
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/enrollments/{id}/join-days [get]
func swaggerV2EnrollmentsIdJoinDaysGet45() {}

// @Tags API v2-客户端-报名
// @Summary 提交 /api/v2/enrollments/{id}/joins
// @Security ClientToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param day formData string true "日期"
// @Param user_id formData string false "用户ID"
// @Param forms formData string false "表单数据"
// @Param enrollId formData string false "报名项目 ID"
// @Param token formData string false "会话令牌"
// @Success 200 {object} response.Resp
// @Router /api/v2/enrollments/{id}/joins [post]
func swaggerV2EnrollmentsIdJoinsPost46() {}

// @Tags API v2-客户端-报名
// @Summary 提交 /api/v2/enrollments/{id}/submissions
// @Security ClientToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param user_id formData string true "用户ID"
// @Param forms formData string false "打卡表单数据JSON"
// @Param enrollId formData string false "报名项目 ID"
// @Param token formData string false "会话令牌"
// @Success 200 {object} response.Resp
// @Router /api/v2/enrollments/{id}/submissions [post]
func swaggerV2EnrollmentsIdSubmissionsPost47() {}

// @Tags API v2-客户端-赛事活动
// @Summary 提交 /api/v2/events/{id}/participants
// @Security ClientToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param user_id formData string true "用户ID"
// @Param forms formData string false "报名表单数据(JSON)"
// @Param token formData string false "会话令牌"
// @Success 200 {object} response.Resp
// @Router /api/v2/events/{id}/participants [post]
func swaggerV2EventsIdParticipantsPost48() {}

// @Tags API v2-客户端-赛事活动
// @Summary 查询 /api/v2/events/{id}/participants
// @Security ClientToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/events/{id}/participants [get]
func swaggerV2EventsIdParticipantsGet49() {}

// @Tags API v2-客户端-赛事活动
// @Summary 查询 /api/v2/events/{id}/dynamics
// @Security ClientToken
// @Param id path int true "id"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Resp
// @Router /api/v2/events/{id}/dynamics [get]
func swaggerV2EventsIdDynamicsGet50() {}

// @Tags API v2-客户端-赛事活动
// @Summary 提交 /api/v2/events/{id}/dynamics
// @Security ClientToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param user_id formData string true "用户ID"
// @Param title formData string false "动态标题"
// @Param content formData string false "动态内容"
// @Param images formData string false "图片列表(JSON)"
// @Param videos formData string false "视频列表(JSON)"
// @Success 200 {object} response.Resp
// @Router /api/v2/events/{id}/dynamics [post]
func swaggerV2EventsIdDynamicsPost51() {}

// @Tags API v2-客户端-赛事活动
// @Summary 查询 /api/v2/events/{id}/scores
// @Security ClientToken
// @Param id path int true "id"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Resp
// @Router /api/v2/events/{id}/scores [get]
func swaggerV2EventsIdScoresGet52() {}

// @Tags API v2-客户端-赛事活动
// @Summary 提交 /api/v2/events/{id}/scores
// @Security ClientToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param participant_id formData string true "参赛者ID"
// @Param score formData string true "评分"
// @Param judge_id formData string true "评委ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/events/{id}/scores [post]
func swaggerV2EventsIdScoresPost53() {}

// @Tags API v2-客户端-考试
// @Summary 提交 /api/v2/exams/{id}/start
// @Security ClientToken
// @Param id path int true "id"
// @Param deviceId query string false "设备标识"
// @Success 200 {object} response.Resp
// @Router /api/v2/exams/{id}/start [post]
func swaggerV2ExamsIdStartPost54() {}

// @Tags API v2-客户端-考试
// @Summary 查询 /api/v2/exam-records/{id}
// @Security ClientToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/exam-records/{id} [get]
func swaggerV2ExamRecordsIdGet55() {}

// @Tags API v2-客户端-考试
// @Summary 更新 /api/v2/exam-records/{id}/answers
// @Security ClientToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param answers formData string true "答案JSON"
// @Success 200 {object} response.Resp
// @Router /api/v2/exam-records/{id}/answers [put]
func swaggerV2ExamRecordsIdAnswersPut56() {}

// @Tags API v2-后台管理-认证
// @Summary 提交 /api/v2/admin/auth/login
// @Accept application/x-www-form-urlencoded
// @Param name formData string true "用户名"
// @Param password formData string true "密码"
// @Param pwd formData string false "密码（兼容参数）"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/auth/login [post]
func swaggerV2AdminAuthLoginPost57() {}

// @Tags API v2-后台管理-首页
// @Summary 查询 /api/v2/admin/home
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/home [get]
func swaggerV2AdminHomeGet58() {}

// @Tags API v2-后台管理-首页
// @Summary 删除 /api/v2/admin/home/recommendations
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/home/recommendations [delete]
func swaggerV2AdminHomeRecommendationsDelete59() {}

// @Tags API v2-后台管理-管理员
// @Summary 查询 /api/v2/admin/managers
// @Security AdminToken
// @Param keyword query string false "关键词"
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/managers [get]
func swaggerV2AdminManagersGet60() {}

// @Tags API v2-后台管理-管理员
// @Summary 提交 /api/v2/admin/managers
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param name formData string true "用户名"
// @Param password formData string true "密码"
// @Param desc formData string false "描述"
// @Param phone formData string false "手机号"
// @Param roleId formData string false "角色 ID"
// @Param deptIds formData string false "部门 ID 列表"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/managers [post]
func swaggerV2AdminManagersPost61() {}

// @Tags API v2-后台管理-管理员
// @Summary 查询 /api/v2/admin/managers/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/managers/{id} [get]
func swaggerV2AdminManagersIdGet62() {}

// @Tags API v2-后台管理-管理员
// @Summary 更新 /api/v2/admin/managers/{id}
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param name formData string false "用户名"
// @Param desc formData string false "描述"
// @Param phone formData string false "手机号"
// @Param pic formData string false "图片地址"
// @Param password formData string false "密码"
// @Param roleId formData string false "角色 ID"
// @Param deptIds formData string false "部门 ID 列表"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/managers/{id} [put]
func swaggerV2AdminManagersIdPut63() {}

// @Tags API v2-后台管理-管理员
// @Summary 删除 /api/v2/admin/managers/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/managers/{id} [delete]
func swaggerV2AdminManagersIdDelete64() {}

// @Tags API v2-后台管理-管理员
// @Summary 删除 /api/v2/admin/managers
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param ids formData string false "ID 列表，多个值用逗号分隔"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/managers [delete]
func swaggerV2AdminManagersDelete65() {}

// @Tags API v2-后台管理-管理员
// @Summary 变更 /api/v2/admin/managers/{id}/status
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param status formData string true "状态"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/managers/{id}/status [patch]
func swaggerV2AdminManagersIdStatusPatch66() {}

// @Tags API v2-后台管理-管理员
// @Summary 变更 /api/v2/admin/managers/{id}/password
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param password formData string true "新密码"
// @Param oldPassword formData string false "原密码"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/managers/{id}/password [patch]
func swaggerV2AdminManagersIdPasswordPatch67() {}

// @Tags API v2-后台管理-管理员
// @Summary 查询 /api/v2/admin/admin-sessions
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/admin-sessions [get]
func swaggerV2AdminAdminSessionsGet68() {}

// @Tags API v2-后台管理-管理员
// @Summary 提交 /api/v2/admin/admin-sessions/{id}/force-offline
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param token formData string false "会话令牌"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/admin-sessions/{id}/force-offline [post]
func swaggerV2AdminAdminSessionsIdForceOfflinePost69() {}

// @Tags API v2-后台管理-管理员
// @Summary 提交 /api/v2/admin/admin-sessions/batch-force-offline
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param ids formData string false "ID 列表，多个值用逗号分隔"
// @Param tokens formData string false "会话令牌列表"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/admin-sessions/batch-force-offline [post]
func swaggerV2AdminAdminSessionsBatchForceOfflinePost70() {}

// @Tags API v2-后台管理-认证
// @Summary 提交 /api/v2/admin/auth/logout
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/auth/logout [post]
func swaggerV2AdminAuthLogoutPost71() {}

// @Tags API v2-后台管理-日志
// @Summary 查询 /api/v2/admin/logs
// @Security AdminToken
// @Param search query string false "搜索关键词"
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/logs [get]
func swaggerV2AdminLogsGet72() {}

// @Tags API v2-后台管理-日志
// @Summary 删除 /api/v2/admin/logs
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param ids formData string false "ID 列表，多个值用逗号分隔"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/logs [delete]
func swaggerV2AdminLogsDelete73() {}

// @Tags API v2-后台管理-系统设置
// @Summary 更新 /api/v2/admin/settings
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param key formData string true "设置键名"
// @Param value formData string true "设置值"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/settings [put]
func swaggerV2AdminSettingsPut74() {}

// @Tags API v2-后台管理-系统设置
// @Summary 查询 /api/v2/admin/settings/content
// @Security AdminToken
// @Param key query string true "设置键名"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/settings/content [get]
func swaggerV2AdminSettingsContentGet() {}

// @Tags API v2-后台管理-系统设置
// @Summary 更新 /api/v2/admin/settings/content
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param key formData string true "设置键名"
// @Param value formData string true "设置值"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/settings/content [put]
func swaggerV2AdminSettingsContentPut75() {}

// @Tags API v2-后台管理-系统设置
// @Summary 查询 /api/v2/admin/settings/mini-qr
// @Security AdminToken
// @Param page query string false "页面路径"
// @Param scene query string false "场景值"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/settings/mini-qr [get]
func swaggerV2AdminSettingsMiniQrGet76() {}

// @Tags API v2-后台管理-系统设置
// @Summary 查询 /api/v2/admin/settings/debug-token
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/settings/debug-token [get]
func swaggerV2AdminSettingsDebugTokenGet77() {}

// @Tags API v2-后台管理-文件上传
// @Summary 上传后台图片或视频
// @Security AdminToken
// @Accept multipart/form-data
// @Param file formData file true "图片或视频文件（最大 20MB）"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/uploads [post]
func swaggerV2AdminUploadsPost() {}

// @Tags API v2-后台管理-钉钉集成
// @Summary 查询 /api/v2/admin/dingtalk/settings
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/settings [get]
func swaggerV2AdminDingTalkSettingsGet() {}

// @Tags API v2-后台管理-钉钉集成
// @Summary 更新 /api/v2/admin/dingtalk/settings
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param scope formData string false "保存范围"
// @Param tokenExpire formData string false "令牌有效期（秒）"
// @Param redisPrefix formData string false "Redis 键前缀"
// @Param singleLogin formData string false "是否限制单点登录"
// @Param selfBind formData string false "是否允许用户自助绑定"
// @Param appName formData string false "应用名称"
// @Param logoText formData string false "应用标识文字"
// @Param logoUrl formData string false "应用 Logo 地址"
// @Param appUrl formData string false "应用访问地址"
// @Param corpId formData string false "企业 ID"
// @Param appKey formData string false "钉钉应用 AppKey"
// @Param appSecret formData string false "钉钉应用 AppSecret"
// @Param agentId formData string false "钉钉应用 AgentId"
// @Param unifiedAppId formData string false "统一应用 ID"
// @Param notifyMode formData string false "通知方式"
// @Param robotCode formData string false "钉钉机器人编码"
// @Param notifyEnabled formData string false "是否启用通知"
// @Param corpConfigs formData string false "多企业配置（JSON）"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/settings [put]
func swaggerV2AdminDingTalkSettingsPut() {}

// @Tags API v2-后台管理-钉钉集成
// @Summary 提交 /api/v2/admin/dingtalk/settings/notification-test
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param corpId formData string false "企业 ID"
// @Param dingTalkUserId formData string false "钉钉用户 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/settings/notification-test [post]
func swaggerV2AdminDingTalkSettingsNotificationTestPost() {}

// @Tags API v2-后台管理-钉钉集成
// @Summary 查询 /api/v2/admin/dingtalk/user-bindings
// @Security AdminToken
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
// @Param corpId query string false "企业 ID"
// @Param keyword query string false "关键词"
// @Param enabled query string false "是否启用"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/user-bindings [get]
func swaggerV2AdminDingTalkUserBindingsGet() {}

// @Tags API v2-后台管理-钉钉集成
// @Summary 提交 /api/v2/admin/dingtalk/user-bindings
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id formData string false "ID"
// @Param corpId formData string false "企业 ID"
// @Param dingTalkUserId formData string false "钉钉用户 ID"
// @Param unionId formData string false "钉钉 UnionId"
// @Param userId formData string false "用户 ID"
// @Param enabled formData string false "是否启用"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/user-bindings [post]
func swaggerV2AdminDingTalkUserBindingsPost() {}

// @Tags API v2-后台管理-钉钉集成
// @Summary 更新 /api/v2/admin/dingtalk/user-bindings/{id}/status
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param enabled formData string false "是否启用"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/user-bindings/{id}/status [patch]
func swaggerV2AdminDingTalkUserBindingsIDStatusPatch() {}

// @Tags API v2-后台管理-钉钉集成
// @Summary 删除 /api/v2/admin/dingtalk/user-bindings/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/user-bindings/{id} [delete]
func swaggerV2AdminDingTalkUserBindingsIDDelete() {}

// @Tags API v2-后台管理-钉钉集成
// @Summary 查询 /api/v2/admin/dingtalk/perf-reviews
// @Security AdminToken
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/perf-reviews [get]
func swaggerV2AdminDingTalkPerfReviewsGet() {}

// @Tags API v2-后台管理-钉钉集成
// @Summary 批量删除 /api/v2/admin/dingtalk/perf-reviews
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param ids formData string false "ID 列表，多个值用逗号分隔"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/perf-reviews [delete]
func swaggerV2AdminDingTalkPerfReviewsDelete() {}

// @Tags API v2-后台管理-钉钉集成
// @Summary 查询 /api/v2/admin/dingtalk/perf-reviews/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/perf-reviews/{id} [get]
func swaggerV2AdminDingTalkPerfReviewsIDGet() {}

// @Tags API v2-后台管理-钉钉集成
// @Summary 删除 /api/v2/admin/dingtalk/perf-reviews/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/perf-reviews/{id} [delete]
func swaggerV2AdminDingTalkPerfReviewsIDDelete() {}

// @Tags API v2-后台管理-钉钉集成
// @Summary 查询 /api/v2/admin/dingtalk/perf-histories
// @Security AdminToken
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
// @Param keyword query string false "关键词"
// @Param reviewNo query string false "绩效单号"
// @Param byAccount query string false "操作账号"
// @Param action query string false "操作类型"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/perf-histories [get]
func swaggerV2AdminDingTalkPerfHistoriesGet() {}

// @Tags API v2-后台管理-钉钉集成
// @Summary 批量删除 /api/v2/admin/dingtalk/perf-histories
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param ids formData string false "ID 列表，多个值用逗号分隔"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/perf-histories [delete]
func swaggerV2AdminDingTalkPerfHistoriesDelete() {}

// @Tags API v2-后台管理-钉钉集成
// @Summary 删除 /api/v2/admin/dingtalk/perf-histories/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk/perf-histories/{id} [delete]
func swaggerV2AdminDingTalkPerfHistoriesIDDelete() {}

// @Tags API v2-后台管理-用户管理
// @Summary 查询 /api/v2/admin/users
// @Security AdminToken
// @Param page query string false "页码"
// @Param size query string false "每页数量"
// @Param keyword query string false "关键词"
// @Param pageSize query string false "每页数量"
// @Param sort query string false "排序值"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users [get]
func swaggerV2AdminUsersGet78() {}

// @Tags API v2-后台管理-用户管理
// @Summary 提交 /api/v2/admin/users
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param name formData string true "用户名"
// @Param mobile formData string false "手机号"
// @Param positionId formData string false "岗位ID"
// @Param managerUserId formData string false "直属上级用户ID"
// @Param pic formData string false "头像URL"
// @Param forms formData string false "扩展表单数据JSON"
// @Param deptIds formData string false "部门 ID 列表"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users [post]
func swaggerV2AdminUsersPost79() {}

// @Tags API v2-后台管理-用户管理
// @Summary 查询 /api/v2/admin/users/by-openid/{openid}
// @Security AdminToken
// @Param openid path string true "openid"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/by-openid/{openid} [get]
func swaggerV2AdminUsersByOpenidGetCompat() {}

// @Tags API v2-后台管理-用户管理
// @Summary 查询 /api/v2/admin/users/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/{id} [get]
func swaggerV2AdminUsersIdGet80() {}

// @Tags API v2-后台管理-用户管理
// @Summary 更新 /api/v2/admin/users/{id}
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param name formData string false "用户名"
// @Param mobile formData string false "手机号"
// @Param positionId formData string false "岗位ID"
// @Param managerUserId formData string false "直属上级用户ID"
// @Param pic formData string false "头像URL"
// @Param forms formData string false "扩展表单数据JSON"
// @Param deptIds formData string false "部门 ID 列表"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/{id} [put]
func swaggerV2AdminUsersIdPut81() {}

// @Tags API v2-后台管理-用户管理
// @Summary 删除 /api/v2/admin/users/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/{id} [delete]
func swaggerV2AdminUsersIdDelete82() {}

// @Tags API v2-后台管理-用户管理
// @Summary 删除 /api/v2/admin/users
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param ids formData string false "ID 列表，多个值用逗号分隔"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users [delete]
func swaggerV2AdminUsersDelete83() {}

// @Tags API v2-后台管理-用户管理
// @Summary 变更 /api/v2/admin/users/{id}/status
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param status formData string true "状态"
// @Param reason formData string false "原因"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/{id}/status [patch]
func swaggerV2AdminUsersIdStatusPatch84() {}

// @Tags API v2-后台管理-用户管理
// @Summary 变更 /api/v2/admin/users/{id}/password
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/{id}/password [patch]
func swaggerV2AdminUsersIdPasswordPatch85() {}

// @Tags API v2-后台管理-用户管理
// @Summary 查询 /api/v2/admin/users/form-fields
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/form-fields [get]
func swaggerV2AdminUsersFormFieldsGet86() {}

// @Tags API v2-后台管理-用户管理
// @Summary 更新 /api/v2/admin/users/form-fields
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param fields formData string true "字段JSON数组"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/form-fields [put]
func swaggerV2AdminUsersFormFieldsPut87() {}

// @Tags API v2-后台管理-用户管理
// @Summary 查询 /api/v2/admin/users/data
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/data [get]
func swaggerV2AdminUsersDataGet88() {}

// @Tags API v2-后台管理-用户管理
// @Summary 查询 /api/v2/admin/users/data/export
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/data/export [get]
func swaggerV2AdminUsersDataExportGet89() {}

// @Tags API v2-后台管理-用户管理
// @Summary 删除 /api/v2/admin/users/data/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/data/{id} [delete]
func swaggerV2AdminUsersDataIdDelete90() {}

// @Tags API v2-后台管理-用户管理
// @Summary 查询 /api/v2/admin/user-sessions
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/user-sessions [get]
func swaggerV2AdminUserSessionsGet91() {}

// @Tags API v2-后台管理-用户管理
// @Summary 提交 /api/v2/admin/user-sessions/{id}/force-offline
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param token formData string false "会话令牌"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/user-sessions/{id}/force-offline [post]
func swaggerV2AdminUserSessionsIdForceOfflinePost92() {}

// @Tags API v2-后台管理-用户管理
// @Summary 提交 /api/v2/admin/user-sessions/batch-force-offline
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param ids formData string false "ID 列表，多个值用逗号分隔"
// @Param tokens formData string false "会话令牌列表"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/user-sessions/batch-force-offline [post]
func swaggerV2AdminUserSessionsBatchForceOfflinePost93() {}

// @Tags API v2-后台管理-通知公告
// @Summary 查询 /api/v2/admin/news
// @Security AdminToken
// @Param page query string false "页码"
// @Param size query string false "每页数量"
// @Param keyword query string false "关键词"
// @Param pageSize query string false "每页数量"
// @Param sort query string false "排序值"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news [get]
func swaggerV2AdminNewsGet94() {}

// @Tags API v2-后台管理-通知公告
// @Summary 提交 /api/v2/admin/news
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param title formData string false "标题"
// @Param desc formData string false "描述"
// @Param cateId formData string false "分类 ID"
// @Param cateName formData string false "分类名称"
// @Param content formData string false "内容"
// @Param img formData string false "图片地址"
// @Param order formData string false "排序值"
// @Param sortOrder formData string false "排序值（兼容参数）"
// @Param deptId formData string false "部门 ID"
// @Param publishDeptIds formData string false "发布部门 ID 列表"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news [post]
func swaggerV2AdminNewsPost95() {}

// @Tags API v2-后台管理-通知公告
// @Summary 查询 /api/v2/admin/news/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news/{id} [get]
func swaggerV2AdminNewsIdGet96() {}

// @Tags API v2-后台管理-通知公告
// @Summary 更新 /api/v2/admin/news/{id}
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param title formData string false "标题"
// @Param desc formData string false "描述"
// @Param cateId formData string false "分类 ID"
// @Param cateName formData string false "分类名称"
// @Param content formData string false "内容"
// @Param img formData string false "图片地址"
// @Param order formData string false "排序值"
// @Param sortOrder formData string false "排序值（兼容参数）"
// @Param deptId formData string false "部门 ID"
// @Param publishDeptIds formData string false "发布部门 ID 列表"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news/{id} [put]
func swaggerV2AdminNewsIdPut97() {}

// @Tags API v2-后台管理-通知公告
// @Summary 删除 /api/v2/admin/news/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news/{id} [delete]
func swaggerV2AdminNewsIdDelete98() {}

// @Tags API v2-后台管理-通知公告
// @Summary 删除 /api/v2/admin/news
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param ids formData string false "ID 列表，多个值用逗号分隔"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news [delete]
func swaggerV2AdminNewsDelete99() {}

// @Tags API v2-后台管理-通知公告
// @Summary 变更 /api/v2/admin/news/{id}/status
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param status formData string true "状态"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news/{id}/status [patch]
func swaggerV2AdminNewsIdStatusPatch100() {}

// @Tags API v2-后台管理-通知公告
// @Summary 变更 /api/v2/admin/news/{id}/recommendation
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param vouch formData int true "推荐(1=推荐 0=取消)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news/{id}/recommendation [patch]
func swaggerV2AdminNewsIdRecommendationPatch1001() {}

// @Tags API v2-后台管理-通知公告
// @Summary 变更 /api/v2/admin/news/{id}/sort
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param sort formData string true "排序值"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news/{id}/sort [patch]
func swaggerV2AdminNewsIdSortPatch101() {}

// @Tags API v2-后台管理-通知公告
// @Summary 变更 /api/v2/admin/news/{id}/forms
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param forms formData string false "表单数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news/{id}/forms [patch]
func swaggerV2AdminNewsIdFormsPatch102() {}

// @Tags API v2-后台管理-通知公告
// @Summary 变更 /api/v2/admin/news/{id}/picture
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param pic formData string false "图片数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news/{id}/picture [patch]
func swaggerV2AdminNewsIdPicturePatch103() {}

// @Tags API v2-后台管理-通知公告
// @Summary 变更 /api/v2/admin/news/{id}/content
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param content formData string false "内容"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/news/{id}/content [patch]
func swaggerV2AdminNewsIdContentPatch104() {}

// @Tags API v2-后台管理-报名管理
// @Summary 查询 /api/v2/admin/enrollments
// @Security AdminToken
// @Param page query string false "页码"
// @Param size query string false "每页数量"
// @Param keyword query string false "关键词"
// @Param pageSize query string false "每页数量"
// @Param sort query string false "排序值"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments [get]
func swaggerV2AdminEnrollmentsGet105() {}

// @Tags API v2-后台管理-报名管理
// @Summary 提交 /api/v2/admin/enrollments
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param title formData string false "标题"
// @Param cateId formData string false "分类 ID"
// @Param cateName formData string false "分类名称"
// @Param startTime formData string false "开始时间"
// @Param endTime formData string false "结束时间"
// @Param sort formData string false "排序值"
// @Param cover formData string false "封面地址"
// @Param desc formData string false "描述"
// @Param joinForms formData string false "报名表单定义（JSON）"
// @Param enrollForms formData string false "打卡表单定义（JSON）"
// @Param allowRepeat formData string false "是否允许重复打卡"
// @Param dailyLimit formData string false "每日次数上限"
// @Param deptId formData string false "部门 ID"
// @Param publishDeptIds formData string false "发布部门 ID 列表"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments [post]
func swaggerV2AdminEnrollmentsPost106() {}

// @Tags API v2-后台管理-报名管理
// @Summary 查询 /api/v2/admin/enrollments/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id} [get]
func swaggerV2AdminEnrollmentsIdGet107() {}

// @Tags API v2-后台管理-报名管理
// @Summary 更新 /api/v2/admin/enrollments/{id}
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param title formData string false "标题"
// @Param cateId formData string false "分类ID"
// @Param cateName formData string false "分类名称"
// @Param startTime formData string false "开始时间"
// @Param endTime formData string false "结束时间"
// @Param sort formData string false "排序"
// @Param cover formData string false "封面图URL"
// @Param desc formData string false "描述"
// @Param allowRepeat formData string false "是否允许重复打卡"
// @Param dailyLimit formData string false "每日次数上限"
// @Param joinForms formData string false "报名表单定义（JSON）"
// @Param enrollForms formData string false "打卡表单定义（JSON）"
// @Param deptId formData string false "部门 ID"
// @Param publishDeptIds formData string false "发布部门 ID 列表"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id} [put]
func swaggerV2AdminEnrollmentsIdPut108() {}

// @Tags API v2-后台管理-报名管理
// @Summary 删除 /api/v2/admin/enrollments/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id} [delete]
func swaggerV2AdminEnrollmentsIdDelete109() {}

// @Tags API v2-后台管理-报名管理
// @Summary 删除 /api/v2/admin/enrollments
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param ids formData string false "ID 列表，多个值用逗号分隔"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments [delete]
func swaggerV2AdminEnrollmentsDelete110() {}

// @Tags API v2-后台管理-报名管理
// @Summary 变更 /api/v2/admin/enrollments/{id}/status
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param status formData string true "状态"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/status [patch]
func swaggerV2AdminEnrollmentsIdStatusPatch111() {}

// @Tags API v2-后台管理-报名管理
// @Summary 变更 /api/v2/admin/enrollments/{id}/sort
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param sort formData string true "排序值"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/sort [patch]
func swaggerV2AdminEnrollmentsIdSortPatch112() {}

// @Tags API v2-后台管理-报名管理
// @Summary 变更 /api/v2/admin/enrollments/{id}/recommendation
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param vouch formData string true "推荐值"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/recommendation [patch]
func swaggerV2AdminEnrollmentsIdRecommendationPatch113() {}

// @Tags API v2-后台管理-报名管理
// @Summary 变更 /api/v2/admin/enrollments/{id}/forms
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param forms formData string false "表单数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/forms [patch]
func swaggerV2AdminEnrollmentsIdFormsPatch114() {}

// @Tags API v2-后台管理-报名管理
// @Summary 提交 /api/v2/admin/enrollments/{id}/clear
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/clear [post]
func swaggerV2AdminEnrollmentsIdClearPost115() {}

// @Tags API v2-后台管理-报名管理
// @Summary 查询 /api/v2/admin/enrollments/{id}/joins
// @Security AdminToken
// @Param id path int true "id"
// @Param page query string false "页码"
// @Param size query string false "每页数量"
// @Param keyword query string false "关键词"
// @Param pageSize query string false "每页数量"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/joins [get]
func swaggerV2AdminEnrollmentsIdJoinsGet116() {}

// @Tags API v2-后台管理-报名管理
// @Summary 删除 /api/v2/admin/enrollments/{id}/joins/{joinId}
// @Security AdminToken
// @Param id path int true "id"
// @Param joinId path int true "joinId"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/joins/{joinId} [delete]
func swaggerV2AdminEnrollmentsIdJoinsJoinIdDelete117() {}

// @Tags API v2-后台管理-报名管理
// @Summary 删除 /api/v2/admin/enrollments/{id}/joins
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param ids formData string false "ID 列表，多个值用逗号分隔"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/joins [delete]
func swaggerV2AdminEnrollmentsIdJoinsDelete118() {}

// @Tags API v2-后台管理-报名管理
// @Summary 查询 /api/v2/admin/enrollments/{id}/users
// @Security AdminToken
// @Param id path int true "id"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/users [get]
func swaggerV2AdminEnrollmentsIdUsersGet119() {}

// @Tags API v2-后台管理-报名管理
// @Summary 查询 /api/v2/admin/enrollments/{id}/stats
// @Security AdminToken
// @Param id path int true "id"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/stats [get]
func swaggerV2AdminEnrollmentsIdStatsGet120() {}

// @Tags API v2-后台管理-报名管理
// @Summary 删除 /api/v2/admin/enrollments/{id}/users/{userId}
// @Security AdminToken
// @Param id path int true "id"
// @Param userId path int true "userId"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/users/{userId} [delete]
func swaggerV2AdminEnrollmentsIdUsersUserIdDelete121() {}

// @Tags API v2-后台管理-报名管理
// @Summary 删除 /api/v2/admin/enrollments/{id}/users
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param userIds formData string false "用户 ID 列表"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/users [delete]
func swaggerV2AdminEnrollmentsIdUsersDelete122() {}

// @Tags API v2-后台管理-报名管理
// @Summary 更新 /api/v2/admin/enrollments/{id}/users/{userId}/forms
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param userId path int true "userId"
// @Param forms formData string false "表单数据（JSON）"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/users/{userId}/forms [put]
func swaggerV2AdminEnrollmentsIdUsersUserIdFormsPut123() {}

// @Tags API v2-后台管理-报名管理
// @Summary 查询 /api/v2/admin/enrollments/{id}/export
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/export [get]
func swaggerV2AdminEnrollmentsIdExportGet124() {}

// @Tags API v2-后台管理-报名管理
// @Summary 提交 /api/v2/admin/enrollments/{id}/export
// @Security AdminToken
// @Param id path int true "id"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/export [post]
func swaggerV2AdminEnrollmentsIdExportPost125() {}

// @Tags API v2-后台管理-报名管理
// @Summary 删除 /api/v2/admin/enrollments/{id}/export
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/enrollments/{id}/export [delete]
func swaggerV2AdminEnrollmentsIdExportDelete126() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 查询 /api/v2/admin/events
// @Security AdminToken
// @Param keyword query string false "搜索关键词"
// @Param type query string false "活动类型"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param sort query string false "排序"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events [get]
func swaggerV2AdminEventsGet127() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 提交 /api/v2/admin/events
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param title formData string true "活动标题"
// @Param type formData int false "活动类型(1=活动 2=赛事)"
// @Param cateId formData string false "分类ID"
// @Param cateName formData string false "分类名称"
// @Param status formData int false "状态"
// @Param order formData int false "排序"
// @Param regStart formData int false "报名开始时间(时间戳)"
// @Param regEnd formData int false "报名结束时间(时间戳)"
// @Param eventStart formData int false "活动开始时间(时间戳)"
// @Param eventEnd formData int false "活动结束时间(时间戳)"
// @Param forms formData string false "报名表单(JSON)"
// @Param scoreFields formData string false "评分字段(JSON)"
// @Param qr formData string false "二维码URL"
// @Param obj formData string false "扩展对象(JSON)"
// @Param publishDeptIds formData string false "发布部门IDs(逗号分隔)"
// @Param deptId formData int false "所属部门ID"
// @Param organizers formData string false "组织者列表(JSON或逗号分隔)"
// @Param assistants formData string false "协助者列表(JSON或逗号分隔)"
// @Param referees formData string false "裁判列表(JSON或逗号分隔)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events [post]
func swaggerV2AdminEventsPost128() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 查询 /api/v2/admin/events/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id} [get]
func swaggerV2AdminEventsIdGet129() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 更新 /api/v2/admin/events/{id}
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param title formData string false "活动标题"
// @Param type formData int false "活动类型"
// @Param cateId formData string false "分类ID"
// @Param cateName formData string false "分类名称"
// @Param status formData int false "状态"
// @Param order formData int false "排序"
// @Param regStart formData int false "报名开始时间(时间戳)"
// @Param regEnd formData int false "报名结束时间(时间戳)"
// @Param eventStart formData int false "活动开始时间(时间戳)"
// @Param eventEnd formData int false "活动结束时间(时间戳)"
// @Param forms formData string false "报名表单(JSON)"
// @Param scoreFields formData string false "评分字段(JSON)"
// @Param qr formData string false "二维码URL"
// @Param obj formData string false "扩展对象(JSON)"
// @Param publishDeptIds formData string false "发布部门IDs(逗号分隔)"
// @Param deptId formData int false "所属部门ID"
// @Param organizers formData string false "组织者列表(JSON或逗号分隔)"
// @Param assistants formData string false "协助者列表(JSON或逗号分隔)"
// @Param referees formData string false "裁判列表(JSON或逗号分隔)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id} [put]
func swaggerV2AdminEventsIdPut130() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 删除 /api/v2/admin/events/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id} [delete]
func swaggerV2AdminEventsIdDelete131() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 删除 /api/v2/admin/events
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param ids formData string true "活动ID列表(逗号分隔)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events [delete]
func swaggerV2AdminEventsDelete132() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 变更 /api/v2/admin/events/{id}/status
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param status formData int true "状态(1=启用 0=禁用)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/status [patch]
func swaggerV2AdminEventsIdStatusPatch133() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 变更 /api/v2/admin/events/{id}/recommendation
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param vouch formData int true "推荐(1=推荐 0=取消)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/recommendation [patch]
func swaggerV2AdminEventsIdRecommendationPatch134() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 变更 /api/v2/admin/events/{id}/top
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param top formData int true "置顶(1=置顶 0=取消)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/top [patch]
func swaggerV2AdminEventsIdTopPatch135() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 查询 /api/v2/admin/events/{id}/participants
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/participants [get]
func swaggerV2AdminEventsIdParticipantsGet136() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 更新 /api/v2/admin/events/{id}/participants/{participantId}
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param participantId path int true "participantId"
// @Param forms formData string false "表单数据(JSON)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/participants/{participantId} [put]
func swaggerV2AdminEventsIdParticipantsParticipantIdPut137() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 删除 /api/v2/admin/events/{id}/participants/{participantId}
// @Security AdminToken
// @Param id path int true "id"
// @Param participantId path int true "participantId"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/participants/{participantId} [delete]
func swaggerV2AdminEventsIdParticipantsParticipantIdDelete138() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 删除 /api/v2/admin/events/{id}/participants
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param ids formData string true "参与记录ID列表(逗号分隔)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/participants [delete]
func swaggerV2AdminEventsIdParticipantsDelete139() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 查询 /api/v2/admin/events/{id}/dynamics
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/dynamics [get]
func swaggerV2AdminEventsIdDynamicsGet140() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 提交 /api/v2/admin/events/{id}/dynamics
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param title formData string false "动态标题"
// @Param content formData string false "动态内容"
// @Param images formData string false "图片列表(JSON)"
// @Param videos formData string false "视频列表(JSON)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/dynamics [post]
func swaggerV2AdminEventsIdDynamicsPost141() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 更新 /api/v2/admin/events/{id}/dynamics/{dynamicId}
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param dynamicId path int true "dynamicId"
// @Param title formData string false "动态标题"
// @Param content formData string false "动态内容"
// @Param images formData string false "图片列表(JSON)"
// @Param videos formData string false "视频列表(JSON)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/dynamics/{dynamicId} [put]
func swaggerV2AdminEventsIdDynamicsDynamicIdPut142() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 删除 /api/v2/admin/events/{id}/dynamics/{dynamicId}
// @Security AdminToken
// @Param id path int true "id"
// @Param dynamicId path int true "dynamicId"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/dynamics/{dynamicId} [delete]
func swaggerV2AdminEventsIdDynamicsDynamicIdDelete143() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 删除 /api/v2/admin/events/{id}/dynamics
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param ids formData string true "动态ID列表(逗号分隔)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/dynamics [delete]
func swaggerV2AdminEventsIdDynamicsDelete144() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 查询 /api/v2/admin/events/{id}/scores
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/scores [get]
func swaggerV2AdminEventsIdScoresGet145() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 提交 /api/v2/admin/events/{id}/scores
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param score formData string true "评分"
// @Param participantId formData string false "参赛者ID(新增时必填)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/scores [post]
func swaggerV2AdminEventsIdScoresPost1451() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 更新 /api/v2/admin/events/{id}/scores/{scoreId}
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param scoreId path int true "scoreId"
// @Param score formData string true "评分"
// @Param eventId formData string false "活动ID(新增时必填)"
// @Param participantId formData string false "参赛者ID(新增时必填)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/events/{id}/scores/{scoreId} [put]
func swaggerV2AdminEventsIdScoresScoreIdPut146() {}

// @Tags API v2-后台管理-赛事活动
// @Summary 查询 /api/v2/admin/event-dept-users
// @Security AdminToken
// @Param deptIds query string true "部门ID列表(逗号分隔)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/event-dept-users [get]
func swaggerV2AdminEventDeptUsersGet147() {}

// @Tags API v2-后台管理-字典管理
// @Summary 查询 /api/v2/admin/dict/types
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dict/types [get]
func swaggerV2AdminDictTypesGet148() {}

// @Tags API v2-后台管理-字典管理
// @Summary 新增字典类型
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param typeCode formData string false "字典类型编码"
// @Param typeName formData string false "字典类型名称"
// @Param remark formData string false "备注"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dict/types [post]
func swaggerV2AdminDictTypesPost() {}

// @Tags API v2-后台管理-字典管理
// @Summary 查询 /api/v2/admin/dict/items
// @Security AdminToken
// @Param typeCode query string true "类型编码"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dict/items [get]
func swaggerV2AdminDictItemsGet149() {}

// @Tags API v2-后台管理-字典管理
// @Summary 提交 /api/v2/admin/dict/items
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param typeCode formData string false "字典类型编码"
// @Param typeName formData string false "字典类型名称"
// @Param label formData string false "字典项显示名称"
// @Param value formData string false "字典项值"
// @Param remark formData string false "备注"
// @Param sort formData string false "排序值"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dict/items [post]
func swaggerV2AdminDictItemsPost150() {}

// @Tags API v2-后台管理-字典管理
// @Summary 更新 /api/v2/admin/dict/items/{id}
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param sort formData string false "排序值"
// @Param label formData string false "字典项显示名称"
// @Param value formData string false "字典项值"
// @Param remark formData string false "备注"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dict/items/{id} [put]
func swaggerV2AdminDictItemsIdPut151() {}

// @Tags API v2-后台管理-字典管理
// @Summary 删除 /api/v2/admin/dict/items/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dict/items/{id} [delete]
func swaggerV2AdminDictItemsIdDelete152() {}

// @Tags API v2-后台管理-字典管理
// @Summary 删除 /api/v2/admin/dict/types/{typeCode}/items
// @Security AdminToken
// @Param typeCode path string true "typeCode"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dict/types/{typeCode}/items [delete]
func swaggerV2AdminDictTypesTypeCodeItemsDelete153() {}

// @Tags API v2-后台管理-字典管理
// @Summary 变更 /api/v2/admin/dict/types/{typeCode}
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param typeCode path string true "typeCode"
// @Param typeName formData string false "字典类型名称"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dict/types/{typeCode} [patch]
func swaggerV2AdminDictTypesTypeCodePatch154() {}

// @Tags API v2-后台管理-字典管理
// @Summary 更新字典类型
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param typeCode path string true "typeCode"
// @Param typeName formData string false "字典类型名称"
// @Param remark formData string false "备注"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dict/types/{typeCode} [put]
func swaggerV2AdminDictTypesTypeCodePut() {}

// @Tags API v2-后台管理-字典管理
// @Summary 删除字典类型及其数据
// @Security AdminToken
// @Param typeCode path string true "typeCode"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dict/types/{typeCode} [delete]
func swaggerV2AdminDictTypesTypeCodeDelete() {}

// @Tags API v2-后台管理-组织管理
// @Summary 查询 /api/v2/admin/departments/tree
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/departments/tree [get]
func swaggerV2AdminDepartmentsTreeGet155() {}

// @Tags API v2-后台管理-组织管理
// @Summary 提交 /api/v2/admin/departments
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param name formData string true "部门名称"
// @Param parentId formData int false "父部门ID"
// @Param sort formData int false "排序"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/departments [post]
func swaggerV2AdminDepartmentsPost156() {}

// @Tags API v2-后台管理-组织管理
// @Summary 更新 /api/v2/admin/departments/{id}
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param name formData string true "部门名称"
// @Param parentId formData int false "父部门ID"
// @Param sort formData int false "排序"
// @Param status formData int false "状态(1=启用 0=禁用)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/departments/{id} [put]
func swaggerV2AdminDepartmentsIdPut157() {}

// @Tags API v2-后台管理-组织管理
// @Summary 删除 /api/v2/admin/departments/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/departments/{id} [delete]
func swaggerV2AdminDepartmentsIdDelete158() {}

// @Tags API v2-后台管理-组织管理
// @Summary 查询 /api/v2/admin/positions
// @Security AdminToken
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/positions [get]
func swaggerV2AdminPositionsGet() {}

// @Tags API v2-后台管理-组织管理
// @Summary 提交 /api/v2/admin/positions
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param name formData string false "名称"
// @Param sort formData string false "排序值"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/positions [post]
func swaggerV2AdminPositionsPost() {}

// @Tags API v2-后台管理-组织管理
// @Summary 更新 /api/v2/admin/positions/{id}
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param name formData string false "名称"
// @Param sort formData string false "排序值"
// @Param status formData string false "状态"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/positions/{id} [put]
func swaggerV2AdminPositionsIdPut() {}

// @Tags API v2-后台管理-组织管理
// @Summary 删除 /api/v2/admin/positions/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/positions/{id} [delete]
func swaggerV2AdminPositionsIdDelete() {}

// @Tags API v2-后台管理-角色权限
// @Summary 查询 /api/v2/admin/roles
// @Security AdminToken
// @Param keyword query string false "搜索关键词"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/roles [get]
func swaggerV2AdminRolesGet159() {}

// @Tags API v2-后台管理-角色权限
// @Summary 提交 /api/v2/admin/roles
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param name formData string true "角色名称"
// @Param remark formData string false "备注"
// @Param sort formData int false "排序"
// @Param dataScope formData int false "数据权限范围(1=全部 2=本部门及子部门 3=本人 4=自定义部门)"
// @Param adminPermissionKeys formData string false "后台权限编码列表(逗号分隔)"
// @Param adminApiPermissionKeys formData string false "后台接口权限编码列表(逗号分隔)"
// @Param deptIds formData string false "部门ID列表(逗号分隔)"
// @Param allowAdminLogin formData string false "是否允许登录管理后台"
// @Param clientMenuKeys formData string false "客户端菜单键列表"
// @Param dingtalkH5MenuKeys formData string false "H5App 菜单键列表"
// @Param clientApiPermissionKeys formData string false "客户端 API 权限键列表"
// @Param dingtalkH5ApiPermissionKeys formData string false "H5App API 权限键列表"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/roles [post]
func swaggerV2AdminRolesPost160() {}

// @Tags API v2-后台管理-角色权限
// @Summary 更新 /api/v2/admin/roles/{id}
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param name formData string true "角色名称"
// @Param remark formData string false "备注"
// @Param sort formData int false "排序"
// @Param status formData int false "状态(1=启用 0=禁用)"
// @Param dataScope formData int false "数据权限范围(1=全部 2=本部门及子部门 3=本人 4=自定义部门)"
// @Param adminPermissionKeys formData string false "后台权限编码列表(逗号分隔)"
// @Param adminApiPermissionKeys formData string false "后台接口权限编码列表(逗号分隔)"
// @Param deptIds formData string false "部门ID列表(逗号分隔)"
// @Param allowAdminLogin formData string false "是否允许登录管理后台"
// @Param clientMenuKeys formData string false "客户端菜单键列表"
// @Param dingtalkH5MenuKeys formData string false "H5App 菜单键列表"
// @Param clientApiPermissionKeys formData string false "客户端 API 权限键列表"
// @Param dingtalkH5ApiPermissionKeys formData string false "H5App API 权限键列表"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/roles/{id} [put]
func swaggerV2AdminRolesIdPut161() {}

// @Tags API v2-后台管理-角色权限
// @Summary 删除 /api/v2/admin/roles/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/roles/{id} [delete]
func swaggerV2AdminRolesIdDelete162() {}

// @Tags API v2-后台管理-角色权限
// @Summary 删除 /api/v2/admin/roles
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param ids formData string true "角色ID列表(逗号分隔)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/roles [delete]
func swaggerV2AdminRolesDelete163() {}

// @Tags API v2-后台管理-角色权限
// @Summary 查询 /api/v2/admin/roles/application-permissions
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/roles/application-permissions [get]
func swaggerV2AdminRolesApplicationPermissionsGet164() {}

// @Tags API v2-后台管理-角色权限
// @Summary 查询 /api/v2/admin/permissions/tree
// @Security AdminToken
// @Param platform query string false "平台"
// @Param types query string false "权限类型，逗号分隔"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/permissions/tree [get]
func swaggerV2AdminPermissionsTreeGet165() {}

// @Tags API v2-后台管理-角色权限
// @Summary 查询 /api/v2/admin/permissions
// @Security AdminToken
// @Param platform query string false "平台"
// @Param types query string false "权限类型，逗号分隔"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/permissions [get]
func swaggerV2AdminPermissionsGet166() {}

// @Tags API v2-后台管理-角色权限
// @Summary 提交 /api/v2/admin/permissions
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param key formData string true "权限键"
// @Param name formData string true "权限名称"
// @Param platform formData string true "所属平台"
// @Param type formData string true "权限类型"
// @Param parentKey formData string false "父权限键"
// @Param resourcePath formData string false "受保护资源路径"
// @Param path formData string false "菜单或路由路径"
// @Param perms formData string false "权限标识"
// @Param icon formData string false "图标名称"
// @Param sort formData int false "排序值"
// @Param status formData int false "状态"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/permissions [post]
func swaggerV2AdminPermissionsPost167() {}

// @Tags API v2-后台管理-角色权限
// @Summary 更新 /api/v2/admin/permissions/{key}
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param key path string true "权限编码"
// @Param originalKey formData string false "原权限键"
// @Param oldKey formData string false "原权限键（兼容参数）"
// @Param name formData string true "权限名称"
// @Param platform formData string true "所属平台"
// @Param type formData string true "权限类型"
// @Param parentKey formData string false "父权限键"
// @Param resourcePath formData string false "受保护资源路径"
// @Param path formData string false "菜单或路由路径"
// @Param perms formData string false "权限标识"
// @Param icon formData string false "图标名称"
// @Param sort formData int false "排序值"
// @Param status formData int false "状态"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/permissions/{key} [put]
func swaggerV2AdminPermissionsKeyPut168() {}

// @Tags API v2-后台管理-角色权限
// @Summary 删除 /api/v2/admin/permissions/{key}
// @Security AdminToken
// @Param key path string true "权限编码"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/permissions/{key} [delete]
func swaggerV2AdminPermissionsKeyDelete169() {}

// @Tags API v2-后台管理-角色权限
// @Summary 查询 /api/v2/admin/me/menus
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/me/menus [get]
func swaggerV2AdminMeMenusGet169() {}

// @Tags API v2-后台管理-角色权限
// @Summary 查询 /api/v2/admin/me/perms
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/me/perms [get]
func swaggerV2AdminMePermsGet170() {}

// @Tags API v2-后台管理-管理员
// @Summary 变更 /api/v2/admin/me/password
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id formData string true "管理员ID"
// @Param password formData string true "新密码"
// @Param oldPassword formData string false "原密码"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/me/password [patch]
func swaggerV2AdminMePasswordPatch1701() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/survey-types
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-types [get]
func swaggerV2AdminSurveyTypesGet171() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 提交 /api/v2/admin/survey-schema/parse
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param schema formData string true "Schema JSON"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-schema/parse [post]
func swaggerV2AdminSurveySchemaParsePost172() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 提交 /api/v2/admin/survey-expressions/evaluate
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param expr formData string true "表达式"
// @Param env formData string false "环境变量JSON"
// @Param asBool formData bool false "是否返回布尔值"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-expressions/evaluate [post]
func swaggerV2AdminSurveyExpressionsEvaluatePost173() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/survey-report/enroll
// @Security AdminToken
// @Param enrollId query string true "打卡项目ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-report/enroll [get]
func swaggerV2AdminSurveyReportEnrollGet174() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/survey-report/enroll/export
// @Security AdminToken
// @Param enrollId query string true "打卡项目ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-report/enroll/export [get]
func swaggerV2AdminSurveyReportEnrollExportGet175() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/survey-report/event
// @Security AdminToken
// @Param eventId query string true "活动ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-report/event [get]
func swaggerV2AdminSurveyReportEventGet176() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/survey-report/event/export
// @Security AdminToken
// @Param eventId query string true "活动ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-report/event/export [get]
func swaggerV2AdminSurveyReportEventExportGet177() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/survey-report/survey
// @Security AdminToken
// @Param surveyId query int true "问卷ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-report/survey [get]
func swaggerV2AdminSurveyReportSurveyGet178() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/survey-report/survey/export
// @Security AdminToken
// @Param surveyId query int true "问卷ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-report/survey/export [get]
func swaggerV2AdminSurveyReportSurveyExportGet179() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/surveys
// @Security AdminToken
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param keyword query string false "关键词"
// @Param category query string false "分类"
// @Param status query int false "状态(0草稿 1发布)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys [get]
func swaggerV2AdminSurveysGet180() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 提交 /api/v2/admin/surveys
// @Security AdminToken
// @Accept application/json
// @Param survey body model.Survey true "问卷数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys [post]
func swaggerV2AdminSurveysPost181() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/surveys/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id} [get]
func swaggerV2AdminSurveysIdGet182() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 更新 /api/v2/admin/surveys/{id}
// @Security AdminToken
// @Accept application/json
// @Param id path int true "id"
// @Param survey body model.Survey true "问卷数据（需包含ID）"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id} [put]
func swaggerV2AdminSurveysIdPut183() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 删除 /api/v2/admin/surveys/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id} [delete]
func swaggerV2AdminSurveysIdDelete184() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 变更 /api/v2/admin/surveys/{id}/status
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param status formData int true "状态(0草稿 1发布)"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id}/status [patch]
func swaggerV2AdminSurveysIdStatusPatch185() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 提交 /api/v2/admin/surveys/{id}/copy
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id}/copy [post]
func swaggerV2AdminSurveysIdCopyPost186() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/surveys/{id}/statistics
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id}/statistics [get]
func swaggerV2AdminSurveysIdStatisticsGet187() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/surveys/{id}/responses
// @Security AdminToken
// @Param id path int true "id"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id}/responses [get]
func swaggerV2AdminSurveysIdResponsesGet188() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/surveys/{id}/responses/export
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id}/responses/export [get]
func swaggerV2AdminSurveysIdResponsesExportGet189() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/survey-responses/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-responses/{id} [get]
func swaggerV2AdminSurveyResponsesIdGet190() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 删除 /api/v2/admin/survey-responses/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-responses/{id} [delete]
func swaggerV2AdminSurveyResponsesIdDelete191() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 删除 /api/v2/admin/survey-responses
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param ids formData string true "逗号分隔的答卷ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-responses [delete]
func swaggerV2AdminSurveyResponsesDelete192() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/surveys/{id}/channels
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id}/channels [get]
func swaggerV2AdminSurveysIdChannelsGet193() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 提交 /api/v2/admin/surveys/{id}/channels
// @Security AdminToken
// @Accept application/json
// @Param id path int true "id"
// @Param channel body model.SurveyChannel true "渠道数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id}/channels [post]
func swaggerV2AdminSurveysIdChannelsPost194() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 删除 /api/v2/admin/survey-channels/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-channels/{id} [delete]
func swaggerV2AdminSurveyChannelsIdDelete195() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 提交 /api/v2/admin/survey-resources
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param file formData file true "文件"
// @Param surveyId formData int true "问卷ID"
// @Param resType formData string true "资源类型: bg/header"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-resources [post]
func swaggerV2AdminSurveyResourcesPost196() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/surveys/{id}/resources
// @Security AdminToken
// @Param id path int true "id"
// @Param resType query string false "资源类型: bg/header，为空则返回全部"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/surveys/{id}/resources [get]
func swaggerV2AdminSurveysIdResourcesGet197() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 删除 /api/v2/admin/survey-resources/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-resources/{id} [delete]
func swaggerV2AdminSurveyResourcesIdDelete198() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/survey-question-bank
// @Security AdminToken
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
// @Param keyword query string false "关键词"
// @Param category query string false "分类"
// @Param type query string false "类型"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-question-bank [get]
func swaggerV2AdminSurveyQuestionBankGet199() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 提交 /api/v2/admin/survey-question-bank
// @Security AdminToken
// @Accept application/json
// @Param body body SurveyQuestionBankRequest true "题库题目"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-question-bank [post]
func swaggerV2AdminSurveyQuestionBankPost200() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 更新 /api/v2/admin/survey-question-bank/{id}
// @Security AdminToken
// @Accept application/json
// @Param id path int true "id"
// @Param body body SurveyQuestionBankRequest true "题库题目"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-question-bank/{id} [put]
func swaggerV2AdminSurveyQuestionBankIdPut201() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 删除 /api/v2/admin/survey-question-bank/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-question-bank/{id} [delete]
func swaggerV2AdminSurveyQuestionBankIdDelete202() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/survey-question-bank/categories
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-question-bank/categories [get]
func swaggerV2AdminSurveyQuestionBankCategoriesGet203() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/survey-notifications
// @Security AdminToken
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
// @Param sourceType query string false "来源类型"
// @Param sourceId query string false "来源业务 ID"
// @Param userId query string false "用户 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-notifications [get]
func swaggerV2AdminSurveyNotificationsGet204() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 变更 /api/v2/admin/survey-notifications/{id}/read
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param all formData bool false "是否将全部通知标记为已读"
// @Param userId formData string false "用户 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-notifications/{id}/read [patch]
func swaggerV2AdminSurveyNotificationsIdReadPatch205() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/survey-notifications/unread-count
// @Security AdminToken
// @Param userId query string false "用户 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-notifications/unread-count [get]
func swaggerV2AdminSurveyNotificationsUnreadCountGet206() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 查询 /api/v2/admin/survey-template-presets
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-template-presets [get]
func swaggerV2AdminSurveyTemplatePresetsGet207() {}

// @Tags API v2-后台管理-问卷管理
// @Summary 更新 /api/v2/admin/survey-template-presets
// @Security AdminToken
// @Accept application/json
// @Param body body SurveyTemplatePresetsRequest true "模板预设列表"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/survey-template-presets [put]
func swaggerV2AdminSurveyTemplatePresetsPut208() {}

// @Tags API v2-后台管理-考试管理
// @Summary 查询 /api/v2/admin/exams
// @Security AdminToken
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param keyword query string false "关键词"
// @Param category query string false "分类"
// @Param status query int false "状态"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams [get]
func swaggerV2AdminExamsGet209() {}

// @Tags API v2-后台管理-考试管理
// @Summary 提交 /api/v2/admin/exams
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param title formData string true "标题"
// @Param description formData string false "描述"
// @Param category formData string false "分类"
// @Param tags formData string false "标签"
// @Param visibility formData int false "可见性"
// @Param allowMulti formData int false "允许多次"
// @Param anonymous formData int false "匿名"
// @Param showResult formData int false "显示结果"
// @Param startTime formData int false "开始时间"
// @Param endTime formData int false "结束时间"
// @Param maxResponse formData int false "最大答卷数"
// @Param duration formData int false "答题时长"
// @Param maxAttempts formData int false "最大次数"
// @Param showScore formData int false "显示分数"
// @Param status formData int false "状态"
// @Param schema formData string false "题目JSON"
// @Param deptIds formData string false "部门ID"
// @Param mode formData string false "模式"
// @Param settings formData string false "设置JSON"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams [post]
func swaggerV2AdminExamsPost210() {}

// @Tags API v2-后台管理-考试管理
// @Summary 查询 /api/v2/admin/exams/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id} [get]
func swaggerV2AdminExamsIdGet211() {}

// @Tags API v2-后台管理-考试管理
// @Summary 更新 /api/v2/admin/exams/{id}
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param title formData string true "标题"
// @Param description formData string false "描述"
// @Param category formData string false "分类"
// @Param tags formData string false "标签"
// @Param visibility formData int false "可见性"
// @Param allowMulti formData int false "允许多次"
// @Param anonymous formData int false "匿名"
// @Param showResult formData int false "显示结果"
// @Param startTime formData int false "开始时间"
// @Param endTime formData int false "结束时间"
// @Param maxResponse formData int false "最大答卷数"
// @Param duration formData int false "答题时长"
// @Param maxAttempts formData int false "最大次数"
// @Param showScore formData int false "显示分数"
// @Param status formData int false "状态"
// @Param schema formData string false "题目JSON"
// @Param deptIds formData string false "部门ID"
// @Param mode formData string false "模式"
// @Param settings formData string false "设置JSON"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id} [put]
func swaggerV2AdminExamsIdPut212() {}

// @Tags API v2-后台管理-考试管理
// @Summary 删除 /api/v2/admin/exams/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id} [delete]
func swaggerV2AdminExamsIdDelete213() {}

// @Tags API v2-后台管理-考试管理
// @Summary 变更 /api/v2/admin/exams/{id}/status
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param status formData int true "状态"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id}/status [patch]
func swaggerV2AdminExamsIdStatusPatch214() {}

// @Tags API v2-后台管理-考试管理
// @Summary 查询 /api/v2/admin/exams/{id}/records
// @Security AdminToken
// @Param id path int true "id"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id}/records [get]
func swaggerV2AdminExamsIdRecordsGet215() {}

// @Tags API v2-后台管理-考试管理
// @Summary 查询 /api/v2/admin/exams/{id}/records/{recordId}
// @Security AdminToken
// @Param id path int true "id"
// @Param recordId path int true "recordId"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id}/records/{recordId} [get]
func swaggerV2AdminExamsIdRecordsRecordIdGet216() {}

// @Tags API v2-后台管理-考试管理
// @Summary 删除 /api/v2/admin/exams/{id}/records/{recordId}
// @Security AdminToken
// @Param id path int true "id"
// @Param recordId path int true "recordId"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id}/records/{recordId} [delete]
func swaggerV2AdminExamsIdRecordsRecordIdDelete217() {}

// @Tags API v2-后台管理-考试管理
// @Summary 删除 /api/v2/admin/exams/{id}/records
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param id path int true "id"
// @Param ids formData string true "逗号分隔的记录ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id}/records [delete]
func swaggerV2AdminExamsIdRecordsDelete218() {}

// @Tags API v2-后台管理-考试管理
// @Summary 查询 /api/v2/admin/exams/{id}/statistics
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id}/statistics [get]
func swaggerV2AdminExamsIdStatisticsGet219() {}

// @Tags API v2-后台管理-考试管理
// @Summary 提交 /api/v2/admin/exam-resources
// @Security AdminToken
// @Accept application/x-www-form-urlencoded
// @Param file formData file true "文件"
// @Param examId formData int true "考试ID"
// @Param resType formData string true "资源类型: bg/header"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exam-resources [post]
func swaggerV2AdminExamResourcesPost220() {}

// @Tags API v2-后台管理-考试管理
// @Summary 查询 /api/v2/admin/exams/{id}/resources
// @Security AdminToken
// @Param id path int true "id"
// @Param resType query string false "资源类型: bg/header"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exams/{id}/resources [get]
func swaggerV2AdminExamsIdResourcesGet221() {}

// @Tags API v2-后台管理-考试管理
// @Summary 删除 /api/v2/admin/exam-resources/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exam-resources/{id} [delete]
func swaggerV2AdminExamResourcesIdDelete222() {}

// @Tags API v2-后台管理-考试管理
// @Summary 查询 /api/v2/admin/exam-question-bank
// @Security AdminToken
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "搜索关键词"
// @Param category query string false "分类"
// @Param type query string false "类型"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exam-question-bank [get]
func swaggerV2AdminExamQuestionBankGet223() {}

// @Tags API v2-后台管理-考试管理
// @Summary 提交 /api/v2/admin/exam-question-bank
// @Security AdminToken
// @Accept application/json
// @Param title body string true "题干"
// @Param type body string true "题型"
// @Param schema body string true "完整 formkit JSON"
// @Param category body string false "分类"
// @Param tags body string false "标签"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exam-question-bank [post]
func swaggerV2AdminExamQuestionBankPost224() {}

// @Tags API v2-后台管理-考试管理
// @Summary 更新 /api/v2/admin/exam-question-bank/{id}
// @Security AdminToken
// @Accept application/json
// @Param id path int true "id"
// @Param title body string false "题干"
// @Param type body string false "题型"
// @Param schema body string false "formkit JSON"
// @Param category body string false "分类"
// @Param tags body string false "标签"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exam-question-bank/{id} [put]
func swaggerV2AdminExamQuestionBankIdPut225() {}

// @Tags API v2-后台管理-考试管理
// @Summary 删除 /api/v2/admin/exam-question-bank/{id}
// @Security AdminToken
// @Param id path int true "id"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exam-question-bank/{id} [delete]
func swaggerV2AdminExamQuestionBankIdDelete226() {}

// @Tags API v2-后台管理-考试管理
// @Summary 查询 /api/v2/admin/exam-question-bank/categories
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exam-question-bank/categories [get]
func swaggerV2AdminExamQuestionBankCategoriesGet227() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询工作流定义列表
// @Security AdminToken
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
// @Param status query string false "状态"
// @Param keyword query string false "关键词"
// @Param category query string false "分类"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-definitions [get]
func swaggerV2AdminWorkflowDefinitionsGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 新建工作流定义
// @Security AdminToken
// @Accept multipart/form-data
// @Description 同时兼容 application/json；上传流程图标时使用 multipart/form-data
// @Param key formData string true "流程编码"
// @Param name formData string true "流程名称"
// @Param description formData string false "流程描述"
// @Param category formData string false "流程分类"
// @Param draft formData string false "流程草稿定义（JSON）"
// @Param logo formData file false "流程图标（PNG、JPG、JPEG 或 WebP，最大 2MB）"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-definitions [post]
func swaggerV2AdminWorkflowDefinitionsPost() {}

// @Tags API v2-后台管理-工作流
// @Summary 复制工作流定义
// @Security AdminToken
// @Accept multipart/form-data
// @Description 仅复制源流程当前设计草稿；名称、编码、分类、说明和图标使用本次请求，发布版本和版本历史不复制。兼容 application/json
// @Param id path int true "源流程定义 ID"
// @Param key formData string true "新流程编码"
// @Param name formData string true "新流程名称"
// @Param description formData string false "新流程描述"
// @Param category formData string false "新流程分类"
// @Param logo formData file false "新流程图标（PNG、JPG、JPEG 或 WebP，最大 2MB）"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-definitions/{id}/copy [post]
func swaggerV2AdminWorkflowDefinitionsIDCopyPost() {}

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
// @Accept multipart/form-data
// @Description 同时兼容 application/json；上传流程图标时使用 multipart/form-data
// @Param id path int true "流程定义 ID"
// @Param name formData string false "流程名称"
// @Param description formData string false "流程描述"
// @Param category formData string false "流程分类"
// @Param status formData int false "流程状态"
// @Param draft formData string false "流程草稿定义（JSON）"
// @Param logo formData file false "流程图标（PNG、JPG、JPEG 或 WebP，最大 2MB）"
// @Param removeLogo formData string false "是否移除流程图标"
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
// @Accept application/json
// @Param id path int true "流程定义 ID"
// @Param body body workflowservice.PublishRequest true "发布配置"
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
// @Summary 查询工作流版本变更内容
// @Security AdminToken
// @Param id path int true "流程定义 ID"
// @Param version path int true "目标版本"
// @Param compareTo query int false "对比版本；不传时使用发布时基准版本"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-definitions/{id}/versions/{version}/changes [get]
func swaggerV2AdminWorkflowDefinitionsIDVersionChangesGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 删除未被引用的历史工作流版本
// @Security AdminToken
// @Param id path int true "流程定义 ID"
// @Param version path int true "待删除版本"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-definitions/{id}/versions/{version} [delete]
func swaggerV2AdminWorkflowDefinitionsIDVersionDelete() {}

// @Tags API v2-后台管理-工作流
// @Summary 回滚工作流版本并生成新发布版本
// @Security AdminToken
// @Accept application/json
// @Param id path int true "流程定义 ID"
// @Param version path int true "回滚来源版本"
// @Param body body workflowservice.RollbackRequest false "回滚说明"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-definitions/{id}/versions/{version}/rollback [post]
func swaggerV2AdminWorkflowDefinitionsIDVersionRollbackPost() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询组织审批身份列表
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-org-approver-identities [get]
func swaggerV2AdminWorkflowOrgApproverIdentitiesGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询组织审批身份人员配置
// @Security AdminToken
// @Param subjectId query string false "组织主体 ID"
// @Param subjectType query string false "组织主体类型"
// @Param departmentId query string false "部门 ID（兼容参数）"
// @Param identityCode query string false "审批身份编码"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-org-approver-assignments [get]
func swaggerV2AdminWorkflowOrgApproverAssignmentsGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 保存组织审批身份人员配置
// @Security AdminToken
// @Accept application/json
// @Param body body workflowservice.SaveOrgApproverAssignmentsRequest true "组织主体、审批身份与关联用户"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-org-approver-assignments [put]
func swaggerV2AdminWorkflowOrgApproverAssignmentsPut() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询可发起的已发布工作流
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-published-definitions [get]
func swaggerV2AdminWorkflowPublishedDefinitionsGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询已发布工作流表单 Schema
// @Security AdminToken
// @Param id path int true "流程定义 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-published-definitions/{id} [get]
func swaggerV2AdminWorkflowPublishedDefinitionsIDGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询工作流可选用户
// @Security AdminToken
// @Param page query string false "页码"
// @Param size query string false "每页数量"
// @Param keyword query string false "关键词"
// @Param pageSize query string false "每页数量"
// @Param sort query string false "排序值"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-user-options [get]
func swaggerV2AdminWorkflowUserOptionsGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询工作流可选部门树
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-department-options [get]
func swaggerV2AdminWorkflowDepartmentOptionsGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询工作流实例列表
// @Security AdminToken
// @Param definitionId query string false "流程定义 ID"
// @Param definitionCategory query string false "流程定义分类"
// @Param status query string false "状态"
// @Param businessType query string false "业务类型"
// @Param businessKey query string false "业务键"
// @Param starterId query string false "发起人 ID"
// @Param startTimeFrom query string false "开始时间起始值（毫秒）"
// @Param startTimeTo query string false "开始时间截止值（毫秒）"
// @Param endTimeFrom query string false "结束时间起始值（毫秒）"
// @Param endTimeTo query string false "结束时间截止值（毫秒）"
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-instances [get]
func swaggerV2AdminWorkflowInstancesGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 启动工作流实例
// @Security AdminToken
// @Description starterId 为业务发起人，实际操作人从管理员登录态获取并保存为 operatorId
// @Accept application/json
// @Param body body AdminWorkflowStartInstanceRequest true "流程定义、业务标识与表单数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-instances [post]
func swaggerV2AdminWorkflowInstancesPost() {}

// @Tags API v2-后台管理-工作流
// @Summary 批量删除终态工作流实例
// @Security AdminToken
// @Description 仅允许删除已完成、已驳回或已取消的实例；删除后保留审计数据
// @Accept application/json
// @Param body body AdminWorkflowInstanceDeleteRequest true "流程实例 ID 列表"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-instances [delete]
func swaggerV2AdminWorkflowInstancesDelete() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询工作流实例详情
// @Security AdminToken
// @Param id path string true "流程实例 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-instances/{id} [get]
func swaggerV2AdminWorkflowInstancesIDGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 删除单个终态工作流实例
// @Security AdminToken
// @Description 仅允许删除已完成、已驳回或已取消的实例；删除后保留审计数据
// @Param id path string true "流程实例 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-instances/{id} [delete]
func swaggerV2AdminWorkflowInstancesIDDelete() {}

// @Tags API v2-后台管理-工作流
// @Summary 恢复工作流实例的定时节点
// @Security AdminToken
// @Param id path string true "流程实例 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-instances/{id}/resume [post]
func swaggerV2AdminWorkflowInstancesIDResumePost() {}

// @Tags API v2-后台管理-工作流
// @Summary 取消运行中的工作流实例
// @Security AdminToken
// @Accept application/json
// @Param id path string true "流程实例 ID"
// @Param body body WorkflowReasonRequest false "取消原因"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-instances/{id}/cancel [post]
func swaggerV2AdminWorkflowInstancesIDCancelPost() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询工作流任务列表
// @Security AdminToken
// @Param instanceId query string false "流程实例 ID"
// @Param assigneeId query string false "处理人 ID"
// @Param status query string false "状态"
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-tasks [get]
func swaggerV2AdminWorkflowTasksGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 完成工作流任务
// @Description reject 会终止整个流程；return 会退回至已执行过的上游人工节点并继续运行，未传 returnTargetNodeId 时默认上一人工节点
// @Security AdminToken
// @Accept application/json
// @Param id path string true "流程任务 ID"
// @Param body body WorkflowCompleteTaskRequest true "审批动作、意见与表单数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-tasks/{id}/complete [post]
func swaggerV2AdminWorkflowTasksIDCompletePost() {}

// @Tags API v2-后台管理-工作流
// @Summary 删除工作流任务
// @Security AdminToken
// @Param id path string true "流程任务 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-tasks/{id} [delete]
func swaggerV2AdminWorkflowTasksIDDelete() {}

// @Tags API v2-后台管理-工作流
// @Summary 查询工作流通知投递记录
// @Security AdminToken
// @Param instanceId query string false "流程实例 ID"
// @Param recipientUserId query string false "接收人用户 ID"
// @Param kind query string false "通知事件"
// @Param channel query string false "通知渠道"
// @Param status query string false "投递状态"
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-notifications [get]
func swaggerV2AdminWorkflowNotificationsGet() {}

// @Tags API v2-后台管理-工作流
// @Summary 投递到期的工作流通知
// @Security AdminToken
// @Accept application/json
// @Param body body WorkflowDispatchDueRequest false "单次投递上限"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-notifications/dispatch-due [post]
func swaggerV2AdminWorkflowNotificationsDispatchDuePost() {}

// @Tags API v2-后台管理-工作流
// @Summary 重试单条工作流通知
// @Security AdminToken
// @Param id path string true "通知 Outbox ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/workflow-notifications/{id}/retry [post]
func swaggerV2AdminWorkflowNotificationsIDRetryPost() {}

// @Tags API v2-后台管理-定时任务
// @Summary 查询定时任务列表
// @Security AdminToken
// @Param keyword query string false "关键词"
// @Param handlerType query string false "处理器类型"
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
// @Param enabled query string false "是否启用"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/scheduled-tasks [get]
func swaggerV2AdminScheduledTasksGet() {}

// @Tags API v2-后台管理-定时任务
// @Summary 创建定时任务
// @Security AdminToken
// @Accept application/json
// @Param body body scheduledtaskapp.CreateTaskRequest true "定时任务配置"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/scheduled-tasks [post]
func swaggerV2AdminScheduledTasksPost() {}

// @Tags API v2-后台管理-定时任务
// @Summary 查询定时任务详情
// @Security AdminToken
// @Param id path int true "定时任务 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/scheduled-tasks/{id} [get]
func swaggerV2AdminScheduledTasksIDGet() {}

// @Tags API v2-后台管理-定时任务
// @Summary 更新定时任务
// @Security AdminToken
// @Accept application/json
// @Param id path int true "定时任务 ID"
// @Param body body scheduledtaskapp.UpdateTaskRequest true "定时任务配置"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/scheduled-tasks/{id} [put]
func swaggerV2AdminScheduledTasksIDPut() {}

// @Tags API v2-后台管理-定时任务
// @Summary 删除定时任务
// @Security AdminToken
// @Param id path int true "定时任务 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/scheduled-tasks/{id} [delete]
func swaggerV2AdminScheduledTasksIDDelete() {}

// @Tags API v2-后台管理-定时任务
// @Summary 启用或停用定时任务
// @Security AdminToken
// @Accept application/json
// @Param id path int true "定时任务 ID"
// @Param body body ScheduledTaskStatusRequest true "启用状态与当前版本号"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/scheduled-tasks/{id}/status [patch]
func swaggerV2AdminScheduledTasksIDStatusPatch() {}

// @Tags API v2-后台管理-定时任务
// @Summary 立即运行定时任务
// @Security AdminToken
// @Param id path int true "定时任务 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/scheduled-tasks/{id}/run [post]
func swaggerV2AdminScheduledTasksIDRunPost() {}

// @Tags API v2-后台管理-定时任务
// @Summary 预览 Cron 执行时间
// @Security AdminToken
// @Accept application/json
// @Param body body ScheduledTaskCronPreviewRequest true "Cron 表达式、时区与预览数量"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/scheduled-tasks/cron-preview [post]
func swaggerV2AdminScheduledTasksCronPreviewPost() {}

// @Tags API v2-后台管理-定时任务
// @Summary 查询可用任务处理器
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/scheduled-task-handlers [get]
func swaggerV2AdminScheduledTaskHandlersGet() {}

// @Tags API v2-后台管理-定时任务
// @Summary 查询定时任务运行记录
// @Security AdminToken
// @Param taskId query string false "定时任务 ID"
// @Param status query string false "状态"
// @Param triggerType query string false "触发类型"
// @Param workerId query string false "执行节点 ID"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/scheduled-task-runs [get]
func swaggerV2AdminScheduledTaskRunsGet() {}

// @Tags API v2-后台管理-定时任务
// @Summary 查询定时任务运行详情和日志
// @Security AdminToken
// @Param id path string true "运行记录 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/scheduled-task-runs/{id} [get]
func swaggerV2AdminScheduledTaskRunsIDGet() {}

// @Tags API v2-后台管理-定时任务
// @Summary 重试失败的定时任务运行
// @Security AdminToken
// @Param id path string true "运行记录 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/scheduled-task-runs/{id}/retry [post]
func swaggerV2AdminScheduledTaskRunsIDRetryPost() {}

// @Tags API v2-后台管理-定时任务
// @Summary 取消定时任务运行
// @Security AdminToken
// @Param id path string true "运行记录 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/scheduled-task-runs/{id}/cancel [post]
func swaggerV2AdminScheduledTaskRunsIDCancelPost() {}

// @Tags API v2-后台管理-定时任务
// @Summary 查询定时任务执行节点
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/scheduled-task-workers [get]
func swaggerV2AdminScheduledTaskWorkersGet() {}

// @Tags API v2-后台管理-站内信
// @Summary 查询当前管理员站内信
// @Security AdminToken
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/in-app-notifications [get]
func swaggerV2AdminInAppNotificationsGet() {}

// @Tags API v2-后台管理-站内信
// @Summary 手动发送站内信
// @Security AdminToken
// @Accept application/json
// @Param body body InAppNotificationSendRequest true "站内信内容与收件范围"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/in-app-notifications [post]
func swaggerV2AdminInAppNotificationsPost() {}

// @Tags API v2-后台管理-站内信
// @Summary 查询当前管理员未读站内信数量
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/in-app-notifications/unread-count [get]
func swaggerV2AdminInAppNotificationsUnreadCountGet() {}

// @Tags API v2-后台管理-站内信
// @Summary 查询可选站内信收件范围
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/in-app-notifications/recipient-options [get]
func swaggerV2AdminInAppNotificationsRecipientOptionsGet() {}

// @Tags API v2-后台管理-站内信
// @Summary 标记当前管理员的一条站内信为已读
// @Security AdminToken
// @Param id path int true "站内信 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/in-app-notifications/{id}/read [patch]
func swaggerV2AdminInAppNotificationsIDReadPatch() {}

// @Tags API v2-后台管理-站内信
// @Summary 标记当前管理员的全部站内信为已读
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/in-app-notifications/read-all [patch]
func swaggerV2AdminInAppNotificationsReadAllPatch() {}

// @Tags API v2-后台管理-钉钉通知
// @Summary 查询可选钉钉通知收件范围
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk-notifications/recipient-options [get]
func swaggerV2AdminDingTalkNotificationsRecipientOptionsGet() {}

// @Tags API v2-后台管理-钉钉通知
// @Summary 手动发送钉钉通知
// @Security AdminToken
// @Accept application/json
// @Param body body DingTalkNotificationSendRequest true "钉钉通知内容与收件范围"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk-notifications [post]
func swaggerV2AdminDingTalkNotificationsPost() {}

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
// @Summary 查询我的 OA 流程发起草稿
// @Security ClientToken
// @Param definitionId path int true "流程定义 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/workflows/drafts/{definitionId} [get]
func swaggerV2WorkflowsDraftsDefinitionIDGet() {}

// @Tags API v2-客户端-OA流程
// @Summary 保存我的 OA 流程发起草稿
// @Security ClientToken
// @Accept application/json
// @Param definitionId path int true "流程定义 ID"
// @Param body body WorkflowStartDraftRequest true "发布版本与表单草稿"
// @Success 200 {object} response.Resp
// @Router /api/v2/workflows/drafts/{definitionId} [put]
func swaggerV2WorkflowsDraftsDefinitionIDPut() {}

// @Tags API v2-客户端-OA流程
// @Summary 删除我的 OA 流程发起草稿
// @Security ClientToken
// @Param definitionId path int true "流程定义 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/workflows/drafts/{definitionId} [delete]
func swaggerV2WorkflowsDraftsDefinitionIDDelete() {}

// @Tags API v2-客户端-OA流程
// @Summary 发起 OA 流程
// @Security ClientToken
// @Accept application/json
// @Param body body WorkflowStartInstanceRequest true "流程定义、业务标识与表单数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/workflows/instances [post]
func swaggerV2WorkflowsInstancesPost() {}

// @Tags API v2-客户端-OA流程
// @Summary 查询我的 OA 流程申请
// @Security ClientToken
// @Param scope query string false "查询范围: started, handled, copied" default(started)
// @Param definitionId query string false "流程定义 ID"
// @Param definitionCategory query string false "流程定义分类"
// @Param status query string false "状态"
// @Param businessType query string false "业务类型"
// @Param businessKey query string false "业务键"
// @Param startTimeFrom query string false "开始时间起始值（毫秒）"
// @Param startTimeTo query string false "开始时间截止值（毫秒）"
// @Param endTimeFrom query string false "结束时间起始值（毫秒）"
// @Param endTimeTo query string false "结束时间截止值（毫秒）"
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
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
// @Accept application/json
// @Param id path string true "流程实例 ID"
// @Param body body WorkflowReasonRequest false "撤回原因"
// @Success 200 {object} response.Resp
// @Router /api/v2/workflows/instances/{id}/withdraw [post]
func swaggerV2WorkflowsInstancesIDWithdrawPost() {}

// @Tags API v2-客户端-OA流程
// @Summary 查询我的 OA 流程任务
// @Security ClientToken
// @Param instanceId query string false "流程实例 ID"
// @Param status query string false "状态"
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
// @Success 200 {object} response.Resp
// @Router /api/v2/workflows/tasks [get]
func swaggerV2WorkflowsTasksGet() {}

// @Tags API v2-客户端-OA流程
// @Summary 处理我的 OA 流程任务
// @Description reject 会终止整个流程；return 会退回至已执行过的上游人工节点并继续运行，未传 returnTargetNodeId 时默认上一人工节点
// @Security ClientToken
// @Accept application/json
// @Param id path string true "流程任务 ID"
// @Param body body WorkflowCompleteTaskRequest true "审批动作、意见与表单数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/workflows/tasks/{id}/complete [post]
func swaggerV2WorkflowsTasksIDCompletePost() {}
