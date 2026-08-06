package mgr

type pagedListResponse struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
}
