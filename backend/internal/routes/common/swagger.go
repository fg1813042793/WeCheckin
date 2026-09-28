package common

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/swagger"
	swaggerFiles "github.com/swaggo/files"
	"github.com/swaggo/swag"
)

var openAPIVersionPattern = regexp.MustCompile(`^3\.0\.\d+$`)

type swaggerDefinitionVersion struct {
	Swagger string `json:"swagger"`
	OpenAPI string `json:"openapi"`
}

func RegisterSwagger(h *server.Hertz) {
	url := swagger.URL("/swagger/doc.json")
	swaggerHandler := swagger.WrapHandler(swaggerFiles.Handler, url)
	h.GET("/swagger", func(ctx context.Context, c *app.RequestContext) {
		c.Redirect(302, []byte("/swagger/index.html"))
	})
	h.GET("/swagger/doc.json", func(ctx context.Context, c *app.RequestContext) {
		serveSwaggerDocument(c)
	})
	h.GET("/swagger/*any", func(ctx context.Context, c *app.RequestContext) {
		if string(c.Path()) == "/swagger/index.html" {
			serveSwaggerIndex(c)
			return
		}
		swaggerHandler(ctx, c)
	})
}

func serveSwaggerDocument(c *app.RequestContext) {
	document, err := swag.ReadDoc()
	if err != nil || !validSwaggerDefinition(document) {
		c.Header("Cache-Control", "no-store")
		c.Data(http.StatusInternalServerError, "application/json; charset=utf-8", []byte(`{"error":"Swagger definition is invalid"}`))
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(document))
}

func validSwaggerDefinition(document string) bool {
	var version swaggerDefinitionVersion
	if err := json.Unmarshal([]byte(document), &version); err != nil {
		return false
	}
	return version.Swagger == "2.0" || openAPIVersionPattern.MatchString(version.OpenAPI)
}
