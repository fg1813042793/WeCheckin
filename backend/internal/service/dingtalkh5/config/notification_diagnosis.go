package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

type DingTalkH5NotificationDiagnosis struct {
	Success         bool                                  `json:"success"`
	CheckedAt       string                                `json:"checkedAt"`
	CorpID          string                                `json:"corpId"`
	CorpName        string                                `json:"corpName"`
	NotifyEnabled   int                                   `json:"notifyEnabled"`
	NotifyMode      string                                `json:"notifyMode"`
	AppKeyMasked    string                                `json:"appKeyMasked"`
	AppSecretSet    bool                                  `json:"appSecretSet"`
	AgentID         string                                `json:"agentId"`
	UnifiedAppID    string                                `json:"unifiedAppId"`
	RobotCodeMasked string                                `json:"robotCodeMasked"`
	RecipientUserID string                                `json:"recipientUserId"`
	Conclusion      string                                `json:"conclusion"`
	Steps           []DingTalkH5NotificationDiagnosisStep `json:"steps"`
}

type DingTalkH5NotificationDiagnosisStep struct {
	Name       string                 `json:"name"`
	Status     string                 `json:"status"`
	Method     string                 `json:"method,omitempty"`
	Endpoint   string                 `json:"endpoint,omitempty"`
	DurationMs int64                  `json:"durationMs,omitempty"`
	Request    map[string]interface{} `json:"request,omitempty"`
	Response   map[string]interface{} `json:"response,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

// DiagnoseDingTalkH5WorkNotificationContext performs the same legacy work-notification
// token and asyncsend_v2 path used by AgentId/OA mode, but returns sanitized evidence.
func DiagnoseDingTalkH5WorkNotificationContext(ctx context.Context, corpID, recipientUserID string) (DingTalkH5NotificationDiagnosis, error) {
	corpID = strings.TrimSpace(corpID)
	config, err := loadDingTalkH5CorpConfigContext(ctx, corpID)
	if err != nil {
		return DingTalkH5NotificationDiagnosis{}, err
	}
	config = normalizeDingTalkH5CorpConfig(config)
	recipientUserID, err = diagnosisRecipientForCorpContext(ctx, config.CorpID, recipientUserID)
	if err != nil {
		return DingTalkH5NotificationDiagnosis{}, err
	}

	result := DingTalkH5NotificationDiagnosis{
		CheckedAt:       time.Now().Format(time.RFC3339),
		CorpID:          config.CorpID,
		CorpName:        config.CorpName,
		NotifyEnabled:   config.NotifyEnabled,
		NotifyMode:      config.NotifyMode,
		AppKeyMasked:    dingTalkH5MaskLogValue(config.AppKey),
		AppSecretSet:    config.AppSecretSet,
		AgentID:         config.AgentID,
		UnifiedAppID:    config.UnifiedAppID,
		RobotCodeMasked: dingTalkH5MaskLogValue(config.RobotCode),
		RecipientUserID: recipientUserID,
	}
	result.Steps = append(result.Steps, dingTalkH5DiagnosisConfigStep(config, recipientUserID))

	client := defaultDingTalkIdentityClient{}
	accessToken, tokenStep := client.diagnoseOldAccessTokenContext(ctx, config.AppKey, config.AppSecret)
	result.Steps = append(result.Steps, tokenStep)
	if tokenStep.Status != "success" {
		result.Conclusion = dingTalkH5NotificationDiagnosisConclusion(result)
		return result, nil
	}

	sendStep := client.diagnoseOldAgentWorkNotificationContext(ctx, accessToken, config.AgentID, recipientUserID)
	result.Steps = append(result.Steps, sendStep)
	result.Success = sendStep.Status == "success"
	result.Conclusion = dingTalkH5NotificationDiagnosisConclusion(result)
	return result, nil
}

func diagnosisRecipientForCorpContext(ctx context.Context, corpID, recipientUserID string) (string, error) {
	recipientUserID = strings.TrimSpace(recipientUserID)
	if recipientUserID != "" {
		return recipientUserID, nil
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return "", fmt.Errorf("database is not initialized")
	}

	var binding model.DingTalkH5UserBinding
	err := db.Where("`corp_id` = ? AND `enabled` = 1 AND `dingtalk_user_id` <> ''", strings.TrimSpace(corpID)).
		Order("`id` ASC").
		First(&binding).Error
	if err == nil {
		return strings.TrimSpace(binding.DingTalkUserID), nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	return "", fmt.Errorf("请先在钉钉用户绑定中配置至少一个可用接收人")
}

func dingTalkH5DiagnosisConfigStep(config DingTalkH5CorpConfig, recipientUserID string) DingTalkH5NotificationDiagnosisStep {
	return DingTalkH5NotificationDiagnosisStep{
		Name:   "读取后台企业应用配置",
		Status: "success",
		Request: map[string]interface{}{
			"corpId": config.CorpID,
		},
		Response: map[string]interface{}{
			"corpId":          config.CorpID,
			"corpName":        config.CorpName,
			"notifyMode":      config.NotifyMode,
			"notifyEnabled":   config.NotifyEnabled,
			"appKey":          dingTalkH5MaskLogValue(config.AppKey),
			"appSecretSet":    config.AppSecretSet,
			"agentId":         config.AgentID,
			"unifiedAppId":    config.UnifiedAppID,
			"robotCode":       dingTalkH5MaskLogValue(config.RobotCode),
			"recipientUserId": recipientUserID,
		},
	}
}

func (client defaultDingTalkIdentityClient) diagnoseOldAccessTokenContext(ctx context.Context, appKey, appSecret string) (string, DingTalkH5NotificationDiagnosisStep) {
	params := url.Values{}
	params.Set("appkey", strings.TrimSpace(appKey))
	params.Set("appsecret", strings.TrimSpace(appSecret))
	endpoint := strings.TrimRight(client.apiBaseURL(), "/") + "/gettoken?" + params.Encode()
	step := DingTalkH5NotificationDiagnosisStep{
		Name:     "获取旧版 access_token",
		Method:   http.MethodGet,
		Endpoint: strings.TrimRight(client.apiBaseURL(), "/") + "/gettoken?",
		Request: map[string]interface{}{
			"appkey":    dingTalkH5MaskLogValue(appKey),
			"appsecret": "***",
		},
	}
	if strings.TrimSpace(appKey) == "" || strings.TrimSpace(appSecret) == "" {
		step.Status = "failed"
		step.Error = "请先配置钉钉 H5 AppKey 和 AppSecret"
		return "", step
	}

	start := time.Now()
	var payload struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	err := client.doJSONContext(ctx, http.MethodGet, endpoint, nil, &payload)
	step.DurationMs = time.Since(start).Milliseconds()
	if err != nil {
		step.Status = "failed"
		step.Error = err.Error()
		return "", step
	}
	step.Response = map[string]interface{}{
		"errcode":             payload.ErrCode,
		"errmsg":              payload.ErrMsg,
		"accessTokenReceived": strings.TrimSpace(payload.AccessToken) != "",
		"expiresIn":           payload.ExpiresIn,
	}
	if payload.ErrCode != 0 {
		step.Status = "failed"
		step.Error = fmt.Sprintf("获取钉钉访问凭证失败：%s", dingTalkErrorMessage(payload.ErrMsg, payload.ErrCode))
		return "", step
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		step.Status = "failed"
		step.Error = "钉钉访问凭证为空"
		return "", step
	}
	step.Status = "success"
	return strings.TrimSpace(payload.AccessToken), step
}

func (client defaultDingTalkIdentityClient) diagnoseOldAgentWorkNotificationContext(ctx context.Context, accessToken, rawAgentID, recipientUserID string) DingTalkH5NotificationDiagnosisStep {
	step := DingTalkH5NotificationDiagnosisStep{
		Name:     "发送旧版工作通知 asyncsend_v2",
		Method:   http.MethodPost,
		Endpoint: strings.TrimRight(client.apiBaseURL(), "/") + "/topapi/message/corpconversation/asyncsend_v2?",
		Request: map[string]interface{}{
			"userid_list": strings.TrimSpace(recipientUserID),
			"msgtype":     "text",
		},
	}
	agentID, err := strconv.ParseInt(strings.TrimSpace(rawAgentID), 10, 64)
	if err != nil || agentID <= 0 {
		step.Status = "failed"
		step.Error = "请先配置钉钉内部应用 AgentId"
		return step
	}
	if strings.TrimSpace(recipientUserID) == "" {
		step.Status = "failed"
		step.Error = "钉钉通知接收人不能为空"
		return step
	}
	step.Request["agent_id"] = agentID

	params := url.Values{}
	params.Set("access_token", strings.TrimSpace(accessToken))
	endpoint := strings.TrimRight(client.apiBaseURL(), "/") + "/topapi/message/corpconversation/asyncsend_v2?" + params.Encode()
	body, _ := json.Marshal(map[string]interface{}{
		"agent_id":    agentID,
		"userid_list": strings.TrimSpace(recipientUserID),
		"msg": map[string]interface{}{
			"msgtype": "text",
			"text": map[string]string{
				"content": "WeCheckin 钉钉通知诊断测试，请忽略。",
			},
		},
	})
	var payload struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		TaskID  int64  `json:"task_id"`
	}
	start := time.Now()
	err = client.doJSONContext(ctx, http.MethodPost, endpoint, body, &payload)
	step.DurationMs = time.Since(start).Milliseconds()
	if err != nil {
		step.Status = "failed"
		step.Error = err.Error()
		return step
	}
	step.Response = map[string]interface{}{
		"errcode": payload.ErrCode,
		"errmsg":  payload.ErrMsg,
		"taskId":  payload.TaskID,
	}
	if payload.ErrCode != 0 {
		step.Status = "failed"
		step.Error = fmt.Sprintf("发送钉钉工作通知失败：%s", dingTalkErrorMessage(payload.ErrMsg, payload.ErrCode))
		return step
	}
	step.Status = "success"
	return step
}

func dingTalkH5NotificationDiagnosisConclusion(result DingTalkH5NotificationDiagnosis) string {
	if result.Success {
		return "旧版工作通知链路可用：access_token 获取成功，asyncsend_v2 发送成功。"
	}
	for _, step := range result.Steps {
		if step.Status != "failed" {
			continue
		}
		message := step.Error
		if message == "" {
			message = "钉钉接口返回失败"
		}
		if strings.Contains(strings.ToLower(message), "agentid") && strings.Contains(message, "不合法") {
			invalidAgentIDEvidence := fmt.Sprintf("agentId【%s】不合法", result.AgentID)
			return fmt.Sprintf("旧版 access_token 已获取成功，但 asyncsend_v2 拒绝 AgentId。当前证据=%s，钉钉返回：%s。建议带调用链路向钉钉工单确认该 Client ID/Secret 是否具备此 AgentId 的工作通知发送权限。", invalidAgentIDEvidence, message)
		}
		return fmt.Sprintf("%s 失败：%s", step.Name, message)
	}
	return "诊断未完成，请检查后台配置和钉钉用户绑定。"
}

func dingTalkH5MaskLogValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	if len(value) <= 8 {
		return strings.Repeat("*", len(value))
	}
	return value[:4] + "..." + value[len(value)-4:]
}
