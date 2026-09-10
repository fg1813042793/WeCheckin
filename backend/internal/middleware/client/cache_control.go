package client

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

const privateNoStore = "private, no-store"

func NoStore() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		c.Header("Cache-Control", privateNoStore)
		c.Next(ctx)
	}
}
