package role

import (
	"context"
	"strings"

	"gorm.io/gorm"

	menuservice "wecheckin/backend/internal/app/service/menu"
	"wecheckin/backend/internal/app/support/access"
	"wecheckin/backend/internal/app/support/adminaccess"
	"wecheckin/backend/internal/app/support/appapiperm"
	"wecheckin/backend/internal/app/support/appmenuperm"
	permissionsupport "wecheckin/backend/internal/app/support/permission"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

type ListItem struct {
	ID                          uint     `json:"id"`
	Name                        string   `json:"name"`
	Remark                      string   `json:"remark"`
	Sort                        int      `json:"sort"`
	Status                      int      `json:"status"`
	AllowAdminLogin             int      `json:"allowAdminLogin"`
	DataScope                   int      `json:"dataScope"`
	AddTime                     int64    `json:"addTime"`
	EditTime                    int64    `json:"editTime"`
	AdminPermissionKeys         []string `json:"adminPermissionKeys"`
	AdminAPIPermissionKeys      []string `json:"adminApiPermissionKeys"`
	DeptIDs                     []uint   `json:"deptIds"`
	ClientMenuKeys              []string `json:"clientMenuKeys"`
	DingTalkH5MenuKeys          []string `json:"dingtalkH5MenuKeys"`
	ClientAPIPermissionKeys     []string `json:"clientApiPermissionKeys"`
	DingTalkH5APIPermissionKeys []string `json:"dingtalkH5ApiPermissionKeys"`
}

type ListResponse struct {
	List  []ListItem `json:"list"`
	Total int64      `json:"total"`
}

type ApplicationPermissionNode struct {
	Key       string                      `json:"key"`
	Name      string                      `json:"name"`
	ParentKey string                      `json:"parentKey"`
	Children  []ApplicationPermissionNode `json:"children,omitempty"`
}

type ApplicationPermissionTreeResponse struct {
	Client        []ApplicationPermissionNode `json:"client"`
	DingTalkH5    []ApplicationPermissionNode `json:"dingtalkH5"`
	ClientAPI     []ApplicationPermissionNode `json:"clientApi"`
	DingTalkH5API []ApplicationPermissionNode `json:"dingtalkH5Api"`
}

func GetList(adminID uint, keyword string, page, pageSize int) (*ListResponse, error) {
	return GetListContext(context.Background(), adminID, keyword, page, pageSize)
}

func GetListContext(ctx context.Context, adminID uint, keyword string, page, pageSize int) (*ListResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var admin model.Admin
	db.First(&admin, adminID)

	var conditions []func(*gorm.DB) *gorm.DB
	if !adminaccess.IsReservedSuperAdminRoleContext(ctx, db, admin.RoleID) {
		scope, err := permissionsupport.DataScopeContext(ctx, db, admin.ID, admin.RoleID)
		if err == nil && scope.Ready && scope.Mode == 1 {
			// 全部数据权限可查看全部角色。
		} else if err == nil && scope.Ready && (scope.Mode == 2 || scope.Mode == 4) {
			deptIDs := access.VisibleDeptIDsWithDBContext(ctx, db, &admin)
			roleIDs, _ := permissionsupport.RoleIDsByCustomDeptIDsContext(ctx, db, deptIDs)
			rid := admin.RoleID
			conditions = append(conditions, func(d *gorm.DB) *gorm.DB {
				if len(roleIDs) == 0 {
					return d.Where("`id` = ?", rid)
				}
				return d.Where("`id` IN ? OR `id` = ?", roleIDs, rid)
			})
		} else {
			rid := admin.RoleID
			conditions = append(conditions, func(d *gorm.DB) *gorm.DB {
				return d.Where("`id` = ?", rid)
			})
		}
	}
	if keyword != "" {
		kw := keyword
		conditions = append(conditions, func(d *gorm.DB) *gorm.DB {
			return d.Where("`role_name` LIKE ?", "%"+kw+"%")
		})
	}
	var total int64
	if err := db.Model(&model.Role{}).Scopes(conditions...).Count(&total).Error; err != nil {
		return nil, err
	}
	var list []model.Role
	if err := db.Model(&model.Role{}).Scopes(conditions...).Order("`role_sort` ASC, `id` ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	assignments, err := loadRoleAssignmentMapsContext(ctx, db, list)
	if err != nil {
		return nil, err
	}
	result := make([]ListItem, len(list))
	for i, r := range list {
		result[i] = ListItem{
			ID:                          r.ID,
			Name:                        r.Name,
			Remark:                      r.Remark,
			Sort:                        r.Sort,
			Status:                      r.Status,
			AllowAdminLogin:             normalizeAllowAdminLogin(r.AllowAdminLogin),
			DataScope:                   r.DataScope,
			AddTime:                     r.AddTime,
			EditTime:                    r.EditTime,
			AdminPermissionKeys:         assignments.AdminPermissionKeys[r.ID],
			AdminAPIPermissionKeys:      assignments.AdminAPIPermissionKeys[r.ID],
			DeptIDs:                     assignments.DeptIDs[r.ID],
			ClientMenuKeys:              assignments.ClientMenuKeys[r.ID],
			DingTalkH5MenuKeys:          assignments.DingTalkH5MenuKeys[r.ID],
			ClientAPIPermissionKeys:     assignments.ClientAPIPermissionKeys[r.ID],
			DingTalkH5APIPermissionKeys: assignments.DingTalkH5APIPermissionKeys[r.ID],
		}
	}
	return &ListResponse{List: result, Total: total}, nil
}

func loadRoleAssignmentMapsContext(ctx context.Context, db *gorm.DB, list []model.Role) (permissionsupport.RoleAssignmentMaps, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return permissionsupport.RoleAssignmentMaps{}, err
		}
	}
	roleIDs := make([]uint, 0, len(list))
	for _, item := range list {
		roleIDs = append(roleIDs, item.ID)
	}
	return permissionsupport.RoleAssignmentMapsContext(ctx, db, roleIDs)
}

func loadRoleAdminPermissionKeyMapContext(ctx context.Context, db *gorm.DB, list []model.Role) (map[uint][]string, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	roleIDs := make([]uint, 0, len(list))
	keysByRole := make(map[uint][]string, len(list))
	for _, item := range list {
		roleIDs = append(roleIDs, item.ID)
		keysByRole[item.ID] = []string{}
	}
	if len(roleIDs) == 0 {
		return keysByRole, nil
	}
	return permissionsupport.RoleAdminPermissionKeyMapContext(ctx, db, roleIDs)
}

func loadRoleAdminAPIPermissionKeyMapContext(ctx context.Context, db *gorm.DB, list []model.Role) (map[uint][]string, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	roleIDs := make([]uint, 0, len(list))
	keysByRole := make(map[uint][]string, len(list))
	for _, item := range list {
		roleIDs = append(roleIDs, item.ID)
		keysByRole[item.ID] = []string{}
	}
	if len(roleIDs) == 0 {
		return keysByRole, nil
	}
	return permissionsupport.RoleAdminAPIPermissionKeyMapContext(ctx, db, roleIDs)
}

func loadRoleDeptIDMapContext(ctx context.Context, db *gorm.DB, list []model.Role) (map[uint][]uint, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	roleIDs := make([]uint, 0, len(list))
	deptIDsByRole := make(map[uint][]uint, len(list))
	for _, item := range list {
		roleIDs = append(roleIDs, item.ID)
		deptIDsByRole[item.ID] = []uint{}
	}
	if len(roleIDs) == 0 {
		return deptIDsByRole, nil
	}
	return permissionsupport.RoleCustomDeptIDMapContext(ctx, db, roleIDs)
}

func loadRoleApplicationMenuKeyMapContext(ctx context.Context, db *gorm.DB, list []model.Role) (map[uint][]string, map[uint][]string, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
	}
	roleIDs := make([]uint, 0, len(list))
	for _, item := range list {
		roleIDs = append(roleIDs, item.ID)
	}
	return permissionsupport.RoleApplicationMenuKeyMapContext(ctx, db, roleIDs)
}

func loadRoleApplicationAPIKeyMapContext(ctx context.Context, db *gorm.DB, list []model.Role) (map[uint][]string, map[uint][]string, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
	}
	roleIDs := make([]uint, 0, len(list))
	for _, item := range list {
		roleIDs = append(roleIDs, item.ID)
	}
	return permissionsupport.RoleApplicationAPIKeyMapContext(ctx, db, roleIDs)
}

func ApplicationPermissionTree() ApplicationPermissionTreeResponse {
	return applicationPermissionTreeWithLabels(nil, nil, nil, nil)
}

func ApplicationPermissionTreeContext(ctx context.Context) ApplicationPermissionTreeResponse {
	if ctx == nil {
		ctx = context.Background()
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return ApplicationPermissionTree()
	}
	clientMenuLabels := applicationPermissionLabelsContext(ctx, db, permissionsupport.PlatformClient, permissionsupport.TypeMenu)
	dingtalkH5MenuLabels := applicationPermissionLabelsContext(ctx, db, permissionsupport.PlatformDingTalkH5, permissionsupport.TypeDirectory, permissionsupport.TypeMenu, permissionsupport.TypeButton)
	clientAPILabels := applicationPermissionLabelsContext(ctx, db, permissionsupport.PlatformClient, permissionsupport.TypeAPI, permissionsupport.TypeAPICategory)
	dingtalkH5APILabels := applicationPermissionLabelsContext(ctx, db, permissionsupport.PlatformDingTalkH5, permissionsupport.TypeAPI, permissionsupport.TypeAPICategory)
	return applicationPermissionTreeWithLabels(clientMenuLabels, dingtalkH5MenuLabels, clientAPILabels, dingtalkH5APILabels)
}

func applicationPermissionTreeWithLabels(clientMenuLabels, dingtalkH5MenuLabels, clientAPILabels, dingtalkH5APILabels map[string]string) ApplicationPermissionTreeResponse {
	return ApplicationPermissionTreeResponse{
		Client:        applicationPermissionNodesWithLabels(appmenuperm.ClientMenuDeclarations(), clientMenuLabels),
		DingTalkH5:    applicationPermissionNodesWithLabels(appmenuperm.DingTalkH5PermissionDeclarations(), dingtalkH5MenuLabels),
		ClientAPI:     applicationAPIPermissionNodesWithLabels(appapiperm.ClientAPICategories(), appapiperm.ClientAPIDeclarations(), clientAPILabels),
		DingTalkH5API: applicationAPIPermissionNodesWithLabels(appapiperm.DingTalkH5APICategories(), appapiperm.DingTalkH5APIDeclarations(), dingtalkH5APILabels),
	}
}

func applicationPermissionNodes(declarations []appmenuperm.Declaration) []ApplicationPermissionNode {
	return applicationPermissionNodesWithLabels(declarations, nil)
}

func applicationPermissionNodesWithLabels(declarations []appmenuperm.Declaration, labels map[string]string) []ApplicationPermissionNode {
	declarationByKey := make(map[string]appmenuperm.Declaration, len(declarations))
	for _, declaration := range declarations {
		declarationByKey[declaration.Key] = declaration
	}
	childrenByParent := make(map[string][]appmenuperm.Declaration, len(declarations))
	roots := make([]appmenuperm.Declaration, 0, len(declarations))
	for _, declaration := range declarations {
		if declaration.ParentKey == "" {
			roots = append(roots, declaration)
			continue
		}
		if _, ok := declarationByKey[declaration.ParentKey]; !ok {
			roots = append(roots, declaration)
			continue
		}
		childrenByParent[declaration.ParentKey] = append(childrenByParent[declaration.ParentKey], declaration)
	}
	result := make([]ApplicationPermissionNode, 0, len(roots))
	for _, root := range roots {
		result = append(result, applicationPermissionNodeFromDeclaration(root, childrenByParent, labels))
	}
	return result
}

func applicationPermissionNodeFromDeclaration(declaration appmenuperm.Declaration, childrenByParent map[string][]appmenuperm.Declaration, labels map[string]string) ApplicationPermissionNode {
	node := ApplicationPermissionNode{
		Key:       declaration.Key,
		Name:      applicationPermissionLabel(declaration.Key, declaration.Name, labels),
		ParentKey: declaration.ParentKey,
	}
	if children := childrenByParent[declaration.Key]; len(children) > 0 {
		node.Children = make([]ApplicationPermissionNode, 0, len(children))
		for _, child := range children {
			node.Children = append(node.Children, applicationPermissionNodeFromDeclaration(child, childrenByParent, labels))
		}
	}
	return node
}

func applicationAPIPermissionNodes(categories []appapiperm.Category, declarations []appapiperm.Declaration) []ApplicationPermissionNode {
	return applicationAPIPermissionNodesWithLabels(categories, declarations, nil)
}

func applicationAPIPermissionNodesWithLabels(categories []appapiperm.Category, declarations []appapiperm.Declaration, labels map[string]string) []ApplicationPermissionNode {
	nodeByKey := make(map[string]*ApplicationPermissionNode, len(categories)+len(declarations))
	roots := make([]*ApplicationPermissionNode, 0, len(categories))
	for _, category := range categories {
		node := &ApplicationPermissionNode{
			Key:  category.Key,
			Name: applicationPermissionLabel(category.Key, category.Name, labels),
		}
		nodeByKey[category.Key] = node
		roots = append(roots, node)
	}
	for _, declaration := range declarations {
		node := ApplicationPermissionNode{
			Key:       declaration.Key,
			Name:      applicationPermissionLabel(declaration.Key, declaration.Name, labels),
			ParentKey: declaration.CategoryKey,
		}
		if parent, ok := nodeByKey[declaration.CategoryKey]; ok {
			parent.Children = append(parent.Children, node)
			continue
		}
		orphan := node
		nodeByKey[declaration.Key] = &orphan
		roots = append(roots, &orphan)
	}
	result := make([]ApplicationPermissionNode, 0, len(roots))
	for _, root := range roots {
		result = append(result, *root)
	}
	return result
}

func applicationPermissionLabelsContext(ctx context.Context, db *gorm.DB, platform string, types ...string) map[string]string {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil
	}
	if db == nil || strings.TrimSpace(platform) == "" || len(types) == 0 {
		return nil
	}
	var rows []model.Permission
	if err := db.WithContext(ctx).
		Select("`permission_key`, `permission_name`").
		Where("`permission_platform` = ? AND `permission_type` IN ? AND `permission_status` = 1", platform, types).
		Find(&rows).Error; err != nil {
		return nil
	}
	labels := make(map[string]string, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(row.Key)
		name := strings.TrimSpace(row.Name)
		if key != "" && name != "" {
			labels[key] = name
		}
	}
	return labels
}

func applicationPermissionLabel(key, fallback string, labels map[string]string) string {
	if label := strings.TrimSpace(labels[key]); label != "" {
		return label
	}
	return fallback
}

func Add(name, remark, addIP string, sort, dataScope int) (uint, error) {
	return AddContext(context.Background(), name, remark, addIP, sort, dataScope)
}

func AddContext(ctx context.Context, name, remark, addIP string, sort, dataScope int) (uint, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	role := model.Role{
		Name:            name,
		Remark:          remark,
		Sort:            sort,
		Status:          1,
		AllowAdminLogin: 1,
		DataScope:       dataScope,
		AddTime:         database.Now(),
		AddIP:           addIP,
	}
	if err := db.Create(&role).Error; err != nil {
		return 0, err
	}
	return role.ID, nil
}

func AddWithAssignmentsContext(ctx context.Context, name, remark, addIP string, sort, dataScope, allowAdminLogin int, adminPermissionKeys, adminAPIPermissionKeys []string, deptIDs []uint, clientMenuKeys, dingtalkH5MenuKeys []string, clientAPIPermissionKeys, dingtalkH5APIPermissionKeys []string) (uint, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	role := model.Role{
		Name:            name,
		Remark:          remark,
		Sort:            sort,
		Status:          1,
		AllowAdminLogin: normalizeAllowAdminLogin(allowAdminLogin),
		DataScope:       dataScope,
		AddTime:         database.Now(),
		AddIP:           addIP,
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&role).Error; err != nil {
			return err
		}
		if err := permissionsupport.SetRoleAdminPermissionKeysTx(tx, role.ID, normalizeAllowAdminLogin(allowAdminLogin), adminPermissionKeys, adminAPIPermissionKeys, dataScope, deptIDs); err != nil {
			return err
		}
		if err := permissionsupport.SetRoleApplicationMenuPermissionsTx(tx, role.ID, clientMenuKeys, dingtalkH5MenuKeys); err != nil {
			return err
		}
		return permissionsupport.SetRoleApplicationAPIPermissionsTx(tx, role.ID, clientAPIPermissionKeys, dingtalkH5APIPermissionKeys)
	})
	if err != nil {
		return 0, err
	}
	menuservice.InvalidateAdminPermCacheForRole(role.ID)
	return role.ID, nil
}

func Edit(id uint, name, remark, addIP string, sort, status, dataScope int) error {
	return EditContext(context.Background(), id, name, remark, addIP, sort, status, dataScope)
}

func EditContext(ctx context.Context, id uint, name, remark, addIP string, sort, status, dataScope int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	updates := map[string]interface{}{
		"role_name":       name,
		"role_remark":     remark,
		"role_sort":       sort,
		"role_status":     status,
		"role_data_scope": dataScope,
		"role_edit_time":  database.Now(),
		"role_edit_ip":    addIP,
	}
	err := db.Model(&model.Role{}).Where("`id` = ?", id).Updates(updates).Error
	if err == nil {
		adminaccess.InvalidateAdminAccessCacheForRole(id)
	}
	return err
}

func EditWithAssignmentsContext(ctx context.Context, id uint, name, remark, addIP string, sort, status, dataScope, allowAdminLogin int, adminPermissionKeys, adminAPIPermissionKeys []string, deptIDs []uint, clientMenuKeys, dingtalkH5MenuKeys []string, clientAPIPermissionKeys, dingtalkH5APIPermissionKeys []string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	updates := map[string]interface{}{
		"role_name":       name,
		"role_remark":     remark,
		"role_sort":       sort,
		"role_status":     status,
		"role_data_scope": dataScope,
		"role_edit_time":  database.Now(),
		"role_edit_ip":    addIP,
	}
	if allowAdminLogin >= 0 {
		updates["role_allow_admin_login"] = normalizeAllowAdminLogin(allowAdminLogin)
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Role{}).Where("`id` = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		loginAllowed := allowAdminLogin
		if loginAllowed < 0 {
			var role model.Role
			if err := tx.Select("role_allow_admin_login").Where("`id` = ?", id).First(&role).Error; err == nil {
				loginAllowed = role.AllowAdminLogin
			}
		}
		if err := permissionsupport.SetRoleAdminPermissionKeysTx(tx, id, normalizeAllowAdminLogin(loginAllowed), adminPermissionKeys, adminAPIPermissionKeys, dataScope, deptIDs); err != nil {
			return err
		}
		if err := permissionsupport.SetRoleApplicationMenuPermissionsTx(tx, id, clientMenuKeys, dingtalkH5MenuKeys); err != nil {
			return err
		}
		return permissionsupport.SetRoleApplicationAPIPermissionsTx(tx, id, clientAPIPermissionKeys, dingtalkH5APIPermissionKeys)
	})
	if err == nil {
		menuservice.InvalidateAdminPermCacheForRole(id)
		adminaccess.InvalidateAdminAccessCacheForRole(id)
		permissionsupport.InvalidateRuntimePermissionCaches()
	}
	return err
}

func normalizeAllowAdminLogin(value int) int {
	if value == 0 {
		return 0
	}
	return 1
}

func Delete(id uint) error {
	return DeleteContext(context.Background(), id)
}

func DeleteContext(ctx context.Context, id uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("`grant_subject_type` = ? AND `grant_subject_id` = ?", permissionsupport.SubjectRole, id).Delete(&model.PermissionGrant{}).Error; err != nil {
			return err
		}
		return tx.Where("`id` = ?", id).Delete(&model.Role{}).Error
	})
	if err == nil {
		menuservice.InvalidateAdminPermCacheForRole(id)
		adminaccess.InvalidateAdminAccessCacheForRole(id)
		permissionsupport.InvalidateRuntimePermissionCaches()
	}
	return err
}

func BatchDelete(ids []uint) error {
	return BatchDeleteContext(context.Background(), ids)
}

func BatchDeleteContext(ctx context.Context, ids []uint) error {
	for _, id := range ids {
		if err := DeleteContext(ctx, id); err != nil {
			return err
		}
	}
	return nil
}
