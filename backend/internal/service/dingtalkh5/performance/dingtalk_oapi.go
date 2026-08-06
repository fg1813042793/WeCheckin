package performance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultDingTalkOAPIBaseURL    = "https://oapi.dingtalk.com"
	defaultDingTalkOpenAPIBaseURL = "https://api.dingtalk.com"
	dingTalkOAPITimeout           = 8 * time.Second
)

type DingTalkUserIdentity struct {
	UserID  string
	UnionID string
	Name    string
}

type DingTalkIdentityClient interface {
	ExchangeAuthCodeContext(ctx context.Context, config DingTalkH5CorpConfig, authCode string) (DingTalkUserIdentity, error)
}

type DingTalkWorkNotificationClient interface {
	SendWorkNotificationContext(ctx context.Context, config DingTalkH5CorpConfig, userIDs []string, notification DingTalkWorkNotificationPayload) error
}

type DingTalkWorkNotificationPayload struct {
	Title      string
	Content    string
	URL        string
	SourceName string
	PicURL     string
}

type defaultDingTalkIdentityClient struct {
	baseURL    string
	httpClient *http.Client
}

func (client defaultDingTalkIdentityClient) ExchangeAuthCodeContext(ctx context.Context, config DingTalkH5CorpConfig, authCode string) (DingTalkUserIdentity, error) {
	authCode = strings.TrimSpace(authCode)
	if authCode == "" {
		return DingTalkUserIdentity{}, fmt.Errorf("免登授权码不能为空")
	}
	appKey := strings.TrimSpace(config.AppKey)
	appSecret := strings.TrimSpace(config.AppSecret)
	if appKey == "" || appSecret == "" {
		return DingTalkUserIdentity{}, fmt.Errorf("请先配置钉钉 H5 AppKey 和 AppSecret")
	}
	accessToken, err := client.accessTokenContext(ctx, appKey, appSecret)
	if err != nil {
		return DingTalkUserIdentity{}, err
	}
	return client.userInfoContext(ctx, accessToken, authCode)
}

func (client defaultDingTalkIdentityClient) SendWorkNotificationContext(ctx context.Context, config DingTalkH5CorpConfig, userIDs []string, notification DingTalkWorkNotificationPayload) error {
	appKey := strings.TrimSpace(config.AppKey)
	appSecret := strings.TrimSpace(config.AppSecret)
	if appKey == "" || appSecret == "" {
		return fmt.Errorf("请先配置钉钉 H5 AppKey 和 AppSecret")
	}
	notification = normalizeDingTalkWorkNotificationPayload(notification)
	notifyMode := normalizeDingTalkH5NotifyMode(config.NotifyMode, config.AgentID, config.RobotCode)
	if notifyMode == "robot" {
		return client.sendRobotWorkNotificationContext(ctx, appKey, appSecret, config.RobotCode, userIDs, notification)
	}
	err := client.sendAgentWorkNotificationContext(ctx, appKey, appSecret, config.AgentID, userIDs, notification)
	if err == nil {
		return nil
	}
	if notifyMode == "agent_fallback" && shouldFallbackDingTalkAgentNotificationToRobot(err) {
		if fallbackErr := client.sendRobotWorkNotificationContext(ctx, appKey, appSecret, config.RobotCode, userIDs, notification); fallbackErr != nil {
			return fmt.Errorf("%v；已尝试新版机器人通知但失败：%w", err, fallbackErr)
		}
		return nil
	}
	return err
}

func shouldFallbackDingTalkAgentNotificationToRobot(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	if !strings.Contains(message, "agentid") {
		return false
	}
	return strings.Contains(message, "不合法") || strings.Contains(message, "invalid")
}

func (client defaultDingTalkIdentityClient) sendAgentWorkNotificationContext(ctx context.Context, appKey, appSecret, rawAgentID string, userIDs []string, notification DingTalkWorkNotificationPayload) error {
	agentID, err := strconv.ParseInt(strings.TrimSpace(rawAgentID), 10, 64)
	if err != nil || agentID <= 0 {
		return fmt.Errorf("请先配置钉钉内部应用 AgentId")
	}
	recipients := normalizeDingTalkUserIDs(userIDs)
	if len(recipients) == 0 {
		return fmt.Errorf("钉钉通知接收人不能为空")
	}
	notification = normalizeDingTalkWorkNotificationPayload(notification)
	if notification.Content == "" {
		return fmt.Errorf("钉钉通知内容不能为空")
	}
	accessToken, err := client.accessTokenContext(ctx, appKey, appSecret)
	if err != nil {
		return err
	}
	params := url.Values{}
	params.Set("access_token", accessToken)
	endpoint := strings.TrimRight(client.apiBaseURL(), "/") + "/topapi/message/corpconversation/asyncsend_v2?" + params.Encode()
	body, _ := json.Marshal(map[string]interface{}{
		"agent_id":    agentID,
		"userid_list": strings.Join(recipients, ","),
		"msg":         dingTalkAgentNotificationMessage(notification),
	})
	var payload struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		TaskID  int64  `json:"task_id"`
	}
	if err := client.doJSONContext(ctx, http.MethodPost, endpoint, body, &payload); err != nil {
		return err
	}
	if payload.ErrCode != 0 {
		return fmt.Errorf("发送钉钉工作通知失败：%s", dingTalkErrorMessage(payload.ErrMsg, payload.ErrCode))
	}
	return nil
}

func (client defaultDingTalkIdentityClient) sendRobotWorkNotificationContext(ctx context.Context, appKey, appSecret, rawRobotCode string, userIDs []string, notification DingTalkWorkNotificationPayload) error {
	robotCode := strings.TrimSpace(rawRobotCode)
	if robotCode == "" {
		robotCode = appKey
	}
	recipients := normalizeDingTalkUserIDs(userIDs)
	if len(recipients) == 0 {
		return fmt.Errorf("钉钉通知接收人不能为空")
	}
	notification = normalizeDingTalkWorkNotificationPayload(notification)
	if notification.Content == "" {
		return fmt.Errorf("钉钉通知内容不能为空")
	}
	accessToken, err := client.openAPIAccessTokenContext(ctx, appKey, appSecret)
	if err != nil {
		return err
	}
	msgKey, msgParam := dingTalkRobotNotificationMessage(notification)
	endpoint := strings.TrimRight(client.openAPIBaseURL(), "/") + "/v1.0/robot/oToMessages/batchSend"
	body, _ := json.Marshal(map[string]interface{}{
		"robotCode": robotCode,
		"userIds":   recipients,
		"msgKey":    msgKey,
		"msgParam":  msgParam,
	})
	var payload struct {
		Code            string `json:"code"`
		Message         string `json:"message"`
		ProcessQueryKey string `json:"processQueryKey"`
	}
	if err := client.doDingTalkJSONContext(ctx, http.MethodPost, endpoint, body, accessToken, &payload); err != nil {
		return err
	}
	if strings.TrimSpace(payload.Code) != "" {
		return fmt.Errorf("发送钉钉机器人通知失败：%s", strings.TrimSpace(payload.Message))
	}
	return nil
}

func normalizeDingTalkWorkNotificationPayload(notification DingTalkWorkNotificationPayload) DingTalkWorkNotificationPayload {
	notification.Title = strings.TrimSpace(notification.Title)
	notification.Content = strings.TrimSpace(notification.Content)
	notification.URL = strings.TrimSpace(notification.URL)
	notification.SourceName = strings.TrimSpace(notification.SourceName)
	notification.PicURL = strings.TrimSpace(notification.PicURL)
	if notification.Title == "" && notification.URL != "" {
		notification.Title = "绩效流程待办"
	}
	return notification
}

func dingTalkAgentNotificationMessage(notification DingTalkWorkNotificationPayload) map[string]interface{} {
	if notification.URL != "" {
		headText := strings.TrimSpace(notification.SourceName)
		if headText == "" {
			headText = notification.Title
		}
		body := map[string]string{
			"title":   notification.Title,
			"content": notification.Content,
		}
		if notification.PicURL != "" {
			body["image"] = notification.PicURL
		}
		return map[string]interface{}{
			"msgtype": "oa",
			"oa": map[string]interface{}{
				"message_url":    notification.URL,
				"pc_message_url": notification.URL,
				"head": map[string]string{
					"bgcolor": "FF1677FF",
					"text":    headText,
				},
				"body": body,
			},
		}
	}
	return map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": notification.Content,
		},
	}
}

func dingTalkRobotNotificationMessage(notification DingTalkWorkNotificationPayload) (string, string) {
	if notification.URL != "" {
		params := map[string]string{
			"title":      notification.Title,
			"text":       notification.Content,
			"messageUrl": notification.URL,
			"picUrl":     notification.PicURL,
		}
		if notification.SourceName != "" {
			params["sourceName"] = notification.SourceName
		}
		msgParam, _ := json.Marshal(params)
		return "sampleLink", string(msgParam)
	}
	msgParam, _ := json.Marshal(map[string]string{"content": notification.Content})
	return "sampleText", string(msgParam)
}

func (client defaultDingTalkIdentityClient) accessTokenContext(ctx context.Context, appKey, appSecret string) (string, error) {
	params := url.Values{}
	params.Set("appkey", appKey)
	params.Set("appsecret", appSecret)
	endpoint := strings.TrimRight(client.apiBaseURL(), "/") + "/gettoken?" + params.Encode()

	var payload struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
	}
	if err := client.doJSONContext(ctx, http.MethodGet, endpoint, nil, &payload); err != nil {
		return "", err
	}
	if payload.ErrCode != 0 {
		return "", fmt.Errorf("获取钉钉访问凭证失败：%s", dingTalkErrorMessage(payload.ErrMsg, payload.ErrCode))
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return "", fmt.Errorf("钉钉访问凭证为空")
	}
	return strings.TrimSpace(payload.AccessToken), nil
}

func (client defaultDingTalkIdentityClient) openAPIAccessTokenContext(ctx context.Context, appKey, appSecret string) (string, error) {
	endpoint := strings.TrimRight(client.openAPIBaseURL(), "/") + "/v1.0/oauth2/accessToken"
	body, _ := json.Marshal(map[string]string{
		"appKey":    appKey,
		"appSecret": appSecret,
	})
	var payload struct {
		AccessToken string `json:"accessToken"`
		ExpireIn    int64  `json:"expireIn"`
		Code        string `json:"code"`
		Message     string `json:"message"`
	}
	if err := client.doJSONContext(ctx, http.MethodPost, endpoint, body, &payload); err != nil {
		return "", err
	}
	if strings.TrimSpace(payload.Code) != "" {
		return "", fmt.Errorf("获取钉钉新版访问凭证失败：%s", strings.TrimSpace(payload.Message))
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return "", fmt.Errorf("钉钉新版访问凭证为空")
	}
	return strings.TrimSpace(payload.AccessToken), nil
}

func (client defaultDingTalkIdentityClient) userInfoContext(ctx context.Context, accessToken, authCode string) (DingTalkUserIdentity, error) {
	params := url.Values{}
	params.Set("access_token", accessToken)
	endpoint := strings.TrimRight(client.apiBaseURL(), "/") + "/topapi/v2/user/getuserinfo?" + params.Encode()

	body, _ := json.Marshal(map[string]string{"code": authCode})
	var payload struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		Result  struct {
			UserID            string `json:"userid"`
			UnionID           string `json:"unionid"`
			AssociatedUnionID string `json:"associated_unionid"`
			Name              string `json:"name"`
		} `json:"result"`
	}
	if err := client.doJSONContext(ctx, http.MethodPost, endpoint, body, &payload); err != nil {
		return DingTalkUserIdentity{}, err
	}
	if payload.ErrCode != 0 {
		return DingTalkUserIdentity{}, fmt.Errorf("获取钉钉用户身份失败：%s", dingTalkErrorMessage(payload.ErrMsg, payload.ErrCode))
	}
	unionID := strings.TrimSpace(payload.Result.UnionID)
	if unionID == "" {
		unionID = strings.TrimSpace(payload.Result.AssociatedUnionID)
	}
	return DingTalkUserIdentity{
		UserID:  strings.TrimSpace(payload.Result.UserID),
		UnionID: unionID,
		Name:    strings.TrimSpace(payload.Result.Name),
	}, nil
}

func (client defaultDingTalkIdentityClient) doJSONContext(ctx context.Context, method, endpoint string, body []byte, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.http().Do(req)
	if err != nil {
		return fmt.Errorf("调用钉钉接口失败：%w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return dingTalkHTTPStatusError(resp.StatusCode, resp.Body)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("解析钉钉接口响应失败：%w", err)
	}
	return nil
}

func (client defaultDingTalkIdentityClient) doDingTalkJSONContext(ctx context.Context, method, endpoint string, body []byte, accessToken string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(accessToken) != "" {
		req.Header.Set("x-acs-dingtalk-access-token", strings.TrimSpace(accessToken))
	}
	resp, err := client.http().Do(req)
	if err != nil {
		return fmt.Errorf("调用钉钉接口失败：%w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return dingTalkHTTPStatusError(resp.StatusCode, resp.Body)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("解析钉钉接口响应失败：%w", err)
	}
	return nil
}

func (client defaultDingTalkIdentityClient) http() *http.Client {
	if client.httpClient != nil {
		return client.httpClient
	}
	return &http.Client{Timeout: dingTalkOAPITimeout}
}

func (client defaultDingTalkIdentityClient) apiBaseURL() string {
	if strings.TrimSpace(client.baseURL) != "" {
		return strings.TrimSpace(client.baseURL)
	}
	return defaultDingTalkOAPIBaseURL
}

func (client defaultDingTalkIdentityClient) openAPIBaseURL() string {
	if strings.TrimSpace(client.baseURL) != "" {
		return strings.TrimSpace(client.baseURL)
	}
	return defaultDingTalkOpenAPIBaseURL
}

func dingTalkErrorMessage(message string, code int) string {
	message = strings.TrimSpace(message)
	if message != "" {
		return message
	}
	return fmt.Sprintf("errcode=%d", code)
}

func dingTalkHTTPStatusError(statusCode int, body io.Reader) error {
	message := ""
	if body != nil {
		data, _ := io.ReadAll(io.LimitReader(body, 2048))
		message = strings.TrimSpace(string(data))
	}
	if message == "" {
		return fmt.Errorf("调用钉钉接口失败：HTTP %d", statusCode)
	}
	return fmt.Errorf("调用钉钉接口失败：HTTP %d：%s", statusCode, message)
}

func normalizeDingTalkUserIDs(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
