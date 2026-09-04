package dingtalk

type SettingsResponse struct {
	CorpID        string               `json:"corpId"`
	AppKey        string               `json:"appKey"`
	AgentID       string               `json:"agentId"`
	UnifiedAppID  string               `json:"unifiedAppId"`
	NotifyMode    string               `json:"notifyMode"`
	RobotCode     string               `json:"robotCode"`
	AppSecretSet  bool                 `json:"appSecretSet"`
	CorpConfigs   []CorpConfigResponse `json:"corpConfigs"`
	TokenExpire   string               `json:"tokenExpire"`
	RedisPrefix   string               `json:"redisPrefix"`
	SingleLogin   int                  `json:"singleLogin"`
	SelfBind      int                  `json:"selfBind"`
	NotifyEnabled int                  `json:"notifyEnabled"`
	AppName       string               `json:"appName"`
	LogoText      string               `json:"logoText"`
	LogoURL       string               `json:"logoUrl"`
	AppURL        string               `json:"appUrl"`
}

type CorpConfigResponse struct {
	CorpID        string `json:"corpId"`
	CorpName      string `json:"corpName"`
	AppKey        string `json:"appKey"`
	AgentID       string `json:"agentId"`
	UnifiedAppID  string `json:"unifiedAppId"`
	AppURL        string `json:"appUrl"`
	NotifyEnabled int    `json:"notifyEnabled"`
	NotifyMode    string `json:"notifyMode"`
	RobotCode     string `json:"robotCode"`
	Enabled       int    `json:"enabled"`
	AppSecretSet  bool   `json:"appSecretSet"`
}
