package service

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
	rd "wecheckin-backend/backend/pkg/redis"
	"wecheckin-backend/backend/pkg/tokenutil"
	"gorm.io/gorm"
)

func AdminHome(adminID uint) (map[string]interface{}, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)

	var userCnt int64
	if admin.Type == 1 || admin.RoleID == 0 {
		database.DB.Model(&model.User{}).Count(&userCnt)
	} else {
		q := database.DB.Model(&model.User{})
		var role model.Role
		if err := database.DB.First(&role, admin.RoleID).Error; err == nil {
			if role.DataScope == 2 || role.DataScope == 4 {
				var deptIDs []uint
				if role.DataScope == 2 {
					deptIDs = getAdminDeptIDs(admin.ID)
				} else {
					deptIDs = GetRoleDeptIDs(admin.RoleID)
				}
				if len(deptIDs) > 0 {
					q = q.Where("`id` IN (SELECT `user_dept_user_id` FROM `user_depts` WHERE `user_dept_dept_id` IN ?)", deptIDs)
				}
			} else if role.DataScope == 3 {
				q = q.Where("1 = 0")
			}
		}
		q.Count(&userCnt)
	}

	var enrollCnt int64
	q := database.DB.Model(&model.Enroll{})
	where, args := BuildDataScopeFilter(&admin, "`enroll_dept_id`", "`enroll_create_by`")
	if where != "" {
		q = q.Where(where, args...)
	}
	q.Count(&enrollCnt)

	var newsCnt int64
	q2 := database.DB.Model(&model.News{})
	where2, args2 := BuildDataScopeFilter(&admin, "`news_dept_id`", "`news_create_by`")
	if where2 != "" {
		q2 = q2.Where(where2, args2...)
	}
	q2.Count(&newsCnt)

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999, now.Location()).UnixMilli()
	var joinCnt int64
	q3 := database.DB.Model(&model.EnrollJoin{}).Where("`enroll_join_add_time` BETWEEN ? AND ?", todayStart, todayEnd)
	// Filter join counts by enrolls the admin can see
	if where != "" {
		q3 = q3.Where("`enroll_join_enroll_id` IN (SELECT `id` FROM `enrolls` WHERE "+where+")", args...)
	}
	q3.Count(&joinCnt)

	var eventCnt int64
	q4 := database.DB.Model(&model.Event{})
	where4, args4 := BuildDataScopeFilter(&admin, "`event_dept_id`", "`event_create_by`")
	if where4 != "" {
		q4 = q4.Where(where4, args4...)
	}
	q4.Count(&eventCnt)

	var eventUserCnt int64
	database.DB.Model(&model.EventParticipant{}).Count(&eventUserCnt)

	var mgrCnt int64
	database.DB.Model(&model.Admin{}).Count(&mgrCnt)

	result := map[string]interface{}{
		"userCnt":      userCnt,
		"enrollCnt":    enrollCnt,
		"newsCnt":      newsCnt,
		"joinCnt":      joinCnt,
		"eventCnt":     eventCnt,
		"eventUserCnt": eventUserCnt,
		"mgrCnt":       mgrCnt,
	}
	return result, nil
}

func ClearVouchData() error {
	return database.DB.Model(&model.Enroll{}).Where("1 = 1").Update("enroll_vouch", 0).Error
}

// genRandomString 使用 crypto/rand 生成 length 个十六进制字符（length 必须为偶数）。
// 熵为 length*4 bit；长度 32 即 128 bit，远高于 UUID v4。
// 熵源失败时 panic（系统熵不可用 = 整个服务都不可信）。
func genRandomString(length int) string {
	if length <= 0 || length%2 != 0 {
		panic(fmt.Sprintf("genRandomString: length must be a positive even number, got %d", length))
	}
	b := make([]byte, length/2)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("genRandomString: crypto/rand failed: %v", err))
	}
	return hex.EncodeToString(b)
}

func InsertLog(logType int, content, adminID, adminName, adminDesc, addIP string) {
	database.DB.Create(&model.Log{
		Type:      logType,
		Content:   content,
		AdminID:   adminID,
		AdminName: adminName,
		AdminDesc: adminDesc,
		AddTime:   database.Now(),
		AddIP:     addIP,
	})
}

func AdminLogin(name, password, addIP, device string) (map[string]interface{}, error) {
	h := md5.Sum([]byte(password))
	passwordMD5 := hex.EncodeToString(h[:])
	var admin model.Admin
	err := database.DB.Where("`admin_name` = ? AND `admin_password` = ?", name, passwordMD5).First(&admin).Error
	if err != nil {
		return nil, fmt.Errorf("账号或密码错误")
	}
	if admin.Status != 1 {
		return nil, fmt.Errorf("账号已禁用")
	}
	token := genRandomString(32)
	admin.LoginCnt++
	admin.LoginTime = database.Now()
	database.DB.Model(&admin).Updates(map[string]interface{}{
		"admin_login_cnt":  admin.LoginCnt,
		"admin_login_time": admin.LoginTime,
	})
	roleName := ""
	if admin.RoleID > 0 {
		var role model.Role
		if err := database.DB.First(&role, admin.RoleID).Error; err == nil {
			roleName = role.Name
		}
	}
	storeAdminToken(&admin, token, addIP, device, roleName)
	InsertLog(1, "管理员登录", strconv.Itoa(int(admin.ID)), admin.Name, admin.Desc, addIP)
	var dataScope int
	if admin.RoleID > 0 {
		var role model.Role
		if err := database.DB.First(&role, admin.RoleID).Error; err == nil {
			dataScope = role.DataScope
		}
	}
	result := map[string]interface{}{
		"token":     token,
		"name":      admin.Name,
		"pic":       GetFullURL(admin.Pic),
		"id":        admin.ID,
		"type":      admin.Type,
		"roleId":    admin.RoleID,
		"dataScope": dataScope,
		"loginCnt":  admin.LoginCnt,
	}
	return result, nil
}

func storeAdminToken(admin *model.Admin, token, addIP, device, roleName string) {
	expire, prefix := tokenutil.GetTokenConfig("admin")
	now := database.Now()
	database.DB.Model(&model.Admin{}).Where("`id` = ?", admin.ID).Updates(map[string]interface{}{
		"admin_token":      token,
		"admin_token_time": now,
	})
	if rd.RDB != nil {
		keyAuth := prefix + "a:" + token
		idStr := strconv.Itoa(int(admin.ID))
		keySet := prefix + "s:" + idStr

		// 单端登录模式：踢掉同 adminID 的所有旧 token
		if tokenutil.IsAdminSingleLogin() {
			if oldTokens, _ := rd.RDB.SMembers(rd.Ctx, keySet).Result(); len(oldTokens) > 0 {
				for _, t := range oldTokens {
					if t != token {
						rd.RDB.Del(rd.Ctx, prefix+"a:"+t)
					}
				}
				rd.RDB.Del(rd.Ctx, keySet)
			}
		}

		info := map[string]interface{}{
			"id":        admin.ID,
			"name":      admin.Name,
			"type":      admin.Type,
			"roleId":    admin.RoleID,
			"roleName":  roleName,
			"desc":      admin.Desc,
			"loginIp":   addIP,
			"loginTime": now,
			"device":    device,
		}
		jsonBytes, _ := json.Marshal(info)
		// Set TTL on s: 2x expire to keep it alive through long sessions
		// (a: slides on every request; s: only updates on token add/remove)
		rd.RDB.Set(rd.Ctx, keyAuth, string(jsonBytes), expire)
		rd.RDB.SAdd(rd.Ctx, keySet, token)
		rd.RDB.Expire(rd.Ctx, keySet, expire*2)
	}
}

func GetMgrList(adminID uint, keyword string, page, pageSize int) (map[string]interface{}, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)
	var conditions []func(*gorm.DB) *gorm.DB
	// Data scope filter for admins
	if admin.Type != 1 && admin.RoleID > 0 {
		var role model.Role
		if err := database.DB.First(&role, admin.RoleID).Error; err == nil {
			if role.DataScope == 2 || role.DataScope == 4 {
				var deptIDs []uint
				if role.DataScope == 2 {
					deptIDs = getAdminDeptIDs(admin.ID)
				} else {
					deptIDs = GetRoleDeptIDs(admin.RoleID)
				}
				if len(deptIDs) > 0 {
					conditions = append(conditions, func(d *gorm.DB) *gorm.DB {
						return d.Where("`id` IN (SELECT `admin_dept_admin_id` FROM `admin_depts` WHERE `admin_dept_dept_id` IN ?)", deptIDs)
					})
				}
			} else if role.DataScope == 3 {
				id := admin.ID
				conditions = append(conditions, func(d *gorm.DB) *gorm.DB {
					return d.Where("`id` = ?", id)
				})
			}
		}
	}
	if keyword != "" {
		kw := keyword
		conditions = append(conditions, func(d *gorm.DB) *gorm.DB {
			return d.Where("`admin_name` LIKE ? OR `admin_phone` LIKE ?", "%"+kw+"%", "%"+kw+"%")
		})
	}
	var total int64
	database.DB.Model(&model.Admin{}).Scopes(conditions...).Count(&total)
	var list []model.Admin
	database.DB.Model(&model.Admin{}).Scopes(conditions...).Order("`admin_add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	result := make([]map[string]interface{}, len(list))
	for i, a := range list {
		roleName := ""
		if a.RoleID > 0 {
			var role model.Role
			if err := database.DB.First(&role, a.RoleID).Error; err == nil {
				roleName = role.Name
			}
		}
		result[i] = map[string]interface{}{
			"id": a.ID, "name": a.Name, "desc": a.Desc, "pic": GetFullURL(a.Pic),
			"phone": a.Phone, "status": a.Status, "type": a.Type,
			"roleId": a.RoleID, "roleName": roleName, "loginCnt": a.LoginCnt,
			"addTime": a.AddTime, "editTime": a.EditTime,
			"deptIds": getAdminDeptIDs(a.ID),
		}
	}
	return map[string]interface{}{"list": result, "total": total}, nil
}

func getAdminDeptIDs(adminID uint) []uint {
	var list []model.AdminDept
	database.DB.Where("`admin_dept_admin_id` = ?", adminID).Find(&list)
	ids := make([]uint, len(list))
	for i, d := range list {
		ids[i] = d.DeptID
	}
	return ids
}

func saveAdminDepts(adminID uint, deptIDs []uint) {
	database.DB.Where("`admin_dept_admin_id` = ?", adminID).Delete(&model.AdminDept{})
	for _, deptID := range deptIDs {
		if deptID > 0 {
			database.DB.Create(&model.AdminDept{AdminID: adminID, DeptID: deptID})
		}
	}
}

func InsertMgr(name, password, desc, phone, addIP string, typ int, roleID uint, deptIDs []uint) error {
	var cnt int64
	database.DB.Model(&model.Admin{}).Where("`admin_name` = ?", name).Count(&cnt)
	if cnt > 0 {
		return fmt.Errorf("管理员已存在")
	}
	h := md5.Sum([]byte(password))
	admin := model.Admin{
		Name:     name,
		Password: hex.EncodeToString(h[:]),
		Desc:     desc,
		Phone:    phone,
		Status:   1,
		Type:     typ,
		RoleID:   roleID,
		AddTime:  database.Now(),
		AddIP:    addIP,
	}
	if err := database.DB.Create(&admin).Error; err != nil {
		return err
	}
	saveAdminDepts(admin.ID, deptIDs)
	return nil
}

func DelMgr(id string) error {
	var admin model.Admin
	if err := database.DB.Where("`id` = ?", id).First(&admin).Error; err != nil {
		return err
	}
	if admin.Type == 1 {
		return fmt.Errorf("超级管理员不可删除")
	}
	ForceOfflineAdmin(id, "")
	database.DB.Where("`admin_dept_admin_id` = ?", id).Delete(&model.AdminDept{})
	return database.DB.Where("`id` = ?", id).Delete(&model.Admin{}).Error
}

func DelMgrs(ids []string) error {
	for _, id := range ids {
		if err := DelMgr(id); err != nil {
			return err
		}
	}
	return nil
}

func GetMgrDetail(id string) (map[string]interface{}, error) {
	var admin model.Admin
	err := database.DB.Where("`id` = ?", id).First(&admin).Error
	if err != nil {
		return nil, err
	}
	uid, _ := strconv.Atoi(id)
	return map[string]interface{}{
		"id": admin.ID, "name": admin.Name, "desc": admin.Desc,
		"pic": GetFullURL(admin.Pic), "phone": admin.Phone, "status": admin.Status,
		"type": admin.Type, "roleId": admin.RoleID,
		"loginCnt": admin.LoginCnt,
		"addTime": admin.AddTime, "editTime": admin.EditTime,
		"deptIds": getAdminDeptIDs(uint(uid)),
	}, nil
}

func EditMgr(id, name, desc, pic, phone, password, addIP string, roleID uint, deptIDs []uint) error {
	updates := map[string]interface{}{
		"admin_name":  name,
		"admin_desc":  desc,
		"admin_phone": phone,
		"admin_role_id": roleID,
		"admin_edit_time": database.Now(),
		"admin_edit_ip":   addIP,
	}
	if pic != "" {
		updates["admin_pic"] = pic
	}
	if password != "" {
		h := md5.Sum([]byte(password))
		updates["admin_password"] = hex.EncodeToString(h[:])
	}
	if err := database.DB.Model(&model.Admin{}).Where("`id` = ?", id).Updates(updates).Error; err != nil {
		return err
	}
	uid, _ := strconv.Atoi(id)
	saveAdminDepts(uint(uid), deptIDs)
	return nil
}

func StatusMgr(id string, status int) error {
	err := database.DB.Model(&model.Admin{}).Where("`id` = ?", id).Update("admin_status", status).Error
	if err == nil && status != 1 {
		ForceOfflineAdmin(id, "")
	}
	return err
}

func PwdMgr(id, oldPassword, newPassword string) error {
	var admin model.Admin
	err := database.DB.Where("`id` = ?", id).First(&admin).Error
	if err != nil {
		return fmt.Errorf("管理员不存在")
	}
	h := md5.Sum([]byte(oldPassword))
	if hex.EncodeToString(h[:]) != admin.Password {
		return fmt.Errorf("旧密码错误")
	}
	h2 := md5.Sum([]byte(newPassword))
	return database.DB.Model(&model.Admin{}).Where("`id` = ?", id).Update("admin_password", hex.EncodeToString(h2[:])).Error
}

func GetLogList(keyword string, page, pageSize int, adminID uint) ([]model.Log, int64, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)
	var list []model.Log
	var total int64
	query := database.DB.Model(&model.Log{})
	if keyword != "" {
		query = query.Where("`log_content` LIKE ? OR `log_admin_name` LIKE ? OR `log_admin_desc` LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	// Data scope
	if admin.Type != 1 && admin.RoleID > 0 {
		var role model.Role
		if err := database.DB.First(&role, admin.RoleID).Error; err == nil {
			if role.DataScope == 2 || role.DataScope == 4 {
				var deptIDs []uint
				if role.DataScope == 2 {
					deptIDs = getAdminDeptIDs(admin.ID)
				} else {
					deptIDs = GetRoleDeptIDs(admin.RoleID)
				}
				if len(deptIDs) > 0 {
					query = query.Where("`log_admin_id` IN (SELECT `admin_dept_admin_id` FROM `admin_depts` WHERE `admin_dept_dept_id` IN ?)", deptIDs)
				}
			} else if role.DataScope == 3 {
				query = query.Where("`log_admin_id` = ?", admin.ID)
			}
		}
	}
	query.Count(&total)
	err := query.Order("`log_add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func ClearLog() error {
	return database.DB.Where("1 = 1").Delete(&model.Log{}).Error
}

func GetOnlineUsers() ([]map[string]interface{}, error) {
	_, prefix := tokenutil.GetTokenConfig("user")
	if rd.RDB == nil {
		return []map[string]interface{}{}, nil
	}
	setPrefix := prefix + "s:"
	authPrefix := prefix + "a:"

	entries, err := scanOnlineSets(setPrefix)
	if err != nil {
		return nil, err
	}
	loadBase := preloadUserBase(entries)
	return buildOnlineRows(entries, authPrefix, loadBase), nil
}

// preloadUserBase fetches all user records in 1 batched query and returns a lookup closure.
func preloadUserBase(entries []onlineEntry) func(uint64) (map[string]interface{}, bool) {
	uids := make([]uint, 0, len(entries))
	for _, e := range entries {
		uids = append(uids, uint(e.uid))
	}
	var users []model.User
	if len(uids) > 0 {
		database.DB.Where("id IN ?", uids).Find(&users)
	}
	userByID := make(map[uint]*model.User, len(users))
	for i := range users {
		userByID[users[i].ID] = &users[i]
	}

	return func(uid uint64) (map[string]interface{}, bool) {
		u, ok := userByID[uint(uid)]
		if !ok {
			return nil, false
		}
		return map[string]interface{}{
			"id":       u.ID,
			"name":     u.Name,
			"mobile":   u.Mobile,
			"pic":      GetFullURL(u.Pic),
			"loginCnt": u.LoginCnt,
		}, true
	}
}

// onlineEntry is a (setKey, uid, tokens) tuple built from Redis SCAN+SMEMBERS.
type onlineEntry struct {
	setKey string
	uid    uint64
	tokens []string
}

// scanOnlineSets SCANs all `s:` Set keys and returns per-user entries with
// their current token members. No DB or per-token I/O.
func scanOnlineSets(setPrefix string) ([]onlineEntry, error) {
	var cursor uint64
	var setKeys []string
	for {
		ks, c, err := rd.RDB.Scan(rd.Ctx, cursor, setPrefix+"*", 100).Result()
		if err != nil {
			return nil, err
		}
		setKeys = append(setKeys, ks...)
		cursor = c
		if cursor == 0 {
			break
		}
	}

	entries := make([]onlineEntry, 0, len(setKeys))
	for _, setKey := range setKeys {
		idStr := strings.TrimPrefix(setKey, setPrefix)
		uid, _ := strconv.ParseUint(idStr, 10, 64)
		if uid == 0 {
			continue
		}
		tokens, _ := rd.RDB.SMembers(rd.Ctx, setKey).Result()
		if len(tokens) == 0 {
			continue
		}
		entries = append(entries, onlineEntry{setKey, uid, tokens})
	}
	return entries, nil
}

// buildOnlineRows takes the entries from scanOnlineSets, fetches per-token info
// in a single pipelined round trip, joins with the per-user base info from
// `loadBase`, and prunes dead token references from Sets.
func buildOnlineRows(entries []onlineEntry, authPrefix string, loadBase func(uid uint64) (map[string]interface{}, bool)) []map[string]interface{} {
	// Pipeline: for each token, GET a:{token} + TTL a:{token}
	pipe := rd.RDB.Pipeline()
	type cmd struct {
		token string
		get   *redis.StringCmd
		ttl   *redis.DurationCmd
	}
	allCmds := make([]cmd, 0)
	for _, e := range entries {
		for _, t := range e.tokens {
			allCmds = append(allCmds, cmd{
				token: t,
				get:   pipe.Get(rd.Ctx, authPrefix+t),
				ttl:   pipe.TTL(rd.Ctx, authPrefix+t),
			})
		}
	}
	if len(allCmds) > 0 {
		_, _ = pipe.Exec(rd.Ctx)
	}

	result := make([]map[string]interface{}, 0)
	idx := 0
	for _, e := range entries {
		base, ok := loadBase(e.uid)
		if !ok {
			idx += len(e.tokens)
			continue
		}
		var deadTokens []string
		for _, t := range e.tokens {
			jsonStr, err := allCmds[idx].get.Result()
			ttl, _ := allCmds[idx].ttl.Result()
			idx++
			if err != nil {
				deadTokens = append(deadTokens, t)
				continue
			}
			var info struct {
				LoginIP   string `json:"loginIp"`
				LoginTime int64  `json:"loginTime"`
				Device    string `json:"device"`
			}
			row := map[string]interface{}{}
			for k, v := range base {
				row[k] = v
			}
			row["token"] = t
			row["ttl"] = int(ttl.Seconds())
			if json.Unmarshal([]byte(jsonStr), &info) == nil {
				row["loginIp"] = info.LoginIP
				row["loginTime"] = info.LoginTime
				row["device"] = info.Device
			} else {
				row["loginIp"] = ""
				row["loginTime"] = int64(0)
				row["device"] = ""
			}
			result = append(result, row)
		}
		if len(deadTokens) > 0 {
			rd.RDB.SRem(rd.Ctx, e.setKey, anyToIface(deadTokens)...)
		}
	}
	return result
}

func anyToIface(ss []string) []interface{} {
	out := make([]interface{}, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

func ForceOfflineUser(idStr, token string) error {
	_, prefix := tokenutil.GetTokenConfig("user")
	if rd.RDB == nil || token == "" {
		return nil
	}
	rd.RDB.Del(rd.Ctx, prefix+"a:"+token)
	rd.RDB.SRem(rd.Ctx, prefix+"s:"+idStr, token)
	if count, _ := rd.RDB.SCard(rd.Ctx, prefix+"s:"+idStr).Result(); count == 0 {
		rd.RDB.Del(rd.Ctx, prefix+"s:"+idStr)
	}
	return nil
}

// BatchForceOfflineUser 批量踢人。items = [{idStr, token}, ...]。
// 用 Redis pipeline 一次完成所有 SREM + DEL，对每个用户最后 SCard==0 时再 DEL Set。
func BatchForceOfflineUser(items []struct {
	IDStr string `json:"idStr"`
	Token string `json:"token"`
}) (int, error) {
	_, prefix := tokenutil.GetTokenConfig("user")
	if rd.RDB == nil || len(items) == 0 {
		return 0, nil
	}
	// Group tokens by user id (one user may have multiple devices selected)
	byID := make(map[string][]string, len(items))
	for _, it := range items {
		if it.Token == "" {
			continue
		}
		byID[it.IDStr] = append(byID[it.IDStr], it.Token)
	}

	pipe := rd.RDB.Pipeline()
	for idStr, tokens := range byID {
		authKeys := make([]string, len(tokens))
		for i, t := range tokens {
			authKeys[i] = prefix + "a:" + t
		}
		pipe.Del(rd.Ctx, authKeys...)
		pipe.SRem(rd.Ctx, prefix+"s:"+idStr, anyToIface(tokens)...)
	}
	if _, err := pipe.Exec(rd.Ctx); err != nil && err != redis.Nil {
		return 0, err
	}
	// 如果 Set 变空就 DEL（清理空 Set）
	for idStr := range byID {
		setKey := prefix + "s:" + idStr
		if n, _ := rd.RDB.SCard(rd.Ctx, setKey).Result(); n == 0 {
			rd.RDB.Del(rd.Ctx, setKey)
		}
	}
	return len(items), nil
}

func GetOnlineAdmins() ([]map[string]interface{}, error) {
	_, prefix := tokenutil.GetTokenConfig("admin")
	if rd.RDB == nil {
		return []map[string]interface{}{}, nil
	}
	setPrefix := prefix + "s:"
	authPrefix := prefix + "a:"

	entries, err := scanOnlineSets(setPrefix)
	if err != nil {
		return nil, err
	}
	loadBase := preloadAdminBase(entries)
	return buildOnlineRows(entries, authPrefix, loadBase), nil
}

// preloadAdminBase fetches all admin records and their roles in 2 batched queries
// (instead of 2N queries), and returns a lookup closure.
func preloadAdminBase(entries []onlineEntry) func(uint64) (map[string]interface{}, bool) {
	uids := make([]uint, 0, len(entries))
	for _, e := range entries {
		uids = append(uids, uint(e.uid))
	}
	var admins []model.Admin
	if len(uids) > 0 {
		database.DB.Where("id IN ?", uids).Find(&admins)
	}
	adminByID := make(map[uint]*model.Admin, len(admins))
	for i := range admins {
		adminByID[admins[i].ID] = &admins[i]
	}

	roleIDs := make([]uint, 0, len(admins))
	for _, a := range admins {
		if a.RoleID > 0 {
			roleIDs = append(roleIDs, a.RoleID)
		}
	}
	var roles []model.Role
	if len(roleIDs) > 0 {
		database.DB.Where("id IN ?", roleIDs).Find(&roles)
	}
	roleByID := make(map[uint]string, len(roles))
	for _, r := range roles {
		roleByID[r.ID] = r.Name
	}

	return func(uid uint64) (map[string]interface{}, bool) {
		a, ok := adminByID[uint(uid)]
		if !ok {
			return nil, false
		}
		roleName := ""
		if a.RoleID > 0 {
			roleName = roleByID[a.RoleID]
		}
		return map[string]interface{}{
			"id":       a.ID,
			"name":     a.Name,
			"desc":     a.Desc,
			"pic":      GetFullURL(a.Pic),
			"type":     a.Type,
			"roleName": roleName,
			"loginCnt": a.LoginCnt,
		}, true
	}
}

func ForceOfflineAdmin(idStr, token string) error {
	_, prefix := tokenutil.GetTokenConfig("admin")
	if rd.RDB == nil || token == "" {
		return nil
	}
	rd.RDB.Del(rd.Ctx, prefix+"a:"+token)
	rd.RDB.SRem(rd.Ctx, prefix+"s:"+idStr, token)
	if count, _ := rd.RDB.SCard(rd.Ctx, prefix+"s:"+idStr).Result(); count == 0 {
		rd.RDB.Del(rd.Ctx, prefix+"s:"+idStr)
	}
	return nil
}

// BatchForceOfflineAdmin 批量踢管理员（pipeline 一次完成）。
func BatchForceOfflineAdmin(items []struct {
	IDStr string `json:"idStr"`
	Token string `json:"token"`
}) (int, error) {
	_, prefix := tokenutil.GetTokenConfig("admin")
	if rd.RDB == nil || len(items) == 0 {
		return 0, nil
	}
	byID := make(map[string][]string, len(items))
	for _, it := range items {
		if it.Token == "" {
			continue
		}
		byID[it.IDStr] = append(byID[it.IDStr], it.Token)
	}

	pipe := rd.RDB.Pipeline()
	for idStr, tokens := range byID {
		authKeys := make([]string, len(tokens))
		for i, t := range tokens {
			authKeys[i] = prefix + "a:" + t
		}
		pipe.Del(rd.Ctx, authKeys...)
		pipe.SRem(rd.Ctx, prefix+"s:"+idStr, anyToIface(tokens)...)
	}
	if _, err := pipe.Exec(rd.Ctx); err != nil && err != redis.Nil {
		return 0, err
	}
	for idStr := range byID {
		setKey := prefix + "s:" + idStr
		if n, _ := rd.RDB.SCard(rd.Ctx, setKey).Result(); n == 0 {
			rd.RDB.Del(rd.Ctx, setKey)
		}
	}
	return len(items), nil
}

func AdminLogout(adminID uint, currentToken string) error {
	_, prefix := tokenutil.GetTokenConfig("admin")
	if rd.RDB == nil {
		return nil
	}
	idStr := strconv.Itoa(int(adminID))
	if currentToken == "" {
		// 兜底：无 token 时清空该 adminID 的所有 session
		tokens, _ := rd.RDB.SMembers(rd.Ctx, prefix+"s:"+idStr).Result()
		for _, t := range tokens {
			rd.RDB.Del(rd.Ctx, prefix+"a:"+t)
		}
		rd.RDB.Del(rd.Ctx, prefix+"s:"+idStr)
		return nil
	}
	rd.RDB.Del(rd.Ctx, prefix+"a:"+currentToken)
	rd.RDB.SRem(rd.Ctx, prefix+"s:"+idStr, currentToken)
	if count, _ := rd.RDB.SCard(rd.Ctx, prefix+"s:"+idStr).Result(); count == 0 {
		rd.RDB.Del(rd.Ctx, prefix+"s:"+idStr)
	}
	return nil
}

func SetSetup(key, value, typ, addIP string) error {
	var setup model.Setup
	result := database.DB.Where("`setup_key` = ?", key).First(&setup)
	if result.Error != nil {
		setup = model.Setup{
			Key:     key,
			Value:   value,
			Type:    typ,
			AddTime: database.Now(),
		}
		return database.DB.Create(&setup).Error
	}
	return database.DB.Model(&setup).Updates(map[string]interface{}{
		"setup_value":  value,
		"setup_type":   typ,
		"setup_edit_time": database.Now(),
	}).Error
}

func SetContentSetup(key, value, addIP string) error {
	var setup model.Setup
	result := database.DB.Where("`setup_key` = ?", key).First(&setup)
	if result.Error != nil {
		setup = model.Setup{
			Key:     key,
			Value:   value,
			AddTime: database.Now(),
		}
		return database.DB.Create(&setup).Error
	}
	return database.DB.Model(&setup).Update("setup_value", value).Error
}

func GetUserByOpenID(openID string) (*model.User, error) {
	var user model.User
	err := database.DB.Where("`user_mini_openid` = ?", openID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByID(id string) (map[string]interface{}, error) {
	var user model.User
	err := database.DB.Where("`id` = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	uid, _ := strconv.Atoi(id)
	deptIDs := getUserDeptIDs(uint(uid))
	var deptNames []string
	var topDeptNames []string
	for _, did := range deptIDs {
		var d model.Department
		if database.DB.First(&d, did).Error == nil {
			deptNames = append(deptNames, d.Name)
		}
		topDeptNames = append(topDeptNames, getTopDeptName(did))
	}
	m := map[string]interface{}{
		"id": user.ID, "name": user.Name, "mobile": user.Mobile,
		"avatar": GetFullURL(user.Pic), "pic": GetFullURL(user.Pic), "status": user.Status,
		"forms": user.Forms, "loginCnt": user.LoginCnt,
		"addTime": user.AddTime, "loginTime": user.LoginTime,
	}
	m["deptIds"] = deptIDs
	m["deptNames"] = deptNames
	m["topDeptNames"] = topDeptNames
	return m, nil
}

func getUserDeptIDs(userID uint) []uint {
	var depts []model.UserDept
	database.DB.Where("`user_dept_user_id` = ?", userID).Find(&depts)
	ids := make([]uint, len(depts))
	for i, d := range depts {
		ids[i] = d.DeptID
	}
	return ids
}

func saveUserDepts(userID uint, deptIDs []uint) {
	database.DB.Where("`user_dept_user_id` = ?", userID).Delete(&model.UserDept{})
	for _, deptID := range deptIDs {
		if deptID > 0 {
			database.DB.Create(&model.UserDept{UserID: userID, DeptID: deptID})
		}
	}
}

func GetUserList(keyword, sortStr string, page, pageSize int, adminID uint) ([]map[string]interface{}, int64, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)
	var list []model.User
	var total int64
	query := database.DB.Model(&model.User{})
	if keyword != "" {
		query = query.Where("`user_name` LIKE ? OR `user_mobile` LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	// Data scope
	deptIDs := getDeptVisibleIDs(&admin)
	if deptIDs != nil {
		query = query.Where("`id` IN (SELECT `user_dept_user_id` FROM `user_depts` WHERE `user_dept_dept_id` IN ?)", deptIDs)
	}
	query.Count(&total)
	orderClause := parseSort(sortStr, map[string]string{
		"name":     "`user_name`",
		"mobile":   "`user_mobile`",
		"status":   "`user_status`",
		"loginCnt": "`user_login_cnt`",
		"addTime":  "`user_add_time`",
	})
	if orderClause == "" {
		orderClause = "`user_add_time` DESC"
	}
	err := query.Order(orderClause).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	result := make([]map[string]interface{}, len(list))
	for i, u := range list {
		m := map[string]interface{}{
			"id": u.ID, "name": u.Name, "mobile": u.Mobile,
			"avatar": GetFullURL(u.Pic), "pic": GetFullURL(u.Pic), "status": u.Status,
			"loginCnt": u.LoginCnt, "addTime": u.AddTime, "loginTime": u.LoginTime,
		}
		m["deptIds"] = getUserDeptIDs(u.ID)
		result[i] = m
	}
	return result, total, nil
}

func AddUser(name, mobile, pic, forms, addIP string, deptIDs []uint) error {
	now := database.Now()
	hash := md5.Sum([]byte(fmt.Sprintf("%s-%d", name, now)))
	miniOpenID := hex.EncodeToString(hash[:])
	user := model.User{
		MiniOpenID: miniOpenID,
		Name:       name,
		Mobile:     mobile,
		Pic:        pic,
		Forms:      forms,
		Status:     1,
		AddTime:    now,
		AddIP:      addIP,
		EditTime:   now,
		EditIP:     addIP,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		return err
	}
	saveUserDepts(user.ID, deptIDs)
	return nil
}

func EditUser(id, name, mobile, pic, forms, addIP string, deptIDs []uint) error {
	updates := map[string]interface{}{
		"user_name":      name,
		"user_mobile":    mobile,
		"user_edit_time": database.Now(),
		"user_edit_ip":   addIP,
	}
	if pic != "" {
		updates["user_pic"] = pic
	}
	if forms != "" {
		updates["user_forms"] = forms
	}
	if err := database.DB.Model(&model.User{}).Where("`id` = ?", id).Updates(updates).Error; err != nil {
		return err
	}
	uid, _ := strconv.Atoi(id)
	saveUserDepts(uint(uid), deptIDs)
	return nil
}

func DelUser(id string) error {
	return database.DB.Where("`id` = ?", id).Delete(&model.User{}).Error
}

func DelUsers(ids []string) error {
	return database.DB.Where("`id` IN ?", ids).Delete(&model.User{}).Error
}

func StatusUser(id string, status int, reason string) error {
	updates := map[string]interface{}{
		"user_status": status,
	}
	if status == 1 {
		updates["user_check_reason"] = ""
	} else if reason != "" {
		updates["user_check_reason"] = reason
	}
	return database.DB.Model(&model.User{}).Where("`id` = ?", id).Updates(updates).Error
}

func ResetUserPassword(id string) error {
	h := md5.Sum([]byte("123456"))
	passwordMD5 := hex.EncodeToString(h[:])
	return database.DB.Model(&model.User{}).Where("`id` = ?", id).Update("user_password", passwordMD5).Error
}

func GetUserFormFields() ([]model.UserFormField, error) {
	var setup model.Setup
	err := database.DB.Where("`setup_key` = ?", "SETUP_USER_FORM_FIELDS").First(&setup).Error
	if err != nil {
		return []model.UserFormField{}, nil
	}
	var list []model.UserFormField
	if setup.Value != "" {
		json.Unmarshal([]byte(setup.Value), &list)
	}
	return list, nil
}

func SaveUserFormFields(fields []model.UserFormField) error {
	jsonData, _ := json.Marshal(fields)
	return SetContentSetup("SETUP_USER_FORM_FIELDS", string(jsonData), "")
}

func GetAdminEnrollList(keyword, sortStr string, page, pageSize int, adminID uint) ([]model.Enroll, int64, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)
	var list []model.Enroll
	var total int64
	query := database.DB.Model(&model.Enroll{})
	if keyword != "" {
		query = query.Where("`enroll_title` LIKE ?", "%"+keyword+"%")
	}
	// Data scope
	where, args := BuildDataScopeFilter(&admin, "`enroll_dept_id`", "`enroll_create_by`")
	if where != "" {
		query = query.Where(where, args...)
	}
	query.Count(&total)
	orderClause := parseSort(sortStr, map[string]string{
		"title":   "`enroll_title`",
		"sort":    "`enroll_order`",
		"status":  "`enroll_status`",
		"isVouch": "`enroll_vouch`",
		"userCnt": "`enroll_user_cnt`",
		"joinCnt": "`enroll_join_cnt`",
		"addTime": "`enroll_add_time`",
	})
	if orderClause == "" {
		orderClause = "`enroll_add_time` DESC"
	}
	err := query.Order(orderClause).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	list = populateEnrollFields(list)
	// Recalculate user count and join count from actual records
	for i := range list {
		eid := strconv.Itoa(int(list[i].ID))
		var userCnt int64
		database.DB.Raw(
			"SELECT COUNT(DISTINCT uid) FROM (SELECT `enroll_join_user_id` AS uid FROM `enroll_joins` WHERE `enroll_join_enroll_id` = ? UNION SELECT `enroll_user_mini_openid` AS uid FROM `enroll_users` WHERE `enroll_user_enroll_id` = ?) AS u",
			eid, eid,
		).Scan(&userCnt)
		list[i].UserCnt = int(userCnt)
		var joinCnt int64
		database.DB.Model(&model.EnrollJoin{}).Where("`enroll_join_enroll_id` = ?", eid).Count(&joinCnt)
		list[i].JoinCnt = int(joinCnt)
	}
	return list, total, nil
}

func GetEnrollDetail(id string) (*model.Enroll, error) {
	var enroll model.Enroll
	err := database.DB.Where("`id` = ?", id).First(&enroll).Error
	if err != nil {
		return nil, err
	}
	var obj enrollObj
	if enroll.Obj != "" {
		json.Unmarshal([]byte(enroll.Obj), &obj)
	}
	if len(obj.Cover) > 0 {
		enroll.Img = GetFullURL(obj.Cover[0])
	}
	enroll.Desc = obj.Desc
	return &enroll, nil
}

func UpdateEnrollForms(id, forms string) error {
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Update("enroll_forms", forms).Error
}

func GetAdminNewsList(keyword, sortStr string, page, pageSize int, adminID uint) ([]model.News, int64, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)
	var list []model.News
	var total int64
	query := database.DB.Model(&model.News{})
	if keyword != "" {
		query = query.Where("`news_title` LIKE ?", "%"+keyword+"%")
	}
	// Data scope
	where, args := BuildDataScopeFilter(&admin, "`news_dept_id`", "`news_create_by`")
	if where != "" {
		query = query.Where(where, args...)
	}
	query.Count(&total)
	orderClause := parseSort(sortStr, map[string]string{
		"title":   "`news_title`",
		"type":    "`news_cate_id`",
		"order":   "`news_order`",
		"status":  "`news_status`",
		"vouch":   "`news_vouch`",
		"addTime": "`news_add_time`",
	})
	if orderClause == "" {
		orderClause = "`news_add_time` DESC"
	}
	err := query.Order(orderClause).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func GetNewsDetail(id string) (*model.News, error) {
	var news model.News
	err := database.DB.Where("`id` = ?", id).First(&news).Error
	if err != nil {
		return nil, err
	}
	news = populateNewsFields([]model.News{news})[0]
	return &news, nil
}

func DelNews(id string) error {
	return database.DB.Where("`id` = ?", id).Delete(&model.News{}).Error
}

func DelNewsList(ids []string) error {
	return database.DB.Where("`id` IN ?", ids).Delete(&model.News{}).Error
}

func GetEnrollUserList(enrollID, keyword string) ([]model.EnrollUser, error) {
	var list []model.EnrollUser
	query := database.DB.Where("`enroll_user_enroll_id` = ?", enrollID)
	if keyword != "" {
		query = query.Where("`enroll_user_mini_openid` IN (SELECT `user_mini_openid` FROM `users` WHERE `user_name` LIKE ?)", "%"+keyword+"%")
	}
	err := query.Order("`enroll_user_add_time` DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	for i := range list {
		var u model.User
		database.DB.Where("`user_mini_openid` = ?", list[i].MiniOpenID).First(&u)
		list[i].EnrollTitle = u.Name
		list[i].UserName = u.Name
		if u.ID > 0 {
			var ud model.UserDept
			database.DB.Where("`user_dept_user_id` = ?", u.ID).First(&ud)
			if ud.DeptID > 0 {
				var dept model.Department
				database.DB.First(&dept, ud.DeptID)
				list[i].DeptName = dept.Name
				list[i].TopDeptName = getTopDeptName(ud.DeptID)
			}
		}
	}
	return list, nil
}

func GetEnrollJoinList(enrollID, keyword string, page, pageSize int) ([]model.EnrollJoin, int64, error) {
	var list []model.EnrollJoin
	var total int64
	query := database.DB.Model(&model.EnrollJoin{})
	if enrollID != "" {
		query = query.Where("`enroll_join_enroll_id` = ?", enrollID)
	}
	if keyword != "" {
		query = query.Where("`enroll_join_user_id` IN (SELECT `user_mini_openid` FROM `users` WHERE `user_name` LIKE ?) OR `enroll_join_user_id` LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Order("`enroll_join_add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	// Populate user names and dept
	userMap := map[string]model.User{}
	var users []model.User
	database.DB.Find(&users)
	for _, u := range users {
		userMap[u.MiniOpenID] = u
	}
	for i := range list {
		u, ok := userMap[list[i].UserID]
		if ok {
			list[i].EnrollTitle = u.Name
			list[i].UserName = u.Name
			var ud model.UserDept
			database.DB.Where("`user_dept_user_id` = ?", u.ID).First(&ud)
			if ud.DeptID > 0 {
				var dept model.Department
				database.DB.First(&dept, ud.DeptID)
				list[i].DeptName = dept.Name
				list[i].TopDeptName = getTopDeptName(ud.DeptID)
			}
		}
	}
	return list, total, nil
}

type EnrollStatItem struct {
	UserID      string `json:"userId"`
	UserName    string `json:"userName"`
	DeptName    string `json:"deptName"`
	TopDeptName string `json:"topDeptName"`
	JoinCnt     int    `json:"joinCnt"`
	DayCnt      int    `json:"dayCnt"`
}

func GetEnrollStats(enrollID, startDay, endDay string) ([]EnrollStatItem, error) {
	var joins []model.EnrollJoin
	query := database.DB.Where("`enroll_join_enroll_id` = ?", enrollID)
	if startDay != "" {
		query = query.Where("`enroll_join_day` >= ?", startDay)
	}
	if endDay != "" {
		query = query.Where("`enroll_join_day` <= ?", endDay)
	}
	err := query.Find(&joins).Error
	if err != nil {
		return nil, err
	}
	type tmp struct{ cnt, days int }
	agg := map[string]*tmp{}
	daySet := map[string]map[string]bool{}
	for _, j := range joins {
		if _, ok := agg[j.UserID]; !ok {
			agg[j.UserID] = &tmp{}
			daySet[j.UserID] = map[string]bool{}
		}
		agg[j.UserID].cnt++
		daySet[j.UserID][j.Day] = true
	}
	var result []EnrollStatItem
	for uid, t := range agg {
		t.days = len(daySet[uid])
		item := EnrollStatItem{
			UserID:  uid,
			JoinCnt: t.cnt,
			DayCnt:  t.days,
		}
		var u model.User
		database.DB.Where("`user_mini_openid` = ?", uid).First(&u)
		item.UserName = u.Name
		if u.ID > 0 {
			var ud model.UserDept
			database.DB.Where("`user_dept_user_id` = ?", u.ID).First(&ud)
			if ud.DeptID > 0 {
				var dept model.Department
				database.DB.First(&dept, ud.DeptID)
				item.DeptName = dept.Name
				item.TopDeptName = getTopDeptName(ud.DeptID)
			}
		}
		result = append(result, item)
	}
	return result, nil
}

func GetEnrollJoinDataURL(enrollID string) (string, error) {
	return "", nil
}

func DeleteEnrollJoinDataExcel(enrollID string) error {
	filename := fmt.Sprintf("export_enroll_%s.csv", enrollID)
	os.Remove(filepath.Join("./uploads", filename))
	return nil
}

func ExportEnrollJoinDataExcel(enrollID, startDay, endDay string) (string, error) {
	var joins []model.EnrollJoin
	query := database.DB.Where("`enroll_join_enroll_id` = ?", enrollID)
	if startDay != "" {
		query = query.Where("`enroll_join_day` >= ?", startDay)
	}
	if endDay != "" {
		query = query.Where("`enroll_join_day` <= ?", endDay)
	}
	query.Order("`enroll_join_add_time` DESC").Find(&joins)

	// Get enroll title
	var enroll model.Enroll
	database.DB.Where("`id` = ?", enrollID).First(&enroll)

	// Get user names
	userNames := map[string]string{}
	var users []model.User
	database.DB.Find(&users)
	for _, u := range users {
		userNames[u.MiniOpenID] = u.Name
	}

	filename := fmt.Sprintf("export_enroll_%s.csv", enrollID)
	filepath := filepath.Join("./uploads", filename)

	f, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// BOM for Excel UTF-8 compatibility
	f.WriteString("\xEF\xBB\xBF")

	writer.Write([]string{"打卡项目", enroll.Title})
	writer.Write([]string{"用户ID", "用户姓名", "打卡日期", "打卡内容", "打卡时间", "IP地址"})
	for _, j := range joins {
		joinTime := time.UnixMilli(j.AddTime).Format("2006-01-02 15:04:05")
		writer.Write([]string{
			j.UserID,
			userNames[j.UserID],
			j.Day,
			j.Forms,
			joinTime,
			j.AddIP,
		})
	}

	return filename, nil
}

func GetUserDataURL() (string, error) {
	return "", nil
}

func DeleteUserDataExcel() error {
	return nil
}

func ExportUserDataExcel() (string, error) {
	return "", nil
}

func SortEnroll(id, sortStr string) error {
	sort, err := strconv.Atoi(sortStr)
	if err != nil {
		return err
	}
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Update("enroll_order", sort).Error
}

func VouchEvent(id string, vouch int) error {
	return database.DB.Model(&model.Event{}).Where("`id` = ?", id).Update("event_vouch", vouch).Error
}

func TopEvent(id string, top int) error {
	return database.DB.Model(&model.Event{}).Where("`id` = ?", id).Update("event_is_top", top).Error
}

func VouchEnroll(id string, vouch int) error {
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Update("enroll_vouch", vouch).Error
}

func StatusEnroll(id string, status int) error {
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Update("enroll_status", status).Error
}

func ClearEnrollAll(id string) error {
	database.DB.Where("`enroll_join_enroll_id` = ?", id).Delete(&model.EnrollJoin{})
	database.DB.Where("`enroll_user_enroll_id` = ?", id).Delete(&model.EnrollUser{})
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Updates(map[string]interface{}{
		"enroll_join_cnt": 0,
		"enroll_user_cnt": 0,
	}).Error
}

func DelEnroll(id string) error {
	database.DB.Where("`enroll_join_enroll_id` = ?", id).Delete(&model.EnrollJoin{})
	database.DB.Where("`enroll_user_enroll_id` = ?", id).Delete(&model.EnrollUser{})
	return database.DB.Where("`id` = ?", id).Delete(&model.Enroll{}).Error
}

func DelEnrolls(ids []string) error {
	for _, id := range ids {
		if err := DelEnroll(id); err != nil {
			return err
		}
	}
	return nil
}

func DelEnrollJoin(id string) error {
	var join model.EnrollJoin
	err := database.DB.Where("`id` = ?", id).First(&join).Error
	if err != nil {
		return err
	}
	database.DB.Delete(&join)
	var enroll model.Enroll
	database.DB.Where("`id` = ?", join.EnrollID).First(&enroll)
	if enroll.JoinCnt > 0 {
		database.DB.Model(&enroll).UpdateColumn("enroll_join_cnt", enroll.JoinCnt-1)
	}
	return nil
}

func DelEnrollJoins(ids []string) error {
	for _, id := range ids {
		if err := DelEnrollJoin(id); err != nil {
			return err
		}
	}
	return nil
}

func RemoveEnrollUser(enrollID, userID string) error {
	database.DB.Where("`enroll_user_enroll_id` = ? AND `enroll_user_mini_openid` = ?", enrollID, userID).Delete(&model.EnrollUser{})
	database.DB.Where("`enroll_join_enroll_id` = ? AND `enroll_join_user_id` = ?", enrollID, userID).Delete(&model.EnrollJoin{})
	var cnt int64
	database.DB.Model(&model.EnrollUser{}).Where("`enroll_user_enroll_id` = ?", enrollID).Count(&cnt)
	database.DB.Model(&model.Enroll{}).Where("`id` = ?", enrollID).Update("enroll_user_cnt", cnt)
	return nil
}

func RemoveEnrollUsers(enrollID string, userIDs []string) error {
	for _, uid := range userIDs {
		if err := RemoveEnrollUser(enrollID, uid); err != nil {
			return err
		}
	}
	return nil
}

func EditEnrollUserForms(enrollID, userID, forms string) error {
	return database.DB.Model(&model.EnrollUser{}).
		Where("`enroll_user_enroll_id` = ? AND `enroll_user_mini_openid` = ?", enrollID, userID).
		Update("enroll_user_forms", forms).Error
}

func SortNews(id, sortStr string) error {
	sort, err := strconv.Atoi(sortStr)
	if err != nil {
		return err
	}
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("news_order", sort).Error
}

func StatusNews(id string, status int) error {
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("news_status", status).Error
}

func InsertNews(title, desc, cateID, cateName, content, qr, pic, forms, addIP, publishDeptIds string, status, order int, deptID, createBy uint) error {
	news := model.News{
		Title:           title,
		Desc:            desc,
		Status:          status,
		CateID:          cateID,
		CateName:        cateName,
		Order:           order,
		Content:         content,
		QR:              qr,
		Pic:             pic,
		Forms:           forms,
		DeptID:          deptID,
		PublishDeptIds:  publishDeptIds,
		CreateBy:        createBy,
		AddTime:         database.Now(),
		AddIP:           addIP,
	}
	return database.DB.Create(&news).Error
}

func EditNews(id, title, desc, cateID, cateName, content, qr, addIP, publishDeptIds string, status, order int, deptID uint) error {
	updates := map[string]interface{}{
		"news_title":              title,
		"news_desc":               desc,
		"news_status":             status,
		"news_cate_id":            cateID,
		"news_cate_name":          cateName,
		"news_order":              order,
		"news_content":            content,
		"news_qr":                 qr,
		"news_dept_id":            deptID,
		"news_publish_dept_ids":   publishDeptIds,
		"news_edit_time":          database.Now(),
		"news_edit_ip":            addIP,
	}
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Updates(updates).Error
}

func UpdateNewsForms(id, forms string) error {
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("news_forms", forms).Error
}

func UpdateNewsPic(id, pic string) error {
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("news_pic", pic).Error
}

func UpdateNewsContent(id, content string) error {
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("news_content", content).Error
}

func InsertEnroll(title, cateID, cateName, forms, joinForms, qr, addIP, publishDeptIds string, status, order, dayCnt int, start, end int64, obj string, allowRepeat bool, dailyLimit int, deptID, createBy uint) error {
	enroll := model.Enroll{
		Title:           title,
		Status:          status,
		CateID:          cateID,
		CateName:        cateName,
		Start:           start,
		End:             end,
		DayCnt:          dayCnt,
		Order:           order,
		Forms:           forms,
		JoinForms:       joinForms,
		QR:              qr,
		Obj:             obj,
		AllowRepeat:     allowRepeat,
		DailyLimit:      dailyLimit,
		DeptID:          deptID,
		PublishDeptIds:  publishDeptIds,
		CreateBy:        createBy,
		AddTime:         database.Now(),
		AddIP:           addIP,
	}
	return database.DB.Create(&enroll).Error
}

func EditEnroll(id, title, cateID, cateName, forms, joinForms, qr, addIP, publishDeptIds string, status, order, dayCnt int, start, end int64, obj string, allowRepeat bool, dailyLimit int, deptID uint) error {
	updates := map[string]interface{}{
		"enroll_title":             title,
		"enroll_status":            status,
		"enroll_cate_id":           cateID,
		"enroll_cate_name":         cateName,
		"enroll_start":             start,
		"enroll_end":               end,
		"enroll_day_cnt":           dayCnt,
		"enroll_order":             order,
		"enroll_forms":             forms,
		"enroll_join_forms":        joinForms,
		"enroll_repeat":            allowRepeat,
		"enroll_limit":             dailyLimit,
		"enroll_dept_id":           deptID,
		"enroll_publish_dept_ids":  publishDeptIds,
		"enroll_qr":                qr,
		"enroll_edit_time":         database.Now(),
		"enroll_edit_ip":           addIP,
	}
	if obj != "" {
		updates["enroll_obj"] = obj
	}
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Updates(updates).Error
}

// ===================== SysDict =====================

func GetDictTypes() ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	rows, err := database.DB.Model(&model.SysDict{}).
		Select("`dict_type_code` as type_code, `dict_type_name` as type_name, COUNT(*) as item_cnt").
		Group("`dict_type_code`, `dict_type_name`").
		Order("MIN(`dict_sort`) ASC, MIN(`dict_add_time`) ASC").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var typeCode, typeName string
		var itemCnt int64
		if err := rows.Scan(&typeCode, &typeName, &itemCnt); err != nil {
			continue
		}
		results = append(results, map[string]interface{}{
			"typeCode": typeCode,
			"typeName": typeName,
			"itemCnt":  itemCnt,
		})
	}
	return results, nil
}

func GetDictByType(typeCode string) ([]model.SysDict, error) {
	var list []model.SysDict
	return list, database.DB.Where("`dict_type_code` = ?", typeCode).Order("`dict_sort` ASC, `id` ASC").Find(&list).Error
}

func AddDictItem(typeCode, typeName, label, value, remark string, sort int) error {
	now := database.Now()
	d := model.SysDict{
		TypeCode: typeCode,
		TypeName: typeName,
		Label:    label,
		Value:    value,
		Sort:     sort,
		Status:   1,
		Remark:   remark,
		AddTime:  now,
		EditTime: now,
	}
	return database.DB.Create(&d).Error
}

func EditDictItem(id, label, value, remark string, sort int) error {
	return database.DB.Model(&model.SysDict{}).Where("`id` = ?", id).Updates(map[string]interface{}{
		"dict_label":   label,
		"dict_value":   value,
		"dict_sort":    sort,
		"dict_remark":  remark,
		"dict_edit_time": database.Now(),
	}).Error
}

func DelDictItem(id string) error {
	return database.DB.Where("`id` = ?", id).Delete(&model.SysDict{}).Error
}

func DelDictByType(typeCode string) error {
	return database.DB.Where("`dict_type_code` = ?", typeCode).Delete(&model.SysDict{}).Error
}

func EditDictTypeName(oldTypeCode, typeCode, typeName string) error {
	updates := map[string]interface{}{"dict_type_name": typeName}
	if oldTypeCode != typeCode {
		updates["dict_type_code"] = typeCode
	}
	return database.DB.Model(&model.SysDict{}).Where("`dict_type_code` = ?", oldTypeCode).Updates(updates).Error
}

// Department

func getDeptDescendantIDs(all []*model.Department, parentIDs []uint) []uint {
	ids := make([]uint, 0)
	idSet := make(map[uint]bool)
	for _, id := range parentIDs {
		idSet[id] = true
	}
	queue := make([]uint, len(parentIDs))
	copy(queue, parentIDs)
	for len(queue) > 0 {
		pid := queue[0]
		queue = queue[1:]
		for _, d := range all {
			if d.ParentID == pid && !idSet[d.ID] {
				idSet[d.ID] = true
				queue = append(queue, d.ID)
			}
		}
	}
	for id := range idSet {
		ids = append(ids, id)
	}
	return ids
}

func getDeptVisibleIDs(admin *model.Admin) []uint {
	if admin.Type == 1 {
		return nil
	}
	var role model.Role
	if err := database.DB.First(&role, admin.RoleID).Error; err != nil {
		return nil
	}
	var all []*model.Department
	database.DB.Find(&all)
	switch role.DataScope {
	case 1:
		return nil
	case 2:
		deptIDs := getAdminDeptIDs(admin.ID)
		if len(deptIDs) == 0 {
			return nil
		}
		return getDeptDescendantIDs(all, deptIDs)
	case 3:
		return getAdminDeptIDs(admin.ID)
	case 4:
		deptIDs := GetRoleDeptIDs(admin.RoleID)
		if len(deptIDs) == 0 {
			return nil
		}
		return getDeptDescendantIDs(all, deptIDs)
	}
	return nil
}

func GetDeptTree(adminID uint) ([]*model.Department, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)
	var list []*model.Department
	if err := database.DB.Order("`dept_sort` ASC, `id` ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	visibleIDs := getDeptVisibleIDs(&admin)
	if visibleIDs != nil {
		visibleSet := make(map[uint]bool)
		for _, id := range visibleIDs {
			visibleSet[id] = true
		}
		var filtered []*model.Department
		for _, d := range list {
			if visibleSet[d.ID] {
				filtered = append(filtered, d)
			}
		}
		// Detach from non-visible parents - set ParentID=0 so they appear as root
		for _, d := range filtered {
			if d.ParentID != 0 && !visibleSet[d.ParentID] {
				d.ParentID = 0
			}
		}
		list = filtered
	}
	return buildDeptTree(list, 0), nil
}

func buildDeptTree(list []*model.Department, pid uint) []*model.Department {
	var tree []*model.Department
	for _, item := range list {
		if item.ParentID == pid {
			item.Children = buildDeptTree(list, item.ID)
			tree = append(tree, item)
		}
	}
	return tree
}

func AddDept(name string, parentID uint, sort int) error {
	dept := model.Department{
		Name:     name,
		ParentID: parentID,
		Sort:     sort,
		Status:   1,
		AddTime:  database.Now(),
	}
	return database.DB.Create(&dept).Error
}

func EditDept(id uint, name string, parentID uint, sort, status int) error {
	updates := map[string]interface{}{
		"dept_name":      name,
		"dept_parent_id": parentID,
		"dept_sort":      sort,
		"dept_status":    status,
		"dept_edit_time": database.Now(),
	}
	return database.DB.Model(&model.Department{}).Where("`id` = ?", id).Updates(updates).Error
}

// ===================== Role =====================

func GetRoleList(adminID uint, keyword string, page, pageSize int) (map[string]interface{}, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)

	var conditions []func(*gorm.DB) *gorm.DB

	// Data scope filter
	if admin.Type != 1 && admin.RoleID > 0 {
		var role model.Role
		if err := database.DB.First(&role, admin.RoleID).Error; err == nil {
			if role.DataScope == 2 || role.DataScope == 4 {
				var deptIDs []uint
				if role.DataScope == 2 {
					deptIDs = getAdminDeptIDs(admin.ID)
				} else {
					deptIDs = GetRoleDeptIDs(admin.RoleID)
				}
				if len(deptIDs) > 0 {
					ids := deptIDs
					rid := admin.RoleID
					conditions = append(conditions, func(d *gorm.DB) *gorm.DB {
						return d.Where("`id` IN (SELECT `role_dept_role_id` FROM `role_depts` WHERE `role_dept_dept_id` IN ?) OR `id` = ?", ids, rid)
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
	database.DB.Model(&model.Role{}).Scopes(conditions...).Count(&total)
	var list []model.Role
	database.DB.Model(&model.Role{}).Scopes(conditions...).Order("`role_sort` ASC, `id` ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	result := make([]map[string]interface{}, len(list))
	for i, r := range list {
		result[i] = map[string]interface{}{
			"id": r.ID, "name": r.Name, "remark": r.Remark,
			"sort": r.Sort, "status": r.Status, "dataScope": r.DataScope,
			"addTime": r.AddTime, "editTime": r.EditTime,
			"menuIds": GetRoleMenuIDs(r.ID),
			"deptIds": GetRoleDeptIDs(r.ID),
		}
	}
	return map[string]interface{}{"list": result, "total": total}, nil
}

func AddRole(name, remark, addIP string, sort, dataScope int) (uint, error) {
	role := model.Role{
		Name:      name,
		Remark:    remark,
		Sort:      sort,
		Status:    1,
		DataScope: dataScope,
		AddTime:   database.Now(),
		AddIP:     addIP,
	}
	err := database.DB.Create(&role).Error
	if err != nil {
		return 0, err
	}
	return role.ID, nil
}

func EditRole(id uint, name, remark, addIP string, sort, status, dataScope int) error {
	updates := map[string]interface{}{
		"role_name":       name,
		"role_remark":     remark,
		"role_sort":       sort,
		"role_status":     status,
		"role_data_scope": dataScope,
		"role_edit_time":  database.Now(),
		"role_edit_ip":    addIP,
	}
	return database.DB.Model(&model.Role{}).Where("`id` = ?", id).Updates(updates).Error
}

func DelRole(id uint) error {
	database.DB.Where("`role_menu_role_id` = ?", id).Delete(&model.RoleMenu{})
	database.DB.Where("`role_dept_role_id` = ?", id).Delete(&model.RoleDept{})
	return database.DB.Where("`id` = ?", id).Delete(&model.Role{}).Error
}

func DelRoles(ids []uint) error {
	for _, id := range ids {
		if err := DelRole(id); err != nil {
			return err
		}
	}
	return nil
}

// ===================== Menu =====================

func GetMenuTree() ([]*model.Menu, error) {
	var list []*model.Menu
	if err := database.DB.Order("`menu_sort` ASC, `id` ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return buildMenuTree(list, 0), nil
}

func GetMenuList() ([]model.Menu, error) {
	var list []model.Menu
	return list, database.DB.Order("`menu_sort` ASC, `id` ASC").Find(&list).Error
}

func buildMenuTree(list []*model.Menu, pid uint) []*model.Menu {
	var tree []*model.Menu
	for _, item := range list {
		if item.ParentID == pid {
			item.Children = buildMenuTree(list, item.ID)
			tree = append(tree, item)
		}
	}
	return tree
}

func AddMenu(name string, parentID uint, path, perms, icon string, sort, mtype int) error {
	return database.DB.Table("menus").Create(map[string]interface{}{
		"menu_name":      name,
		"menu_parent_id": parentID,
		"menu_path":      path,
		"menu_perms":     perms,
		"menu_icon":      icon,
		"menu_sort":      sort,
		"menu_status":    1,
		"menu_type":      mtype,
		"menu_add_time":  database.Now(),
	}).Error
}

func EditMenu(id uint, name string, parentID uint, path, perms, icon string, sort, status, mtype int) error {
	updates := map[string]interface{}{
		"menu_name":      name,
		"menu_parent_id": parentID,
		"menu_path":      path,
		"menu_perms":     perms,
		"menu_icon":      icon,
		"menu_sort":      sort,
		"menu_status":    status,
		"menu_type":      mtype,
		"menu_edit_time": database.Now(),
	}
	return database.DB.Model(&model.Menu{}).Where("`id` = ?", id).Updates(updates).Error
}

func DelMenu(id uint) error {
	database.DB.Where("`role_menu_menu_id` = ?", id).Delete(&model.RoleMenu{})
	database.DB.Where("`menu_parent_id` = ?", id).Delete(&model.Menu{})
	return database.DB.Where("`id` = ?", id).Delete(&model.Menu{}).Error
}

// ===================== RoleMenu =====================

func GetRoleMenuIDs(roleID uint) []uint {
	var list []model.RoleMenu
	database.DB.Where("`role_menu_role_id` = ?", roleID).Find(&list)
	ids := make([]uint, len(list))
	for i, rm := range list {
		ids[i] = rm.MenuID
	}
	return ids
}

func SetRoleMenus(roleID uint, menuIDs []uint) {
	database.DB.Where("`role_menu_role_id` = ?", roleID).Delete(&model.RoleMenu{})
	for _, menuID := range menuIDs {
		if menuID > 0 {
			database.DB.Create(&model.RoleMenu{RoleID: roleID, MenuID: menuID})
		}
	}
}

// GetAdminMenuTree returns the menu tree for an admin (filtered by role)
func GetAdminMenuTree(admin *model.Admin) ([]*model.Menu, error) {
	if admin.Type == 1 {
		return GetMenuTree()
	}
	if admin.RoleID == 0 {
		return []*model.Menu{}, nil
	}
	menuIDs := GetRoleMenuIDs(admin.RoleID)
	if len(menuIDs) == 0 {
		return []*model.Menu{}, nil
	}
	var list []*model.Menu
	database.DB.Where("`id` IN ?", menuIDs).Order("`menu_sort` ASC, `id` ASC").Find(&list)
	return buildMenuTree(list, 0), nil
}

// GetAdminPerms returns the permission keys for an admin
func GetAdminPerms(admin *model.Admin) []string {
	if admin.Type == 1 {
		var all []model.Menu
		database.DB.Where("`menu_perms` != ''").Find(&all)
		var perms []string
		for _, m := range all {
			for _, p := range strings.Split(m.Perms, ",") {
				p = strings.TrimSpace(p)
				if p != "" {
					perms = append(perms, p)
				}
			}
		}
		return perms
	}
	if admin.RoleID == 0 {
		return nil
	}
	menuIDs := GetRoleMenuIDs(admin.RoleID)
	if len(menuIDs) == 0 {
		return nil
	}
	var menus []model.Menu
	database.DB.Where("`id` IN ? AND `menu_perms` != ''", menuIDs).Find(&menus)
	var perms []string
	for _, m := range menus {
		for _, p := range strings.Split(m.Perms, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				perms = append(perms, p)
			}
		}
	}
	return perms
}

// ===================== RoleDept =====================

func GetRoleDeptIDs(roleID uint) []uint {
	var list []model.RoleDept
	database.DB.Where("`role_dept_role_id` = ?", roleID).Find(&list)
	ids := make([]uint, len(list))
	for i, d := range list {
		ids[i] = d.DeptID
	}
	return ids
}

func SetRoleDepts(roleID uint, deptIDs []uint) {
	database.DB.Where("`role_dept_role_id` = ?", roleID).Delete(&model.RoleDept{})
	for _, deptID := range deptIDs {
		if deptID > 0 {
			database.DB.Create(&model.RoleDept{RoleID: roleID, DeptID: deptID})
		}
	}
}

// ===================== DataScope =====================

// BuildDataScopeFilter returns a WHERE condition string and args for data scope filtering.
// Supports per-table field names for dept_id and create_by.
// For User/Mgr tables that use association tables, use a separate approach.
func BuildDataScopeFilter(admin *model.Admin, deptField, createByField string) (string, []interface{}) {
	if admin.Type == 1 {
		return "", nil
	}
	var role model.Role
	if err := database.DB.First(&role, admin.RoleID).Error; err != nil {
		return "", nil
	}
	switch role.DataScope {
	case 1: // 全部
		return "", nil
	case 2: // 本部门
		deptIDs := getAdminDeptIDs(admin.ID)
		if len(deptIDs) == 0 {
			return deptField + " = 0", nil
		}
		args := make([]interface{}, len(deptIDs))
		for i, d := range deptIDs {
			args[i] = d
		}
		return "(" + deptField + " IN ? OR " + deptField + " = 0)", []interface{}{args}
	case 3: // 本人
		return createByField + " = ?", []interface{}{admin.ID}
	case 4: // 自定义
		deptIDs := GetRoleDeptIDs(admin.RoleID)
		if len(deptIDs) == 0 {
			return deptField + " = 0", nil
		}
		args := make([]interface{}, len(deptIDs))
		for i, d := range deptIDs {
			args[i] = d
		}
		return "(" + deptField + " IN ? OR " + deptField + " = 0)", []interface{}{args}
	}
	return "", nil
}

func DelDept(id uint) error {
	tx := database.DB.Begin()
	tx.Where("`dept_parent_id` = ?", id).Delete(&model.Department{})
	tx.Where("`id` = ?", id).Delete(&model.Department{})
	return tx.Commit().Error
}

func parseSort(sortStr string, allowedFields map[string]string) string {
	if sortStr == "" {
		return ""
	}
	parts := strings.Split(sortStr, ",")
	var orders []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, ":", 2)
		field := strings.TrimSpace(kv[0])
		order := "ASC"
		if len(kv) > 1 && strings.ToUpper(strings.TrimSpace(kv[1])) == "DESC" {
			order = "DESC"
		}
		dbField, ok := allowedFields[field]
		if !ok {
			continue
		}
		orders = append(orders, "`"+dbField+"` "+order)
	}
	if len(orders) == 0 {
		return ""
	}
	return strings.Join(orders, ", ")
}
