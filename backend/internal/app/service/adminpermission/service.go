package adminpermission

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	menuservice "wecheckin-backend/backend/internal/app/service/menu"
	permissionsupport "wecheckin-backend/backend/internal/app/support/permission"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

type PermissionNode struct {
	ID            uint              `json:"id"`
	Key           string            `json:"key"`
	PermissionKey string            `json:"permissionKey"`
	Name          string            `json:"name"`
	Platform      string            `json:"platform"`
	Type          string            `json:"type"`
	ParentKey     string            `json:"parentKey"`
	ResourceID    uint              `json:"resourceId"`
	ResourcePath  string            `json:"resourcePath"`
	Path          string            `json:"path"`
	Perms         string            `json:"perms"`
	Icon          string            `json:"icon"`
	Sort          int               `json:"sort"`
	Status        int               `json:"status"`
	Children      []*PermissionNode `json:"children,omitempty"`
}

type SaveRequest struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	Platform     string `json:"platform"`
	Type         string `json:"type"`
	ParentKey    string `json:"parentKey"`
	ResourcePath string `json:"resourcePath"`
	Path         string `json:"path"`
	Perms        string `json:"perms"`
	Icon         string `json:"icon"`
	Sort         int    `json:"sort"`
	Status       int    `json:"status"`
}

func Tree(platform string) ([]*PermissionNode, error) {
	return TreeContext(context.Background(), platform)
}

func TreeContext(ctx context.Context, platform string, types ...string) ([]*PermissionNode, error) {
	filterTypes := normalizePermissionTypes(types...)
	now := time.Now()
	if cached, ok := getPermissionTreeCache(platform, filterTypes, now); ok {
		return cached, nil
	}
	rows, err := ListContext(ctx, platform, filterTypes...)
	if err != nil {
		return nil, err
	}
	tree := buildPermissionTree(permissionRowsToNodes(rows))
	setPermissionTreeCache(platform, filterTypes, tree, now)
	return tree, nil
}

func List(platform string) ([]model.Permission, error) {
	return ListContext(context.Background(), platform)
}

func ListContext(ctx context.Context, platform string, types ...string) ([]model.Permission, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return nil, fmt.Errorf("数据库连接异常")
	}
	query := db.Model(&model.Permission{})
	platform = strings.TrimSpace(platform)
	if platform != "" {
		query = query.Where("`permission_platform` = ?", platform)
	}
	filterTypes := normalizePermissionTypes(types...)
	if len(filterTypes) > 0 {
		query = query.Where("`permission_type` IN ?", filterTypes)
	}
	var list []model.Permission
	err := query.Order("`permission_platform` ASC, `permission_sort` ASC, `id` ASC").Find(&list).Error
	return list, err
}

func normalizePermissionTypes(values ...string) []string {
	result := make([]string, 0)
	seen := make(map[string]struct{})
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			item := strings.TrimSpace(part)
			if item == "" {
				continue
			}
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func Add(req SaveRequest) error {
	return AddContext(context.Background(), req)
}

func AddContext(ctx context.Context, req SaveRequest) error {
	item, err := normalizePermissionRequest(req)
	if err != nil {
		return err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return fmt.Errorf("数据库连接异常")
	}
	if err := permissionsupport.EnsurePermissionSchemaContext(ctx, db); err != nil {
		return err
	}
	var count int64
	if err := db.Model(&model.Permission{}).Where("`permission_key` = ?", item.Key).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("权限编码已存在")
	}
	now := database.Now()
	item.AddTime = now
	item.EditTime = now
	if err := db.Create(&item).Error; err != nil {
		return err
	}
	invalidatePermissionTreeCache()
	menuservice.InvalidateAdminPermCache()
	return nil
}

func Edit(key string, req SaveRequest) error {
	return EditContext(context.Background(), key, req)
}

func EditContext(ctx context.Context, key string, req SaveRequest) error {
	oldKey := strings.TrimSpace(key)
	if oldKey == "" {
		return fmt.Errorf("权限编码不能为空")
	}
	item, err := normalizePermissionRequest(req)
	if err != nil {
		return err
	}
	newKey := item.Key
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return fmt.Errorf("数据库连接异常")
	}
	if err := permissionsupport.EnsurePermissionSchemaContext(ctx, db); err != nil {
		return err
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if oldKey != newKey {
			var count int64
			if err := tx.Model(&model.Permission{}).Where("`permission_key` = ? AND `permission_key` <> ?", newKey, oldKey).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return fmt.Errorf("权限编码已存在")
			}
		}
		updates := map[string]interface{}{
			"permission_key":           newKey,
			"permission_name":          item.Name,
			"permission_platform":      item.Platform,
			"permission_type":          item.Type,
			"permission_parent_key":    item.ParentKey,
			"permission_resource_path": item.ResourcePath,
			"permission_icon":          item.Icon,
			"permission_perms":         item.Perms,
			"permission_sort":          item.Sort,
			"permission_status":        item.Status,
			"permission_edit_time":     database.Now(),
		}
		res := tx.Model(&model.Permission{}).Where("`permission_key` = ?", oldKey).Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if oldKey == newKey {
			return nil
		}
		if err := tx.Model(&model.PermissionGrant{}).Where("`grant_permission_key` = ?", oldKey).Update("grant_permission_key", newKey).Error; err != nil {
			return err
		}
		return tx.Model(&model.Permission{}).Where("`permission_parent_key` = ?", oldKey).Update("permission_parent_key", newKey).Error
	}); err != nil {
		return err
	}
	invalidatePermissionTreeCache()
	menuservice.InvalidateAdminPermCache()
	return nil
}

func Delete(key string) error {
	return DeleteContext(context.Background(), key)
}

func DeleteContext(ctx context.Context, key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("权限编码不能为空")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return fmt.Errorf("数据库连接异常")
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		keys, err := collectDescendantKeys(ctx, tx, []string{key})
		if err != nil {
			return err
		}
		if len(keys) == 0 {
			return nil
		}
		if err := tx.Where("`grant_permission_key` IN ?", keys).Delete(&model.PermissionGrant{}).Error; err != nil {
			return err
		}
		return tx.Where("`permission_key` IN ?", keys).Delete(&model.Permission{}).Error
	})
	if err == nil {
		invalidatePermissionTreeCache()
		menuservice.InvalidateAdminPermCache()
	}
	return err
}

func collectDescendantKeys(ctx context.Context, tx *gorm.DB, roots []string) ([]string, error) {
	seen := map[string]bool{}
	result := make([]string, 0, len(roots))
	frontier := normalizeKeys(roots)
	for len(frontier) > 0 {
		nextFrontier := make([]string, 0)
		for _, key := range frontier {
			if seen[key] {
				continue
			}
			seen[key] = true
			result = append(result, key)
			nextFrontier = append(nextFrontier, key)
		}
		if len(nextFrontier) == 0 {
			break
		}
		if err := ctxErr(ctx); err != nil {
			return nil, err
		}
		var children []model.Permission
		if err := tx.Select("permission_key").Where("`permission_parent_key` IN ?", nextFrontier).Find(&children).Error; err != nil {
			return nil, err
		}
		frontier = make([]string, 0, len(children))
		for _, child := range children {
			if child.Key != "" && !seen[child.Key] {
				frontier = append(frontier, child.Key)
			}
		}
	}
	return result, nil
}

func normalizePermissionRequest(req SaveRequest) (model.Permission, error) {
	key := strings.TrimSpace(req.Key)
	name := strings.TrimSpace(req.Name)
	if key == "" {
		return model.Permission{}, fmt.Errorf("权限编码不能为空")
	}
	if name == "" {
		return model.Permission{}, fmt.Errorf("权限名称不能为空")
	}
	platform := strings.TrimSpace(req.Platform)
	if platform == "" {
		platform = permissionsupport.PlatformAdmin
	}
	ptype := strings.TrimSpace(req.Type)
	if ptype == "" {
		ptype = permissionsupport.TypeMenu
	}
	path := strings.TrimSpace(req.ResourcePath)
	if path == "" {
		path = strings.TrimSpace(req.Path)
	}
	return model.Permission{
		Key:          key,
		Name:         name,
		Platform:     platform,
		Type:         ptype,
		ParentKey:    strings.TrimSpace(req.ParentKey),
		ResourcePath: path,
		Icon:         strings.TrimSpace(req.Icon),
		Perms:        strings.TrimSpace(req.Perms),
		Sort:         req.Sort,
		Status:       req.Status,
	}, nil
}

func permissionRowsToNodes(rows []model.Permission) []*PermissionNode {
	nodes := make([]*PermissionNode, 0, len(rows))
	for _, row := range rows {
		nodes = append(nodes, &PermissionNode{
			ID:            row.ID,
			Key:           row.Key,
			PermissionKey: row.Key,
			Name:          row.Name,
			Platform:      row.Platform,
			Type:          row.Type,
			ParentKey:     row.ParentKey,
			ResourceID:    row.ResourceID,
			ResourcePath:  row.ResourcePath,
			Path:          row.ResourcePath,
			Perms:         row.Perms,
			Icon:          row.Icon,
			Sort:          row.Sort,
			Status:        row.Status,
		})
	}
	return nodes
}

func buildPermissionTree(nodes []*PermissionNode) []*PermissionNode {
	byKey := make(map[string]*PermissionNode, len(nodes))
	for _, node := range nodes {
		node.Children = nil
		byKey[node.Key] = node
	}
	tree := make([]*PermissionNode, 0)
	for _, node := range nodes {
		parent := byKey[node.ParentKey]
		if parent == nil || parent.Key == node.Key {
			tree = append(tree, node)
			continue
		}
		parent.Children = append(parent.Children, node)
	}
	return tree
}

func normalizeKeys(keys []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key != "" && !seen[key] {
			seen[key] = true
			result = append(result, key)
		}
	}
	return result
}

func ctxErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}
