package user

import adminuserservice "wecheckin/backend/internal/service/admin/adminuser"

type userListResponse struct {
	List  []adminuserservice.UserListItem `json:"list"`
	Total int64                           `json:"total"`
}
