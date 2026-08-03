package user

import adminuserservice "wecheckin/backend/internal/app/service/adminuser"

type userListResponse struct {
	List  []adminuserservice.UserListItem `json:"list"`
	Total int64                           `json:"total"`
}
