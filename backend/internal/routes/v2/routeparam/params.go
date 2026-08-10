package routeparam

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

func WithQueryID(next app.HandlerFunc) app.HandlerFunc {
	return WithQueryParam("id", "id", next)
}

func WithFormID(next app.HandlerFunc) app.HandlerFunc {
	return WithFormParam("id", "id", next)
}

func WithQueryParam(name, routeParam string, next app.HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if value := c.Param(routeParam); value != "" {
			c.QueryArgs().Set(name, value)
		}
		next(ctx, c)
	}
}

func WithFormParam(name, routeParam string, next app.HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if value := c.Param(routeParam); value != "" {
			c.PostArgs().Set(name, value)
		}
		next(ctx, c)
	}
}

func WithFormParams(params map[string]string, next app.HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		for name, routeParam := range params {
			if value := c.Param(routeParam); value != "" {
				c.PostArgs().Set(name, value)
			}
		}
		next(ctx, c)
	}
}

func WithBodyOrFormID(next app.HandlerFunc) app.HandlerFunc {
	return WithBodyOrFormParam("id", "id", next)
}

func WithBodyOrFormParam(name, routeParam string, next app.HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		value := c.Param(routeParam)
		if value != "" {
			c.QueryArgs().Set(name, value)
			c.PostArgs().Set(name, value)
			injectJSONParam(c, name, value)
		}
		next(ctx, c)
	}
}

func injectJSONParam(c *app.RequestContext, name, value string) {
	contentType := strings.ToLower(string(c.Request.Header.ContentType()))
	if !strings.Contains(contentType, "json") {
		return
	}
	raw, err := c.Body()
	if err != nil || len(strings.TrimSpace(string(raw))) == 0 {
		return
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return
	}
	if n, err := strconv.ParseUint(value, 10, 64); err == nil {
		payload[name] = n
	} else {
		payload[name] = value
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	c.Request.SetBody(body)
}
