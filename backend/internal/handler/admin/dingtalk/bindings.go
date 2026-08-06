package dingtalk

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/response"
)

type dingTalkUserBindingResponse struct {
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

type dingTalkUserBindingUserOption struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Account    string `json:"account"`
	MiniOpenID string `json:"miniOpenId"`
	Status     int    `json:"status"`
}

type dingTalkUserBindingUserTreeNode struct {
	Value      interface{}                        `json:"value"`
	Label      string                             `json:"label"`
	Type       string                             `json:"type"`
	UserID     uint                               `json:"userId,omitempty"`
	Account    string                             `json:"account,omitempty"`
	MiniOpenID string                             `json:"miniOpenId,omitempty"`
	Status     int                                `json:"status,omitempty"`
	Disabled   bool                               `json:"disabled,omitempty"`
	Count      int                                `json:"count,omitempty"`
	SearchText string                             `json:"searchText,omitempty"`
	Children   []*dingTalkUserBindingUserTreeNode `json:"children,omitempty"`
}

type dingTalkUserBindingCorpOption struct {
	CorpID   string `json:"corpId"`
	CorpName string `json:"corpName"`
	Enabled  int    `json:"enabled"`
}

func (h *AdminDingTalkHandler) GetUserBindings(ctx context.Context, c *app.RequestContext) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		response.Fail(c, "数据库未初始化")
		return
	}

	page := parsePositiveInt(c.Query("page"), 1)
	pageSize := parsePositiveInt(c.Query("pageSize"), 20)
	if pageSize > 100 {
		pageSize = 100
	}
	corpID := strings.TrimSpace(c.Query("corpId"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	enabled := strings.TrimSpace(c.Query("enabled"))

	query := db.Model(&model.DingTalkH5UserBinding{})
	if corpID != "" {
		query = query.Where("`corp_id` = ?", corpID)
	}
	if enabled == "0" || enabled == "1" {
		query = query.Where("`enabled` = ?", enabled)
	}
	if keyword != "" {
		userIDs, err := dingTalkBindingSearchUserIDs(ctx, db, keyword)
		if err != nil {
			response.Fail(c, "获取失败")
			return
		}
		likeKeyword := "%" + keyword + "%"
		if len(userIDs) > 0 {
			query = query.Where("`dingtalk_user_id` LIKE ? OR `union_id` LIKE ? OR `user_id` IN ?", likeKeyword, likeKeyword, userIDs)
		} else {
			query = query.Where("`dingtalk_user_id` LIKE ? OR `union_id` LIKE ?", likeKeyword, likeKeyword)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Fail(c, "获取失败")
		return
	}
	var rows []model.DingTalkH5UserBinding
	if err := query.Order("`id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		response.Fail(c, "获取失败")
		return
	}

	userMap, err := dingTalkBindingUserMap(ctx, db, rows)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	corpOptions, corpMap, err := dingTalkBindingCorpOptions(ctx)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	users, err := dingTalkBindingUsers(ctx, db)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	userOptions := dingTalkBindingUserOptionsFromUsers(users)
	userTreeOptions, err := dingTalkBindingUserTreeOptions(ctx, db, users)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}

	list := make([]dingTalkUserBindingResponse, 0, len(rows))
	for _, row := range rows {
		user := userMap[row.UserID]
		corpName := corpMap[row.CorpID]
		list = append(list, dingTalkUserBindingResponse{
			ID:             row.ID,
			CorpID:         row.CorpID,
			CorpName:       corpName,
			DingTalkUserID: row.DingTalkUserID,
			UnionID:        row.UnionID,
			UserID:         row.UserID,
			UserName:       user.Name,
			UserAccount:    user.Account,
			UserMiniOpenID: user.MiniOpenID,
			Enabled:        row.Enabled,
			AddTime:        row.AddTime,
			EditTime:       row.EditTime,
		})
	}

	response.JSON(c, map[string]interface{}{
		"list":            list,
		"total":           total,
		"corpOptions":     corpOptions,
		"userOptions":     userOptions,
		"userTreeOptions": userTreeOptions,
	})
}

func (h *AdminDingTalkHandler) SaveUserBinding(ctx context.Context, c *app.RequestContext) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		response.Fail(c, "数据库未初始化")
		return
	}

	id := parseUint(c.PostForm("id"))
	corpID := strings.TrimSpace(c.PostForm("corpId"))
	dingTalkUserID := strings.TrimSpace(c.PostForm("dingTalkUserId"))
	unionID := strings.TrimSpace(c.PostForm("unionId"))
	userID := parseUint(c.PostForm("userId"))
	enabled := 1
	if strings.TrimSpace(c.PostForm("enabled")) == "0" {
		enabled = 0
	}
	if corpID == "" {
		response.Fail(c, "请选择钉钉企业")
		return
	}
	if dingTalkUserID == "" {
		response.Fail(c, "请输入钉钉 UserId")
		return
	}
	if userID == 0 {
		response.Fail(c, "请选择本地用户")
		return
	}
	if err := ensureDingTalkBindingCorpExists(db, corpID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	if err := ensureDingTalkBindingUserExists(db, userID); err != nil {
		response.Fail(c, err.Error())
		return
	}

	now := database.Now()
	if err := db.Transaction(func(tx *gorm.DB) error {
		var row model.DingTalkH5UserBinding
		if id > 0 {
			if err := tx.Where("`id` = ?", id).First(&row).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("绑定不存在")
				}
				return err
			}
		} else {
			err := tx.Where("`corp_id` = ? AND `dingtalk_user_id` = ?", corpID, dingTalkUserID).First(&row).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}

		if row.ID == 0 {
			row = model.DingTalkH5UserBinding{
				CorpID:         corpID,
				DingTalkUserID: dingTalkUserID,
				UnionID:        unionID,
				UserID:         userID,
				Enabled:        enabled,
				AddTime:        now,
				EditTime:       now,
			}
			return tx.Create(&row).Error
		}

		updates := map[string]interface{}{
			"corp_id":          corpID,
			"dingtalk_user_id": dingTalkUserID,
			"union_id":         unionID,
			"user_id":          userID,
			"enabled":          enabled,
			"edit_time":        now,
		}
		return tx.Model(&row).Updates(updates).Error
	}); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, nil)
}

func (h *AdminDingTalkHandler) StatusUserBinding(ctx context.Context, c *app.RequestContext) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		response.Fail(c, "数据库未初始化")
		return
	}
	id := parseUint(c.PostForm("id"))
	if id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	enabled := 0
	if strings.TrimSpace(c.PostForm("enabled")) == "1" {
		enabled = 1
	}
	result := db.Model(&model.DingTalkH5UserBinding{}).Where("`id` = ?", id).Updates(map[string]interface{}{
		"enabled":   enabled,
		"edit_time": database.Now(),
	})
	if result.Error != nil {
		response.Fail(c, "保存失败")
		return
	}
	if result.RowsAffected == 0 {
		response.Fail(c, "绑定不存在")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminDingTalkHandler) DeleteUserBinding(ctx context.Context, c *app.RequestContext) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		response.Fail(c, "数据库未初始化")
		return
	}
	id := parseUint(c.PostForm("id"))
	if id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	result := db.Where("`id` = ?", id).Delete(&model.DingTalkH5UserBinding{})
	if result.Error != nil {
		response.Fail(c, "删除失败")
		return
	}
	if result.RowsAffected == 0 {
		response.Fail(c, "绑定不存在")
		return
	}
	response.JSON(c, nil)
}

func dingTalkBindingSearchUserIDs(ctx context.Context, db *gorm.DB, keyword string) ([]uint, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	likeKeyword := "%" + keyword + "%"
	var users []model.User
	if err := db.Select("id").Where("`user_name` LIKE ? OR `user_account` LIKE ? OR `user_mini_openid` LIKE ? OR `user_mobile` LIKE ?", likeKeyword, likeKeyword, likeKeyword, likeKeyword).Limit(200).Find(&users).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	return ids, nil
}

func dingTalkBindingUserMap(ctx context.Context, db *gorm.DB, rows []model.DingTalkH5UserBinding) (map[uint]model.User, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	ids := make([]uint, 0, len(rows))
	seen := map[uint]struct{}{}
	for _, row := range rows {
		if row.UserID == 0 {
			continue
		}
		if _, ok := seen[row.UserID]; ok {
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

func dingTalkBindingCorpOptions(ctx context.Context) ([]dingTalkUserBindingCorpOption, map[string]string, error) {
	configs, err := listDingTalkH5CorpConfigsContext(ctx)
	if err != nil {
		return nil, nil, err
	}
	options := make([]dingTalkUserBindingCorpOption, 0, len(configs))
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
		options = append(options, dingTalkUserBindingCorpOption{
			CorpID:   corpID,
			CorpName: name,
			Enabled:  config.Enabled,
		})
		corpMap[corpID] = name
	}
	return options, corpMap, nil
}

func dingTalkBindingUserOptions(ctx context.Context, db *gorm.DB) ([]dingTalkUserBindingUserOption, error) {
	users, err := dingTalkBindingUsers(ctx, db)
	if err != nil {
		return nil, err
	}
	return dingTalkBindingUserOptionsFromUsers(users), nil
}

func dingTalkBindingUsers(ctx context.Context, db *gorm.DB) ([]model.User, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	var users []model.User
	if err := db.Select("id", "user_name", "user_account", "user_mini_openid", "user_status").Order("`user_status` DESC, `user_name` ASC, `id` ASC").Limit(1000).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func dingTalkBindingUserOptionsFromUsers(users []model.User) []dingTalkUserBindingUserOption {
	options := make([]dingTalkUserBindingUserOption, 0, len(users))
	for _, user := range users {
		options = append(options, dingTalkUserBindingUserOption{
			ID:         user.ID,
			Name:       user.Name,
			Account:    user.Account,
			MiniOpenID: user.MiniOpenID,
			Status:     user.Status,
		})
	}
	return options
}

func dingTalkBindingUserTreeOptions(ctx context.Context, db *gorm.DB, users []model.User) ([]*dingTalkUserBindingUserTreeNode, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
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
	return buildDingTalkBindingUserTree(departments, users, userDepts), nil
}

func buildDingTalkBindingUserTree(departments []model.Department, users []model.User, userDepts []model.UserDept) []*dingTalkUserBindingUserTreeNode {
	deptNodes := make(map[uint]*dingTalkUserBindingUserTreeNode, len(departments))
	roots := make([]*dingTalkUserBindingUserTreeNode, 0, len(departments)+1)
	unassigned := &dingTalkUserBindingUserTreeNode{
		Value:      "dept-unassigned",
		Label:      "未设置部门",
		Type:       "dept",
		Disabled:   true,
		SearchText: "未设置部门",
	}
	roots = append(roots, unassigned)

	for _, dept := range departments {
		if dept.ID == 0 {
			continue
		}
		label := strings.TrimSpace(dept.Name)
		if label == "" {
			label = "未命名部门"
		}
		deptNodes[dept.ID] = &dingTalkUserBindingUserTreeNode{
			Value:      "dept-" + strconv.FormatUint(uint64(dept.ID), 10),
			Label:      label,
			Type:       "dept",
			Disabled:   true,
			SearchText: label,
		}
	}
	for _, dept := range departments {
		node := deptNodes[dept.ID]
		if node == nil {
			continue
		}
		parent := deptNodes[dept.ParentID]
		if parent == nil {
			roots = append(roots, node)
			continue
		}
		parent.Children = append(parent.Children, node)
	}

	firstDeptByUser := make(map[uint]uint, len(userDepts))
	for _, userDept := range userDepts {
		if userDept.UserID == 0 || userDept.DeptID == 0 {
			continue
		}
		if _, ok := firstDeptByUser[userDept.UserID]; !ok {
			firstDeptByUser[userDept.UserID] = userDept.DeptID
		}
	}
	for _, user := range users {
		if user.ID == 0 {
			continue
		}
		parent := deptNodes[firstDeptByUser[user.ID]]
		if parent == nil {
			parent = unassigned
		}
		parent.Children = append(parent.Children, &dingTalkUserBindingUserTreeNode{
			Value:      user.ID,
			Label:      dingTalkBindingUserDisplayName(user),
			Type:       "user",
			UserID:     user.ID,
			Account:    strings.TrimSpace(user.Account),
			MiniOpenID: strings.TrimSpace(user.MiniOpenID),
			Status:     user.Status,
			SearchText: dingTalkBindingUserSearchText(user),
		})
	}
	return pruneAndCountDingTalkBindingUserTree(roots)
}

func pruneAndCountDingTalkBindingUserTree(nodes []*dingTalkUserBindingUserTreeNode) []*dingTalkUserBindingUserTreeNode {
	result := make([]*dingTalkUserBindingUserTreeNode, 0, len(nodes))
	for _, node := range nodes {
		if node == nil {
			continue
		}
		if node.Type == "user" {
			node.Count = 1
			result = append(result, node)
			continue
		}
		node.Children = pruneAndCountDingTalkBindingUserTree(node.Children)
		count := 0
		for _, child := range node.Children {
			count += child.Count
		}
		node.Count = count
		if count > 0 {
			result = append(result, node)
		}
	}
	return result
}

func dingTalkBindingUserDisplayName(user model.User) string {
	name := strings.TrimSpace(user.Name)
	if name != "" {
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

func dingTalkBindingUserSearchText(user model.User) string {
	return strings.Join([]string{
		dingTalkBindingUserDisplayName(user),
		strings.TrimSpace(user.Account),
		strings.TrimSpace(user.MiniOpenID),
		strconv.FormatUint(uint64(user.ID), 10),
	}, " ")
}

func ensureDingTalkBindingCorpExists(db *gorm.DB, corpID string) error {
	var count int64
	if err := db.Model(&model.DingTalkH5CorpConfig{}).Where("`corp_id` = ?", corpID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("钉钉企业配置不存在")
	}
	return nil
}

func ensureDingTalkBindingUserExists(db *gorm.DB, userID uint) error {
	var count int64
	if err := db.Model(&model.User{}).Where("`id` = ?", userID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("本地用户不存在")
	}
	return nil
}

func parsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseUint(value string) uint {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0
	}
	return uint(parsed)
}
