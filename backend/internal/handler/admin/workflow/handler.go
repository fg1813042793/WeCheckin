package workflowhandler

import (
	"bytes"
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/model"
	workflowservice "wecheckin/backend/internal/service/admin/workflow"
	"wecheckin/backend/internal/support/storage"
	"wecheckin/backend/pkg/response"
)

type AdminWorkflowHandler struct{}

func NewAdminWorkflowHandler() *AdminWorkflowHandler { return &AdminWorkflowHandler{} }

func (h *AdminWorkflowHandler) List(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	status := -1
	if raw := c.Query("status"); raw != "" {
		status, _ = strconv.Atoi(raw)
	}
	data, err := workflowservice.GetListContext(ctx, c.Query("keyword"), c.Query("category"), status, page, pageSize)
	if err != nil {
		response.Fail(c, "获取流程定义失败")
		return
	}
	response.JSON(c, data)
}

func (h *AdminWorkflowHandler) Create(ctx context.Context, c *app.RequestContext) {
	admin, ok := currentAdmin(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	var request workflowservice.CreateRequest
	var uploaded *storage.StoredFile
	if isWorkflowMultipart(c) {
		request = workflowCreateRequestFromMultipart(c)
		file, err := optionalWorkflowLogo(c)
		if err != nil {
			response.FailInternal(ctx, c, "admin.workflow.handler", "操作失败，请稍后重试", err)
			return
		}
		if file != nil {
			uploaded, err = saveWorkflowLogo(ctx, file)
			if err != nil {
				response.FailInternal(ctx, c, "admin.workflow.handler", "操作失败，请稍后重试", err)
				return
			}
			request.LogoURL = uploaded.URL
		}
	} else if err := c.BindAndValidate(&request); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	data, err := workflowservice.CreateContext(ctx, admin.ID, request)
	if err != nil {
		storage.RemoveLocal(uploaded)
		response.FailInternal(ctx, c, "admin.workflow.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, data)
}

func (h *AdminWorkflowHandler) Copy(ctx context.Context, c *app.RequestContext) {
	admin, ok := currentAdmin(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	sourceID, ok := routeID(c)
	if !ok {
		response.Fail(c, "源流程定义 ID 无效")
		return
	}
	var request workflowservice.CopyRequest
	var uploaded *storage.StoredFile
	if isWorkflowMultipart(c) {
		request = workflowCopyRequestFromMultipart(c)
		file, err := optionalWorkflowLogo(c)
		if err != nil {
			response.FailInternal(ctx, c, "admin.workflow.handler", "操作失败，请稍后重试", err)
			return
		}
		if file != nil {
			uploaded, err = saveWorkflowLogo(ctx, file)
			if err != nil {
				response.FailInternal(ctx, c, "admin.workflow.handler", "操作失败，请稍后重试", err)
				return
			}
			request.LogoURL = uploaded.URL
		}
	} else if err := c.BindAndValidate(&request); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	data, err := workflowservice.CopyContext(ctx, admin.ID, sourceID, request)
	if err != nil {
		storage.RemoveLocal(uploaded)
		response.FailInternal(ctx, c, "admin.workflow.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, data)
}

func (h *AdminWorkflowHandler) Detail(ctx context.Context, c *app.RequestContext) {
	id, ok := routeID(c)
	if !ok {
		response.Fail(c, "流程定义 ID 无效")
		return
	}
	data, err := workflowservice.GetDetailContext(ctx, id)
	if err != nil {
		response.FailInternal(ctx, c, "admin.workflow.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, data)
}

func (h *AdminWorkflowHandler) Update(ctx context.Context, c *app.RequestContext) {
	admin, ok := currentAdmin(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	id, ok := routeID(c)
	if !ok {
		response.Fail(c, "流程定义 ID 无效")
		return
	}
	var request workflowservice.UpdateRequest
	var uploaded *storage.StoredFile
	if isWorkflowMultipart(c) {
		request = workflowUpdateRequestFromMultipart(c)
		file, err := optionalWorkflowLogo(c)
		if err != nil {
			response.FailInternal(ctx, c, "admin.workflow.handler", "操作失败，请稍后重试", err)
			return
		}
		if file != nil {
			uploaded, err = saveWorkflowLogo(ctx, file)
			if err != nil {
				response.FailInternal(ctx, c, "admin.workflow.handler", "操作失败，请稍后重试", err)
				return
			}
			request.LogoURL = &uploaded.URL
		} else if string(c.FormValue("removeLogo")) == "true" {
			empty := ""
			request.LogoURL = &empty
		}
	} else if err := c.BindAndValidate(&request); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	data, err := workflowservice.UpdateContext(ctx, admin.ID, id, request)
	if err != nil {
		storage.RemoveLocal(uploaded)
		response.FailInternal(ctx, c, "admin.workflow.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, data)
}

func (h *AdminWorkflowHandler) Delete(ctx context.Context, c *app.RequestContext) {
	id, ok := routeID(c)
	if !ok {
		response.Fail(c, "流程定义 ID 无效")
		return
	}
	if err := workflowservice.DeleteContext(ctx, id); err != nil {
		response.FailInternal(ctx, c, "admin.workflow.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, nil)
}

func (h *AdminWorkflowHandler) Validate(ctx context.Context, c *app.RequestContext) {
	id, ok := routeID(c)
	if !ok {
		response.Fail(c, "流程定义 ID 无效")
		return
	}
	data, err := workflowservice.ValidateContext(ctx, id)
	if err != nil {
		response.FailInternal(ctx, c, "admin.workflow.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, data)
}

func (h *AdminWorkflowHandler) Publish(ctx context.Context, c *app.RequestContext) {
	admin, ok := currentAdmin(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	id, ok := routeID(c)
	if !ok {
		response.Fail(c, "流程定义 ID 无效")
		return
	}
	var request workflowservice.PublishRequest
	if len(bytes.TrimSpace(c.Request.Body())) > 0 {
		if err := c.BindAndValidate(&request); err != nil {
			response.Fail(c, "发布配置格式无效")
			return
		}
	}
	data, err := workflowservice.PublishContext(ctx, admin.ID, id, request)
	if err != nil {
		response.FailInternal(ctx, c, "admin.workflow.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, data)
}

func (h *AdminWorkflowHandler) Versions(ctx context.Context, c *app.RequestContext) {
	id, ok := routeID(c)
	if !ok {
		response.Fail(c, "流程定义 ID 无效")
		return
	}
	data, err := workflowservice.GetVersionsContext(ctx, id)
	if err != nil {
		response.FailInternal(ctx, c, "admin.workflow.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, data)
}

func (h *AdminWorkflowHandler) VersionChanges(ctx context.Context, c *app.RequestContext) {
	id, ok := routeID(c)
	if !ok {
		response.Fail(c, "流程定义 ID 无效")
		return
	}
	version, ok := routeVersion(c)
	if !ok {
		response.Fail(c, "流程版本参数无效")
		return
	}
	compareTo := 0
	if raw := c.Query("compareTo"); raw != "" {
		var err error
		compareTo, err = strconv.Atoi(raw)
		if err != nil || compareTo <= 0 {
			response.Fail(c, "对比版本参数无效")
			return
		}
	}
	data, err := workflowservice.GetVersionChangesContext(ctx, id, version, compareTo)
	if err != nil {
		response.FailInternal(ctx, c, "admin.workflow.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, data)
}

func (h *AdminWorkflowHandler) DeleteVersion(ctx context.Context, c *app.RequestContext) {
	id, ok := routeID(c)
	if !ok {
		response.Fail(c, "流程定义 ID 无效")
		return
	}
	version, ok := routeVersion(c)
	if !ok {
		response.Fail(c, "流程版本参数无效")
		return
	}
	if err := workflowservice.DeleteVersionContext(ctx, id, version); err != nil {
		response.FailInternal(ctx, c, "admin.workflow.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, nil)
}

func (h *AdminWorkflowHandler) RollbackVersion(ctx context.Context, c *app.RequestContext) {
	admin, ok := currentAdmin(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	id, ok := routeID(c)
	if !ok {
		response.Fail(c, "流程定义 ID 无效")
		return
	}
	version, ok := routeVersion(c)
	if !ok {
		response.Fail(c, "流程版本参数无效")
		return
	}
	var request workflowservice.RollbackRequest
	if len(bytes.TrimSpace(c.Request.Body())) > 0 {
		if err := c.BindAndValidate(&request); err != nil {
			response.Fail(c, "回滚配置格式无效")
			return
		}
	}
	data, err := workflowservice.RollbackVersionContext(ctx, admin.ID, id, version, request)
	if err != nil {
		response.FailInternal(ctx, c, "admin.workflow.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, data)
}

func (h *AdminWorkflowHandler) OrgApproverIdentities(ctx context.Context, c *app.RequestContext) {
	data, err := workflowservice.ListOrgApproverIdentitiesContext(ctx)
	if err != nil {
		response.Fail(c, "获取组织审批身份失败")
		return
	}
	response.JSON(c, data)
}

func (h *AdminWorkflowHandler) OrgApproverAssignments(ctx context.Context, c *app.RequestContext) {
	subjectID, _ := strconv.ParseUint(c.Query("subjectId"), 10, 64)
	subjectType := c.Query("subjectType")
	if subjectID == 0 {
		subjectID, _ = strconv.ParseUint(c.Query("departmentId"), 10, 64)
		if subjectID > 0 && subjectType == "" {
			subjectType = model.OrgApproverSubjectTypeDepartment
		}
	}
	data, err := workflowservice.ListOrgApproverAssignmentsContext(ctx, subjectType, uint(subjectID), c.Query("identityCode"))
	if err != nil {
		response.Fail(c, "获取组织审批身份人员失败")
		return
	}
	response.JSON(c, data)
}

func (h *AdminWorkflowHandler) SaveOrgApproverAssignments(ctx context.Context, c *app.RequestContext) {
	var request workflowservice.SaveOrgApproverAssignmentsRequest
	if err := c.BindAndValidate(&request); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	if err := workflowservice.SaveOrgApproverAssignmentsContext(ctx, request); err != nil {
		response.FailInternal(ctx, c, "admin.workflow.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, nil)
}

func currentAdmin(c *app.RequestContext) (*model.Admin, bool) {
	value, ok := c.Get("admin")
	if !ok {
		return nil, false
	}
	admin, ok := value.(*model.Admin)
	return admin, ok
}

func routeID(c *app.RequestContext) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(id), err == nil && id > 0
}

func routeVersion(c *app.RequestContext) (int, bool) {
	version, err := strconv.Atoi(c.Param("version"))
	return version, err == nil && version > 0
}
