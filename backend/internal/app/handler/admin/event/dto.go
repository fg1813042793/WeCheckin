package event

type listResponse struct {
	List interface{} `json:"list"`
}

type pagedListResponse struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
}
