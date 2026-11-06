package swagger

import portalapp "wecheckin/backend/internal/modules/publicportal/application"

var _ portalapp.PublicResponse

// @Tags API v2-公共页面
// @Summary 获取公开问卷与考试入口
// @Description 无需登录。仅展示后台选中、启用、仍已发布且允许公开访问的问卷或考试。已结束内容按页面设置隐藏。每项实时返回来源管理名称和 category 分类；分类为空时使用问卷/考试默认标签，kind 仍表示实际来源类型。不返回原表单设置、题目、答案或密码；关闭时 enabled=false 且 items 为空。URL 返回根相对填写路径，由浏览器按当前站点生成访问链接和二维码。使用 no-store 避免停用内容被缓存继续展示。
// @Success 200 {object} response.Resp{data=portalapp.PublicResponse}
// @Router /api/v2/public/activities [get]
func swaggerPublicActivities() {}

// @Tags API v2-后台管理-公共页面
// @Summary 读取公共页面配置与展示项
// @Security AdminToken
// @Description 需要 public-portal:list。来源不可管理时隐藏源管理名称并返回 canManage=false；配置不自动向公众发布。
// @Success 200 {object} response.Resp{data=portalapp.AdminResponse}
// @Router /api/v2/admin/public-portal [get]
func swaggerAdminPublicPortal() {}

// @Tags API v2-后台管理-公共页面
// @Summary 保存公共页面设置
// @Security AdminToken
// @Description 需要 public-portal:edit。version 乐观锁，冲突返回配置已更新。不需要配置公共域名；公共链接和二维码使用当前访问站点。
// @Accept json
// @Param body body portalapp.Settings true "页面设置及当前版本"
// @Success 200 {object} response.Resp{data=portalapp.Settings}
// @Router /api/v2/admin/public-portal/settings [put]
func swaggerAdminPublicPortalSettings() {}

// @Tags API v2-后台管理-公共页面
// @Summary 搜索可展示的问卷或考试
// @Security AdminToken
// @Description 需要 public-portal:list。结果仅包含已发布、公开且处于当前管理员数据范围的来源，供选择时区分管理名称与显示名称。
// @Param kind query string true "来源类型" Enums(survey,exam)
// @Param keyword query string false "管理名称或显示名称关键词"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量，最多100" default(20)
// @Success 200 {object} response.Resp{data=portalapp.OptionsResponse}
// @Router /api/v2/admin/public-portal/options [get]
func swaggerAdminPublicPortalOptions() {}

// @Tags API v2-后台管理-公共页面
// @Summary 添加公共页面展示项
// @Security AdminToken
// @Description 需要 public-portal:edit。后台再次校验来源范围和公开状态，来源类型加 ID 唯一，最多200项。title/description 留空时实时跟随来源显示名称和说明。
// @Accept json
// @Param body body portalapp.ItemInput true "来源、展示文案、启用状态及排序"
// @Success 200 {object} response.Resp{data=portalapp.AdminItem}
// @Router /api/v2/admin/public-portal/items [post]
func swaggerAdminPublicPortalCreateItem() {}

// @Tags API v2-后台管理-公共页面
// @Summary 更新公共页面展示项
// @Security AdminToken
// @Description 需要 public-portal:edit。version 乐观锁；对原来源和新来源的可管理范围做校验，不能通过修改 ID 将无权管理的内容公开。
// @Accept json
// @Param id path int true "展示项 ID"
// @Param body body portalapp.ItemUpdate true "展示项完整配置与版本"
// @Success 200 {object} response.Resp{data=portalapp.AdminItem}
// @Router /api/v2/admin/public-portal/items/{id} [put]
func swaggerAdminPublicPortalUpdateItem() {}

// @Tags API v2-后台管理-公共页面
// @Summary 移除公共页面展示项
// @Security AdminToken
// @Description 仅移除展示配置，不删除关联问卷或考试。需要 public-portal:edit 和来源可管理范围；已删除来源允许创建者或超级管理员清理。
// @Param id path int true "展示项 ID"
// @Param version query int true "当前展示项版本"
// @Success 200 {object} response.Resp{data=portalapp.DeleteResponse}
// @Router /api/v2/admin/public-portal/items/{id} [delete]
func swaggerAdminPublicPortalDeleteItem() {}

// @Tags API v2-后台管理-公共页面
// @Summary 预览已保存公共页面
// @Security AdminToken
// @Description 需要 public-portal:list。仅绕过页面总开关，仍过滤未启用和非公开的来源；不会发布页面或保存未提交的编辑内容。
// @Success 200 {object} response.Resp{data=portalapp.PublicResponse}
// @Router /api/v2/admin/public-portal/preview [get]
func swaggerAdminPublicPortalPreview() {}
