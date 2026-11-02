package swagger

// @Tags API v2-后台管理-审计导出
// @Summary 导出当前筛选的用户列表
// @Description 保持与原列表相同的数据权限；最多1000条，返回JSON分页供客户端生成CSV，并记录导出审计事件。
// @Security AdminToken
// @Param keyword query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量（最多1000）" default(20)
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/users/export [get]
func swaggerUsersAuditExport() {}

// @Tags API v2-后台管理-审计导出
// @Summary 导出当前筛选的管理员列表
// @Description 保持与原列表相同的数据权限；最多1000条，返回JSON分页供客户端生成CSV，并记录导出审计事件。
// @Security AdminToken
// @Param keyword query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量（最多1000）" default(20)
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/managers/export [get]
func swaggerManagersAuditExport() {}

// @Tags API v2-后台管理-审计导出
// @Summary 导出当前筛选的角色列表
// @Description 保持与原列表相同的数据权限；最多1000条，返回JSON分页供客户端生成CSV，并记录导出审计事件。
// @Security AdminToken
// @Param keyword query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量（最多1000）" default(20)
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/roles/export [get]
func swaggerRolesAuditExport() {}
