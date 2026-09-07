package common

import (
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

const swaggerIndexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>WeCheckin API</title>
  <link rel="stylesheet" href="./swagger-ui.css">
  <link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32">
  <style>
    html { box-sizing: border-box; overflow-y: scroll; }
    *, *::before, *::after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="./swagger-ui-bundle.js"></script>
  <script src="./swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: "/swagger/doc.json",
        dom_id: "#swagger-ui",
        validatorUrl: null,
        oauth2RedirectUrl: window.location.origin + window.location.pathname.replace(/index\.html$/, "oauth2-redirect.html"),
        persistAuthorization: false,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout",
        docExpansion: "list",
        deepLinking: true,
        defaultModelsExpandDepth: 1,
        filter: true
      })
    }
  </script>
</body>
</html>
`

func serveSwaggerIndex(c *app.RequestContext) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerIndexHTML))
}
