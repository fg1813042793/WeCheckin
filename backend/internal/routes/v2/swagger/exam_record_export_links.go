package swagger

// @Tags API v2-公开考试
// @Summary 使用限时凭证下载单份考试答卷
// @Description Token 由考试交卷后的通知模版变量 {answerDownloadUrl} 生成；格式、考试和答卷均绑定在 Redis 凭证中，调用方不能通过参数修改。Token 过期、答卷删除或凭证不存在时返回 410。
// @Produce application/octet-stream
// @Param token query string true "限时答卷下载凭证"
// @Success 200 {file} string
// @Failure 410 {string} string "答卷下载链接已失效"
// @Failure 500 {string} string "答卷下载失败，请稍后重试"
// @Router /api/v2/public/exam-record-exports/download [get]
func swaggerPublicExamRecordExportDownload() {}
