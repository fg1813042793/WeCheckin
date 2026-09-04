package dingtalk

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	dingtalkh5config "wecheckin/backend/internal/service/dingtalkh5/config"
	"wecheckin/backend/pkg/database"
)

var (
	ErrBindingNotFound = errors.New("dingtalk user binding not found")
	ErrCorpNotFound    = errors.New("dingtalk corp config not found")
	ErrUserNotFound    = errors.New("local user not found")
)

type UserBinding struct {
	ID             uint   `json:"id"`
	CorpID         string `json:"corpId"`
	CorpName       string `json:"corpName"`
	DingTalkUserID string `json:"dingTalkUserId"`
	UnionID        string `json:"unionId"`
	UserID         uint   `json:"userId"`
	UserName       string `json:"userName"`
	UserAccount    string `json:"userAccount"`
	UserMiniOpenID string `json:"userMiniOpenId"`
	Enabled        int    `json:"enabled"`
	AddTime        int64  `json:"addTime"`
	EditTime       int64  `json:"editTime"`
}

type UserBindingUserOption struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Account    string `json:"account"`
	MiniOpenID string `json:"miniOpenId"`
	Status     int    `json:"status"`
}

type UserBindingTreeNode struct {
	Value      any                    `json:"value"`
	Label      string                 `json:"label"`
	Type       string                 `json:"type"`
	UserID     uint                   `json:"userId,omitempty"`
	Account    string                 `json:"account,omitempty"`
	MiniOpenID string                 `json:"miniOpenId,omitempty"`
	Status     int                    `json:"status,omitempty"`
	Disabled   bool                   `json:"disabled,omitempty"`
	Count      int                    `json:"count,omitempty"`
	SearchText string                 `json:"searchText,omitempty"`
	Children   []*UserBindingTreeNode `json:"children,omitempty"`
}

type UserBindingCorpOption struct {
	CorpID   string `json:"corpId"`
	CorpName string `json:"corpName"`
	Enabled  int    `json:"enabled"`
}

type UserBindingList struct {
	List            []UserBinding           `json:"list"`
	Total           int64                   `json:"total"`
	CorpOptions     []UserBindingCorpOption `json:"corpOptions"`
	UserOptions     []UserBindingUserOption `json:"userOptions"`
	UserTreeOptions []*UserBindingTreeNode  `json:"userTreeOptions"`
}

type UserBindingQuery struct {
	Page     int
	PageSize int
	CorpID   string
	Keyword  string
	Enabled  string
}

type SaveUserBindingInput struct {
	ID             uint
	CorpID         string
	DingTalkUserID string
	UnionID        string
	UserID         uint
	Enabled        int
}

func (service *Service) ListUserBindings(ctx context.Context, query UserBindingQuery) (UserBindingList, error) {
	db, cancel, err := service.contextDB(ctx)
	if err != nil {
		return UserBindingList{}, err
	}
	defer cancel()
	query.Page, query.PageSize = normalizePagination(query.Page, query.PageSize)
	query.CorpID = strings.TrimSpace(query.CorpID)
	query.Keyword = strings.TrimSpace(query.Keyword)
	query.Enabled = strings.TrimSpace(query.Enabled)
	statement := db.Model(&model.DingTalkH5UserBinding{})
	if query.CorpID != "" {
		statement = statement.Where("`corp_id` = ?", query.CorpID)
	}
	if query.Enabled == "0" || query.Enabled == "1" {
		statement = statement.Where("`enabled` = ?", query.Enabled)
	}
	if query.Keyword != "" {
		userIDs, err := bindingSearchUserIDs(ctx, db, query.Keyword)
		if err != nil {
			return UserBindingList{}, err
		}
		like := "%" + query.Keyword + "%"
		if len(userIDs) > 0 {
			statement = statement.Where("`dingtalk_user_id` LIKE ? OR `union_id` LIKE ? OR `user_id` IN ?", like, like, userIDs)
		} else {
			statement = statement.Where("`dingtalk_user_id` LIKE ? OR `union_id` LIKE ?", like, like)
		}
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return UserBindingList{}, err
	}
	var rows []model.DingTalkH5UserBinding
	if err := statement.Order("`id` DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&rows).Error; err != nil {
		return UserBindingList{}, err
	}
	userMap, err := bindingUserMap(ctx, db, rows)
	if err != nil {
		return UserBindingList{}, err
	}
	corpOptions, corpMap, err := bindingCorpOptions(ctx)
	if err != nil {
		return UserBindingList{}, err
	}
	users, err := bindingUsers(ctx, db)
	if err != nil {
		return UserBindingList{}, err
	}
	userTreeOptions, err := bindingUserTreeOptions(ctx, db, users)
	if err != nil {
		return UserBindingList{}, err
	}
	list := make([]UserBinding, 0, len(rows))
	for _, row := range rows {
		user := userMap[row.UserID]
		list = append(list, UserBinding{
			ID: row.ID, CorpID: row.CorpID, CorpName: corpMap[row.CorpID], DingTalkUserID: row.DingTalkUserID,
			UnionID: row.UnionID, UserID: row.UserID, UserName: user.Name, UserAccount: user.Account,
			UserMiniOpenID: user.MiniOpenID, Enabled: row.Enabled, AddTime: row.AddTime, EditTime: row.EditTime,
		})
	}
	return UserBindingList{
		List: list, Total: total, CorpOptions: corpOptions,
		UserOptions: bindingUserOptionsFromUsers(users), UserTreeOptions: userTreeOptions,
	}, nil
}

func (service *Service) SaveUserBinding(ctx context.Context, input SaveUserBindingInput) error {
	db, cancel, err := service.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	input.CorpID = strings.TrimSpace(input.CorpID)
	input.DingTalkUserID = strings.TrimSpace(input.DingTalkUserID)
	input.UnionID = strings.TrimSpace(input.UnionID)
	if err := ensureBindingCorpExists(db, input.CorpID); err != nil {
		return err
	}
	if err := ensureBindingUserExists(db, input.UserID); err != nil {
		return err
	}
	now := database.Now()
	return db.Transaction(func(tx *gorm.DB) error {
		var row model.DingTalkH5UserBinding
		if input.ID > 0 {
			if err := tx.Where("`id` = ?", input.ID).First(&row).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrBindingNotFound
				}
				return err
			}
		} else {
			err := tx.Where("`corp_id` = ? AND `dingtalk_user_id` = ?", input.CorpID, input.DingTalkUserID).First(&row).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		if row.ID == 0 {
			return tx.Create(&model.DingTalkH5UserBinding{
				CorpID: input.CorpID, DingTalkUserID: input.DingTalkUserID, UnionID: input.UnionID,
				UserID: input.UserID, Enabled: input.Enabled, AddTime: now, EditTime: now,
			}).Error
		}
		return tx.Model(&row).Updates(map[string]any{
			"corp_id": input.CorpID, "dingtalk_user_id": input.DingTalkUserID, "union_id": input.UnionID,
			"user_id": input.UserID, "enabled": input.Enabled, "edit_time": now,
		}).Error
	})
}

func (service *Service) SetUserBindingEnabled(ctx context.Context, id uint, enabled int) error {
	db, cancel, err := service.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	result := db.Model(&model.DingTalkH5UserBinding{}).Where("`id` = ?", id).Updates(map[string]any{
		"enabled": enabled, "edit_time": database.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrBindingNotFound
	}
	return nil
}

func (service *Service) DeleteUserBinding(ctx context.Context, id uint) error {
	db, cancel, err := service.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	result := db.Where("`id` = ?", id).Delete(&model.DingTalkH5UserBinding{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrBindingNotFound
	}
	return nil
}

func bindingSearchUserIDs(ctx context.Context, db *gorm.DB, keyword string) ([]uint, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	like := "%" + keyword + "%"
	var users []model.User
	if err := db.Select("id").Where("`user_name` LIKE ? OR `user_account` LIKE ? OR `user_mini_openid` LIKE ? OR `user_mobile` LIKE ?", like, like, like, like).Limit(200).Find(&users).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	return ids, nil
}

func bindingUserMap(ctx context.Context, db *gorm.DB, rows []model.DingTalkH5UserBinding) (map[uint]model.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(rows))
	seen := make(map[uint]struct{}, len(rows))
	for _, row := range rows {
		if row.UserID == 0 {
			continue
		}
		if _, exists := seen[row.UserID]; exists {
			continue
		}
		seen[row.UserID] = struct{}{}
		ids = append(ids, row.UserID)
	}
	result := make(map[uint]model.User, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	var users []model.User
	if err := db.Select("id", "user_name", "user_account", "user_mini_openid").Where("`id` IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	for _, user := range users {
		result[user.ID] = user
	}
	return result, nil
}

func bindingCorpOptions(ctx context.Context) ([]UserBindingCorpOption, map[string]string, error) {
	configs, err := dingtalkh5config.ListDingTalkH5CorpConfigsContext(ctx)
	if err != nil {
		return nil, nil, err
	}
	options := make([]UserBindingCorpOption, 0, len(configs))
	corpMap := make(map[string]string, len(configs))
	for _, config := range configs {
		corpID := strings.TrimSpace(config.CorpID)
		if corpID == "" {
			continue
		}
		name := strings.TrimSpace(config.CorpName)
		if name == "" {
			name = corpID
		}
		options = append(options, UserBindingCorpOption{CorpID: corpID, CorpName: name, Enabled: config.Enabled})
		corpMap[corpID] = name
	}
	return options, corpMap, nil
}

func bindingUsers(ctx context.Context, db *gorm.DB) ([]model.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var users []model.User
	if err := db.Select("id", "user_name", "user_account", "user_mini_openid", "user_status").Order("`user_status` DESC, `user_name` ASC, `id` ASC").Limit(1000).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func bindingUserOptionsFromUsers(users []model.User) []UserBindingUserOption {
	options := make([]UserBindingUserOption, 0, len(users))
	for _, user := range users {
		options = append(options, UserBindingUserOption{
			ID: user.ID, Name: user.Name, Account: user.Account, MiniOpenID: user.MiniOpenID, Status: user.Status,
		})
	}
	return options
}

func bindingUserTreeOptions(ctx context.Context, db *gorm.DB, users []model.User) ([]*UserBindingTreeNode, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var departments []model.Department
	if err := db.Table("`departments`").Select("id", "dept_name", "dept_parent_id", "dept_sort", "dept_status").Order("`dept_sort` ASC, `id` ASC").Find(&departments).Error; err != nil {
		return nil, err
	}
	userIDs := make([]uint, 0, len(users))
	for _, user := range users {
		if user.ID > 0 {
			userIDs = append(userIDs, user.ID)
		}
	}
	var userDepts []model.UserDept
	if len(userIDs) > 0 {
		if err := db.Table("`user_depts`").Where("`user_dept_user_id` IN ?", userIDs).Order("`id` ASC").Find(&userDepts).Error; err != nil {
			return nil, err
		}
	}
	return buildBindingUserTree(departments, users, userDepts), nil
}

func buildBindingUserTree(departments []model.Department, users []model.User, userDepts []model.UserDept) []*UserBindingTreeNode {
	departmentNodes := make(map[uint]*UserBindingTreeNode, len(departments))
	roots := make([]*UserBindingTreeNode, 0, len(departments)+1)
	unassigned := &UserBindingTreeNode{Value: "dept-unassigned", Label: "未设置部门", Type: "dept", Disabled: true, SearchText: "未设置部门"}
	roots = append(roots, unassigned)
	for _, department := range departments {
		if department.ID == 0 {
			continue
		}
		label := strings.TrimSpace(department.Name)
		if label == "" {
			label = "未命名部门"
		}
		departmentNodes[department.ID] = &UserBindingTreeNode{
			Value: "dept-" + strconv.FormatUint(uint64(department.ID), 10), Label: label,
			Type: "dept", Disabled: true, SearchText: label,
		}
	}
	for _, department := range departments {
		node := departmentNodes[department.ID]
		if node == nil {
			continue
		}
		parent := departmentNodes[department.ParentID]
		if parent == nil {
			roots = append(roots, node)
			continue
		}
		parent.Children = append(parent.Children, node)
	}
	firstDepartmentByUser := make(map[uint]uint, len(userDepts))
	for _, userDepartment := range userDepts {
		if userDepartment.UserID == 0 || userDepartment.DeptID == 0 {
			continue
		}
		if _, exists := firstDepartmentByUser[userDepartment.UserID]; !exists {
			firstDepartmentByUser[userDepartment.UserID] = userDepartment.DeptID
		}
	}
	for _, user := range users {
		if user.ID == 0 {
			continue
		}
		parent := departmentNodes[firstDepartmentByUser[user.ID]]
		if parent == nil {
			parent = unassigned
		}
		parent.Children = append(parent.Children, &UserBindingTreeNode{
			Value: user.ID, Label: bindingUserDisplayName(user), Type: "user", UserID: user.ID,
			Account: strings.TrimSpace(user.Account), MiniOpenID: strings.TrimSpace(user.MiniOpenID),
			Status: user.Status, SearchText: bindingUserSearchText(user),
		})
	}
	return pruneAndCountBindingUserTree(roots)
}

func pruneAndCountBindingUserTree(nodes []*UserBindingTreeNode) []*UserBindingTreeNode {
	result := make([]*UserBindingTreeNode, 0, len(nodes))
	for _, node := range nodes {
		if node == nil {
			continue
		}
		if node.Type == "user" {
			node.Count = 1
			result = append(result, node)
			continue
		}
		node.Children = pruneAndCountBindingUserTree(node.Children)
		for _, child := range node.Children {
			node.Count += child.Count
		}
		if node.Count > 0 {
			result = append(result, node)
		}
	}
	return result
}

func bindingUserDisplayName(user model.User) string {
	if name := strings.TrimSpace(user.Name); name != "" {
		return name
	}
	if account := strings.TrimSpace(user.Account); account != "" {
		return account
	}
	if miniOpenID := strings.TrimSpace(user.MiniOpenID); miniOpenID != "" {
		return miniOpenID
	}
	return "用户 " + strconv.FormatUint(uint64(user.ID), 10)
}

func bindingUserSearchText(user model.User) string {
	return strings.Join([]string{
		bindingUserDisplayName(user), strings.TrimSpace(user.Account), strings.TrimSpace(user.MiniOpenID), strconv.FormatUint(uint64(user.ID), 10),
	}, " ")
}

func ensureBindingCorpExists(db *gorm.DB, corpID string) error {
	var count int64
	if err := db.Model(&model.DingTalkH5CorpConfig{}).Where("`corp_id` = ?", corpID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrCorpNotFound
	}
	return nil
}

func ensureBindingUserExists(db *gorm.DB, userID uint) error {
	var count int64
	if err := db.Model(&model.User{}).Where("`id` = ?", userID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrUserNotFound
	}
	return nil
}
