package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/pkg/response"
)

type GeoHandler struct{}

func NewGeoHandler() *GeoHandler {
	return &GeoHandler{}
}

type nominatimResp struct {
	DisplayName string `json:"display_name"`
	Error       string `json:"error"`
}

// ReverseGeocode 经纬度反查地址
func (h *GeoHandler) ReverseGeocode(ctx context.Context, c *app.RequestContext) {
	lat := c.Query("lat")
	lng := c.Query("lng")
	if lat == "" || lng == "" {
		response.Fail(c, "缺少参数")
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	apiURL := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?format=json&lat=%s&lon=%s&zoom=18&accept-language=zh",
		url.QueryEscape(lat), url.QueryEscape(lng))

	req, _ := http.NewRequest("get", apiURL, nil)
	req.Header.Set("User-Agent", "WeCheckin/1.0")

	resp, err := client.Do(req)
	if err != nil {
		response.Fail(c, "获取位置信息失败")
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result nominatimResp
	if err := json.Unmarshal(body, &result); err != nil || result.Error != "" {
		response.Fail(c, "解析位置信息失败")
		return
	}

	response.JSON(c, map[string]string{"address": result.DisplayName})
}
