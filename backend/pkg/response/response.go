package response

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type Resp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Size     int         `json:"size"`
	Page     int         `json:"page"`
	OldTotal int64       `json:"oldTotal,omitempty"`
}

func JSON(ctx *app.RequestContext, data interface{}) {
	ctx.JSON(consts.StatusOK, Resp{Code: 0, Msg: "success", Data: data})
}

func Fail(ctx *app.RequestContext, msg string) {
	ctx.JSON(consts.StatusOK, Resp{Code: 1, Msg: msg})
}

func FailCode(ctx *app.RequestContext, code int, msg string) {
	ctx.JSON(consts.StatusOK, Resp{Code: code, Msg: msg})
}
