package online

type SessionInfo struct {
	Token     string `json:"token"`
	TTL       int    `json:"ttl"`
	LoginIP   string `json:"loginIp"`
	LoginTime int64  `json:"loginTime"`
	Device    string `json:"device"`
}

type AdminSession struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	Pic      string `json:"pic"`
	Type     int    `json:"type"`
	RoleName string `json:"roleName"`
	LoginCnt int    `json:"loginCnt"`
	SessionInfo
}

type UserSession struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Mobile   string `json:"mobile"`
	Pic      string `json:"pic"`
	LoginCnt int    `json:"loginCnt"`
	SessionInfo
}
