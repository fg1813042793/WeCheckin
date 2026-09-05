package config

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

const (
	DingTalkMessageTypeAuto       = "auto"
	DingTalkMessageTypeText       = "text"
	DingTalkMessageTypeImage      = "image"
	DingTalkMessageTypeVoice      = "voice"
	DingTalkMessageTypeFile       = "file"
	DingTalkMessageTypeLink       = "link"
	DingTalkMessageTypeOA         = "oa"
	DingTalkMessageTypeMarkdown   = "markdown"
	DingTalkMessageTypeActionCard = "action_card"
)

type DingTalkWorkNotificationPayload struct {
	MessageType string
	Title       string
	Content     string
	URL         string
	SourceName  string
	PicURL      string
	MediaID     string
	Duration    int
	ButtonTitle string
	HeadColor   string
}

type defaultDingTalkIdentityClient struct {
	baseURL    string
	httpClient *http.Client
}

func DefaultDingTalkIdentityClient() DingTalkIdentityClient {
	return defaultDingTalkIdentityClient{}
}

func DefaultDingTalkWorkNotificationClient() DingTalkWorkNotificationClient {
	return defaultDingTalkIdentityClient{}
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
	if err := validateDingTalkWorkNotificationPayload(notification); err != nil {
		return err
	}
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
	if err := validateDingTalkWorkNotificationPayload(notification); err != nil {
		return err
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
	if err := validateDingTalkRobotNotificationPayload(notification); err != nil {
		return err
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
	notification.MessageType = strings.ToLower(strings.TrimSpace(notification.MessageType))
	notification.Title = strings.TrimSpace(notification.Title)
	notification.Content = strings.TrimSpace(notification.Content)
	notification.URL = strings.TrimSpace(notification.URL)
	notification.SourceName = strings.TrimSpace(notification.SourceName)
	notification.PicURL = strings.TrimSpace(notification.PicURL)
	notification.MediaID = strings.TrimSpace(notification.MediaID)
	notification.ButtonTitle = strings.TrimSpace(notification.ButtonTitle)
	notification.HeadColor = strings.ToUpper(strings.TrimPrefix(strings.TrimSpace(notification.HeadColor), "#"))
	if notification.MessageType == "" {
		notification.MessageType = DingTalkMessageTypeAuto
	}
	if notification.ButtonTitle == "" {
		notification.ButtonTitle = "查看流程"
	}
	if notification.HeadColor == "" {
		notification.HeadColor = "FF1677FF"
	} else if len(notification.HeadColor) == 6 {
		notification.HeadColor = "FF" + notification.HeadColor
	}
	if notification.Title == "" && notification.URL != "" {
		notification.Title = "绩效流程待办"
	}
	return notification
}

func dingTalkAgentNotificationMessage(notification DingTalkWorkNotificationPayload) map[string]interface{} {
	notification = normalizeDingTalkWorkNotificationPayload(notification)
	switch notification.MessageType {
	case DingTalkMessageTypeText:
		return dingTalkTextMessage(notification.Content)
	case DingTalkMessageTypeImage:
		return map[string]interface{}{"msgtype": "image", "image": map[string]string{"media_id": notification.MediaID}}
	case DingTalkMessageTypeVoice:
		return map[string]interface{}{"msgtype": "voice", "voice": map[string]string{
			"media_id": notification.MediaID, "duration": strconv.Itoa(notification.Duration),
		}}
	case DingTalkMessageTypeFile:
		return map[string]interface{}{"msgtype": "file", "file": map[string]string{"media_id": notification.MediaID}}
	case DingTalkMessageTypeLink:
		return map[string]interface{}{"msgtype": "link", "link": map[string]string{
			"title": notification.Title, "text": notification.Content, "message_url": notification.URL, "pic_url": notification.PicURL,
		}}
	case DingTalkMessageTypeMarkdown:
		return map[string]interface{}{"msgtype": "markdown", "markdown": map[string]string{
			"title": notification.Title, "text": notification.Content,
		}}
	case DingTalkMessageTypeActionCard:
		return map[string]interface{}{
			"msgtype": DingTalkMessageTypeActionCard,
			"action_card": map[string]string{
				"title":        notification.Title,
				"markdown":     notification.Content,
				"single_title": notification.ButtonTitle,
				"single_url":   notification.URL,
			},
		}
	case DingTalkMessageTypeOA:
		return dingTalkOAMessage(notification)
	}
	if notification.URL != "" {
		return dingTalkOAMessage(notification)
	}
	return dingTalkTextMessage(notification.Content)
}

func dingTalkTextMessage(content string) map[string]interface{} {
	return map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]string{"content": content},
	}
}

func dingTalkOAMessage(notification DingTalkWorkNotificationPayload) map[string]interface{} {
	headText := strings.TrimSpace(notification.SourceName)
	if headText == "" {
		headText = notification.Title
	}
	body := map[string]string{"title": notification.Title, "content": notification.Content}
	if notification.PicURL != "" {
		body["image"] = notification.PicURL
	}
	return map[string]interface{}{
		"msgtype": "oa",
		"oa": map[string]interface{}{
			"message_url": notification.URL, "pc_message_url": notification.URL,
			"head": map[string]string{"bgcolor": notification.HeadColor, "text": headText},
			"body": body,
		},
	}
}

func dingTalkRobotNotificationMessage(notification DingTalkWorkNotificationPayload) (string, string) {
	notification = normalizeDingTalkWorkNotificationPayload(notification)
	if notification.MessageType == DingTalkMessageTypeActionCard {
		msgParam, _ := json.Marshal(map[string]string{
			"title":       notification.Title,
			"text":        notification.Content,
			"singleTitle": notification.ButtonTitle,
			"singleURL":   notification.URL,
		})
		return "sampleActionCard", string(msgParam)
	}
	if notification.MessageType == DingTalkMessageTypeMarkdown {
		msgParam, _ := json.Marshal(map[string]string{"title": notification.Title, "text": notification.Content})
		return "sampleMarkdown", string(msgParam)
	}
	if notification.MessageType == DingTalkMessageTypeLink || notification.MessageType == DingTalkMessageTypeOA || notification.URL != "" {
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

func validateDingTalkWorkNotificationPayload(notification DingTalkWorkNotificationPayload) error {
	notification = normalizeDingTalkWorkNotificationPayload(notification)
	switch notification.MessageType {
	case DingTalkMessageTypeAuto, DingTalkMessageTypeText:
		if notification.Content == "" {
			return fmt.Errorf("钉钉文本通知内容不能为空")
		}
	case DingTalkMessageTypeImage, DingTalkMessageTypeFile:
		if notification.MediaID == "" {
			return fmt.Errorf("钉钉%s消息 mediaId 不能为空", notification.MessageType)
		}
	case DingTalkMessageTypeVoice:
		if notification.MediaID == "" || notification.Duration <= 0 {
			return fmt.Errorf("钉钉语音消息 mediaId 和时长不能为空")
		}
	case DingTalkMessageTypeLink, DingTalkMessageTypeOA:
		if notification.Title == "" || notification.Content == "" || notification.URL == "" {
			return fmt.Errorf("钉钉%s消息标题、正文和跳转地址不能为空", notification.MessageType)
		}
	case DingTalkMessageTypeActionCard:
		if notification.URL == "" {
			return fmt.Errorf("ActionCard 跳转地址不能为空")
		}
		if notification.Title == "" || notification.Content == "" {
			return fmt.Errorf("钉钉 ActionCard 消息标题和正文不能为空")
		}
	case DingTalkMessageTypeMarkdown:
		if notification.Title == "" || notification.Content == "" {
			return fmt.Errorf("钉钉 Markdown 消息标题和正文不能为空")
		}
	default:
		return fmt.Errorf("不支持的钉钉消息类型：%s", notification.MessageType)
	}
	return nil
}

func validateDingTalkRobotNotificationPayload(notification DingTalkWorkNotificationPayload) error {
	if err := validateDingTalkWorkNotificationPayload(notification); err != nil {
		return err
	}
	switch normalizeDingTalkWorkNotificationPayload(notification).MessageType {
	case DingTalkMessageTypeImage, DingTalkMessageTypeVoice, DingTalkMessageTypeFile:
		return fmt.Errorf("当前钉钉机器人单聊通道不支持 %s 消息，请使用内部应用通知模式", notification.MessageType)
	default:
		return nil
	}
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
