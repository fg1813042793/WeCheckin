package survey

import (
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
)

// writeCSV 写 CSV 文件（带 UTF-8 BOM 供 Excel 识别）
func writeCSV(c *app.RequestContext, filename string, data []byte) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Length", fmt.Sprintf("%d", len(data)+3))
	_, _ = c.Response.BodyWriter().Write([]byte{0xEF, 0xBB, 0xBF})
	_, _ = c.Response.BodyWriter().Write(data)
}
