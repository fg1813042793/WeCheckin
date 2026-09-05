package permission

import (
	"errors"
	"sync"
	"time"
)

const (
	SubjectRole = "role"
	SubjectUser = "user"

	EffectAllow = "allow"
	EffectDeny  = "deny"

	PlatformAdmin      = "admin"
	PlatformClient     = "client"
	PlatformDingTalkH5 = "dingtalk_h5"
	PlatformData       = "data"

	TypeLogin       = "login"
	TypeDirectory   = "directory"
	TypeMenu        = "menu"
	TypeButton      = "button"
	TypeAPI         = "api"
	TypeAPICategory = "api_category"
	TypeData        = "data"

	AdminLoginPermissionKey = "admin:login"
	DataAllPermissionKey    = "data:all"
	DataDeptPermissionKey   = "data:dept"
	DataSelfPermissionKey   = "data:self"
	DataCustomPermissionKey = "data:custom"
	DataExtraPermissionKey  = "data:extra"
)

var ErrPermissionSchemaNotReady = errors.New("权限数据结构尚未初始化，请先执行数据库迁移")

const (
	dingtalkH5MenuPermissionCacheTTL           = 30 * time.Second
	subjectPermissionSetCacheTTL               = 30 * time.Second
	permissionTablesReadyNegativeCacheTTL      = 5 * time.Second
	userRolesTableReadyNegativeCacheTTL        = 5 * time.Second
	permissionGrantRoleAssignmentSelectColumns = "`grant_subject_id`, `grant_permission_key`, `grant_scope_value`"
	permissionGrantKeySelectColumns            = "`grant_subject_id`, `grant_permission_key`, `grant_effect`"
	permissionGrantScopeSelectColumns          = "`grant_subject_id`, `grant_scope_value`"
)

type dingtalkH5MenuPermissionCacheEntry struct {
	keys      []string
	ready     bool
	expiresAt time.Time
}

var dingtalkH5MenuPermissionCache = struct {
	sync.RWMutex
	items map[string]dingtalkH5MenuPermissionCacheEntry
}{
	items: map[string]dingtalkH5MenuPermissionCacheEntry{},
}

type subjectPermissionSetCacheEntry struct {
	allowed   map[string]bool
	denied    map[string]bool
	expiresAt time.Time
}

var subjectPermissionSetCache = struct {
	sync.RWMutex
	items map[string]subjectPermissionSetCacheEntry
}{
	items: map[string]subjectPermissionSetCacheEntry{},
}

var permissionTablesReadyCache = struct {
	sync.RWMutex
	checked     bool
	ready       bool
	schemaReady bool
	checkedAt   time.Time
}{}

var userRolesTableReadyCache = struct {
	sync.RWMutex
	checked   bool
	ready     bool
	checkedAt time.Time
}{}

type DataScope struct {
	Mode    int
	DeptIDs []uint
	Ready   bool
}

type DataScopeExtras struct {
	DeptIDs []uint
	UserIDs []uint
	Ready   bool
}

type permissionSubjectRef struct {
	subjectType string
	subjectID   uint
}

type RoleAssignmentMaps struct {
	AdminPermissionKeys         map[uint][]string
	AdminAPIPermissionKeys      map[uint][]string
	DeptIDs                     map[uint][]uint
	ClientMenuKeys              map[uint][]string
	DingTalkH5MenuKeys          map[uint][]string
	ClientAPIPermissionKeys     map[uint][]string
	DingTalkH5APIPermissionKeys map[uint][]string
}
