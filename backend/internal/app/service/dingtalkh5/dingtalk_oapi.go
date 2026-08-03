package dingtalkh5

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultDingTalkOAPIBaseURL = "https://oapi.dingtalk.com"
	dingTalkOAPITimeout        = 8 * time.Second
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
	SendWorkNotificationContext(ctx context.Context, config DingTalkH5CorpConfig, userIDs []string, content string) error
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

func (client defaultDingTalkIdentityClient) SendWorkNotificationContext(ctx context.Context, config DingTalkH5CorpConfig, userIDs []string, content string) error {
	appKey := strings.TrimSpace(config.AppKey)
	appSecret := strings.TrimSpace(config.AppSecret)
	if appKey == "" || appSecret == "" {
		return fmt.Errorf("请先配置钉钉 H5 AppKey 和 AppSecret")
	}
	agentID, err := strconv.ParseInt(strings.TrimSpace(config.AgentID), 10, 64)
	if err != nil || agentID <= 0 {
		return fmt.Errorf("请先配置钉钉内部应用 AgentId")
	}
	recipients := normalizeDingTalkUserIDs(userIDs)
	if len(recipients) == 0 {
		return fmt.Errorf("钉钉通知接收人不能为空")
	}
	content = strings.TrimSpace(content)
	if content == "" {
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
		"msg": map[string]interface{}{
			"msgtype": "text",
			"text": map[string]string{
				"content": content,
			},
		},
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
		return fmt.Errorf("调用钉钉接口失败：HTTP %d", resp.StatusCode)
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

func dingTalkErrorMessage(message string, code int) string {
	message = strings.TrimSpace(message)
	if message != "" {
		return message
	}
	return fmt.Sprintf("errcode=%d", code)
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
