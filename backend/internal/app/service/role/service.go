package role

import (
	"context"

	"gorm.io/gorm"

	menuservice "wecheckin-backend/backend/internal/app/service/menu"
	"wecheckin-backend/backend/internal/app/support/access"
	"wecheckin-backend/backend/internal/app/support/appmenuperm"
	permissionsupport "wecheckin-backend/backend/internal/app/support/permission"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

type ListItem struct {
	ID                     uint     `json:"id"`
	Name                   string   `json:"name"`
	Remark                 string   `json:"remark"`
	Sort                   int      `json:"sort"`
	Status                 int      `json:"status"`
	AllowAdminLogin        int      `json:"allowAdminLogin"`
	DataScope              int      `json:"dataScope"`
	AddTime                int64    `json:"addTime"`
	EditTime               int64    `json:"editTime"`
	AdminPermissionKeys    []string `json:"adminPermissionKeys"`
	AdminAPIPermissionKeys []string `json:"adminApiPermissionKeys"`
	DeptIDs                []uint   `json:"deptIds"`
	ClientMenuKeys         []string `json:"clientMenuKeys"`
	DingTalkH5MenuKeys     []string `json:"dingtalkH5MenuKeys"`
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
	Client     []ApplicationPermissionNode `json:"client"`
	DingTalkH5 []ApplicationPermissionNode `json:"dingtalkH5"`
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
	if admin.Type != 1 && admin.RoleID > 0 {
		var role model.Role
		if err := db.First(&role, admin.RoleID).Error; err == nil {
			if role.DataScope == 2 || role.DataScope == 4 {
				var deptIDs []uint
				if role.DataScope == 2 {
					deptIDs = access.AdminDeptIDsContext(ctx, admin.ID)
				} else {
					deptIDs, _ = permissionsupport.RoleCustomDeptIDsContext(ctx, db, admin.RoleID)
				}
				if len(deptIDs) > 0 {
					roleIDs, _ := permissionsupport.RoleIDsByCustomDeptIDsContext(ctx, db, deptIDs)
					rid := admin.RoleID
					conditions = append(conditions, func(d *gorm.DB) *gorm.DB {
						if len(roleIDs) == 0 {
							return d.Where("`id` = ?", rid)
						}
						return d.Where("`id` IN ? OR `id` = ?", roleIDs, rid)
					})
				}
			} else if role.DataScope == 3 {
				rid := admin.RoleID
				conditions = append(conditions, func(d *gorm.DB) *gorm.DB {
					return d.Where("`id` = ?", rid)
				})
			}
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
	adminPermissionKeysByRole, err := loadRoleAdminPermissionKeyMapContext(ctx, db, list)
	if err != nil {
		return nil, err
	}
	adminAPIPermissionKeysByRole, err := loadRoleAdminAPIPermissionKeyMapContext(ctx, db, list)
	if err != nil {
		return nil, err
	}
	deptIDsByRole, err := loadRoleDeptIDMapContext(ctx, db, list)
	if err != nil {
		return nil, err
	}
	clientMenuKeysByRole, dingtalkH5MenuKeysByRole, err := loadRoleApplicationMenuKeyMapContext(ctx, db, list)
	if err != nil {
		return nil, err
	}
	result := make([]ListItem, len(list))
	for i, r := range list {
		result[i] = ListItem{
			ID:                     r.ID,
			Name:                   r.Name,
			Remark:                 r.Remark,
			Sort:                   r.Sort,
			Status:                 r.Status,
			AllowAdminLogin:        normalizeAllowAdminLogin(r.AllowAdminLogin),
			DataScope:              r.DataScope,
			AddTime:                r.AddTime,
			EditTime:               r.EditTime,
			AdminPermissionKeys:    adminPermissionKeysByRole[r.ID],
			AdminAPIPermissionKeys: adminAPIPermissionKeysByRole[r.ID],
			DeptIDs:                deptIDsByRole[r.ID],
			ClientMenuKeys:         clientMenuKeysByRole[r.ID],
			DingTalkH5MenuKeys:     dingtalkH5MenuKeysByRole[r.ID],
		}
	}
	return &ListResponse{List: result, Total: total}, nil
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

func ApplicationPermissionTree() ApplicationPermissionTreeResponse {
	return ApplicationPermissionTreeResponse{
		Client:     applicationPermissionNodes(appmenuperm.ClientMenuDeclarations()),
		DingTalkH5: applicationPermissionNodes(appmenuperm.DingTalkH5MenuDeclarations()),
	}
}

func applicationPermissionNodes(declarations []appmenuperm.Declaration) []ApplicationPermissionNode {
	nodeByKey := make(map[string]*ApplicationPermissionNode, len(declarations))
	roots := make([]*ApplicationPermissionNode, 0, len(declarations))
	for _, declaration := range declarations {
		nodeByKey[declaration.Key] = &ApplicationPermissionNode{
			Key:       declaration.Key,
			Name:      declaration.Name,
			ParentKey: declaration.ParentKey,
		}
	}
	for _, declaration := range declarations {
		node := nodeByKey[declaration.Key]
		if declaration.ParentKey == "" {
			roots = append(roots, node)
			continue
		}
		parent, ok := nodeByKey[declaration.ParentKey]
		if !ok {
			roots = append(roots, node)
			continue
		}
		parent.Children = append(parent.Children, *node)
	}
	result := make([]ApplicationPermissionNode, 0, len(roots))
	for _, root := range roots {
		result = append(result, *root)
	}
	return result
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

func AddWithAssignmentsContext(ctx context.Context, name, remark, addIP string, sort, dataScope, allowAdminLogin int, adminPermissionKeys, adminAPIPermissionKeys []string, deptIDs []uint, clientMenuKeys, dingtalkH5MenuKeys []string) (uint, error) {
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
		return permissionsupport.SetRoleApplicationMenuPermissionsTx(tx, role.ID, clientMenuKeys, dingtalkH5MenuKeys)
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
	return db.Model(&model.Role{}).Where("`id` = ?", id).Updates(updates).Error
}

func EditWithAssignmentsContext(ctx context.Context, id uint, name, remark, addIP string, sort, status, dataScope, allowAdminLogin int, adminPermissionKeys, adminAPIPermissionKeys []string, deptIDs []uint, clientMenuKeys, dingtalkH5MenuKeys []string) error {
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
		return permissionsupport.SetRoleApplicationMenuPermissionsTx(tx, id, clientMenuKeys, dingtalkH5MenuKeys)
	})
	if err == nil {
		menuservice.InvalidateAdminPermCacheForRole(id)
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
