package passport

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	passportservice "wecheckin/backend/internal/service/client/passport"
	rd "wecheckin/backend/pkg/redis"
	"wecheckin/backend/pkg/response"
	"wecheckin/backend/pkg/tokenutil"
)

type PassportHandler struct{}

func NewPassportHandler() *PassportHandler { return &PassportHandler{} }

// @Tags 客户端-通行证
// @Summary 用户登录
// @Param user_id formData string true "用户ID"
// @Success 200 {object} response.Resp
// @Router /passport/login [post]
func (h *PassportHandler) Login(ctx context.Context, c *app.RequestContext) {
	userID := c.PostForm("user_id")
	addIP := c.ClientIP()
	device := string(c.UserAgent())
	data, err := passportservice.LoginUserContext(ctx, userID, addIP, device)
	if err != nil {
		response.Fail(c, "登录失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 客户端-通行证
// @Summary 密码登录
// @Param name formData string true "用户名/手机号"
// @Param pwd formData string true "密码"
// @Success 200 {object} response.Resp
// @Router /passport/login_pwd [post]
func (h *PassportHandler) LoginByPwd(ctx context.Context, c *app.RequestContext) {
	name := c.PostForm("name")
	pwd := c.PostForm("pwd")
	if name == "" || pwd == "" {
		var req struct {
			Name string `json:"name"`
			Pwd  string `json:"pwd"`
		}
		if err := c.BindAndValidate(&req); err != nil {
			response.Fail(c, "参数错误")
			return
		}
		name = req.Name
		pwd = req.Pwd
	}
	addIP := c.ClientIP()
	device := string(c.UserAgent())
	data, err := passportservice.LoginByPwdContext(ctx, name, pwd, addIP, device)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

// @Tags 客户端-通行证
// @Summary 获取手机号
// @Param cloud_id formData string true "云ID"
// @Success 200 {object} response.Resp
// @Router /passport/phone [post]
func (h *PassportHandler) GetPhone(ctx context.Context, c *app.RequestContext) {
	cloudID := c.PostForm("cloud_id")
	data, err := passportservice.GetPhone(cloudID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 客户端-通行证
// @Summary 用户注册
// @Param user_id formData string true "用户ID"
// @Param name formData string true "姓名"
// @Param mobile formData string true "手机号"
// @Param pic formData string false "头像"
// @Success 200 {object} response.Resp
// @Router /passport/register [post]
func (h *PassportHandler) Register(ctx context.Context, c *app.RequestContext) {
	userID := c.PostForm("user_id")
	name := c.PostForm("name")
	mobile := c.PostForm("mobile")
	pic := c.PostForm("pic")
	forms := c.PostForm("forms")
	addIP := c.ClientIP()
	device := string(c.UserAgent())
	data, err := passportservice.RegisterUserContext(ctx, userID, mobile, name, pic, forms, 1, addIP, device)
	if err != nil {
		response.Fail(c, "注册失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 客户端-通行证
// @Summary 获取我的详情
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /passport/my_detail [get]
func (h *PassportHandler) GetMyDetail(ctx context.Context, c *app.RequestContext) {
	userIDVal, _ := c.Get("user_openid")
	userID, _ := userIDVal.(string)
	data, err := passportservice.GetMyDetailContext(ctx, userID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

func (h *PassportHandler) Bootstrap(ctx context.Context, c *app.RequestContext) {
	if userIDVal, ok := c.Get("user_id"); ok {
		if userID, ok := userIDVal.(uint); ok && userID > 0 {
			data, err := passportservice.BootstrapByIDContext(ctx, userID)
			if err != nil {
				response.Fail(c, "获取失败")
				return
			}
			response.JSON(c, data)
			return
		}
	}
	userIDVal, _ := c.Get("user_openid")
	userID, _ := userIDVal.(string)
	data, err := passportservice.BootstrapContext(ctx, userID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 客户端-通行证
// @Summary 编辑基本信息
// @Param name formData string false "姓名"
// @Param mobile formData string false "手机号"
// @Param pic formData string false "头像"
// @Param user_id formData string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /passport/edit_base [post]
func (h *PassportHandler) EditBase(ctx context.Context, c *app.RequestContext) {
	name := c.PostForm("name")
	mobile := c.PostForm("mobile")
	pic := c.PostForm("pic")
	forms := c.PostForm("forms")
	userIDVal, _ := c.Get("user_openid")
	userID, _ := userIDVal.(string)
	err := passportservice.EditBaseContext(ctx, userID, mobile, name, pic, forms)
	if err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	response.JSON(c, nil)
}

func (h *PassportHandler) Logout(ctx context.Context, c *app.RequestContext) {
	userIDVal, _ := c.Get("user_id")
	currentToken := string(c.Request.Header.Peek("Authorization"))
	if userID, ok := userIDVal.(uint); ok {
		_, prefix := tokenutil.GetTokenConfig("user")
		if prefix == "" {
			prefix = "user_token:"
		}
		if rd.RDB != nil && currentToken != "" {
			redisCtx, cancel := rd.OperationContext(ctx)
			defer cancel()
			idStr := strconv.Itoa(int(userID))
			rd.RDB.Del(redisCtx, prefix+"a:"+currentToken)
			rd.RDB.SRem(redisCtx, prefix+"s:"+idStr, currentToken)
			if count, _ := rd.RDB.SCard(redisCtx, prefix+"s:"+idStr).Result(); count == 0 {
				rd.RDB.Del(redisCtx, prefix+"s:"+idStr)
			}
		}
	}
	response.JSON(c, nil)
}
