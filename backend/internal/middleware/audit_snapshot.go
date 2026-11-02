package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/pkg/database"
)

const (
	snapshotMaxTargets    = 20
	snapshotMaxRelations  = 1024
	snapshotMaxBodyBytes  = 64 << 10
	snapshotMaxValueBytes = 16 << 10
	snapshotMaxDiffBytes  = 64 << 10
	snapshotMaxChanges    = 128
	snapshotReadTimeout   = 750 * time.Millisecond
)

var errSnapshotLimit = errors.New("snapshot limit exceeded")

// AuditSnapshot must be registered AFTER AdminAuth and AdminPerm, inside the
// request audit collector. Reaching this middleware is the route-authorization
// boundary; row-level authorization may still fail in the handler. Nothing read
// here is attached to the recorder until an explicit successful business result.
// Reads are best-effort observations outside the mutation transaction, so they
// cannot prove attribution when another request writes the same row concurrently.
func AuditSnapshot() app.HandlerFunc { return auditSnapshotWithReader(readAuditSnapshot) }

type snapshotSpec struct {
	family, table, key, name string
	columns                  []string
	mode                     string
}

type snapshotPlan struct {
	spec        snapshotSpec
	ids         []string
	afterIDs    []string
	subjectType string
	identity    string
	relatedID   string
	corpConfigs bool
}

// Private values are bounded, request-local comparison material, never audit
// values. In particular, no secret/content digest is computed or persisted.
type snapshotData struct {
	values  map[string]interface{}
	private map[string]string
	name    string
	count   int
}

type snapshotReader func(context.Context, snapshotPlan) (snapshotData, error)

func auditSnapshotWithReader(read snapshotReader) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		spec, supported := snapshotRoute(string(c.Method()), c.FullPath())
		actor, _ := c.Get("admin")
		admin, authenticated := actor.(*model.Admin)
		if !supported || !authenticated || admin == nil || admin.ID == 0 || c.IsAborted() {
			c.Next(ctx)
			return
		}
		plan, state := snapshotRequestPlan(c, spec)
		var before snapshotData
		if state == "" {
			var err error
			before, err = read(ctx, plan)
			state = snapshotReadState("before", err)
		}
		c.Next(ctx)
		if !snapshotBusinessSuccess(c) {
			return
		}
		if state != "" {
			snapshotUnavailable(ctx, c, state)
			return
		}
		afterPlan := plan
		if len(plan.afterIDs) != 0 {
			afterPlan.ids = plan.afterIDs
		}
		after, err := read(ctx, afterPlan)
		if state = snapshotReadState("after", err); state != "" {
			snapshotUnavailable(ctx, c, state)
			return
		}
		changes, err := snapshotDiff(before, after)
		if err != nil {
			snapshotUnavailable(ctx, c, "diff_limit")
			return
		}
		c.Set("audit_snapshot_state", "complete")
		auditlog.Update(ctx, func(d *auditlog.Details) {
			if d.TargetType == "" {
				d.TargetType = spec.family
			}
			if d.TargetID == "" {
				d.TargetID = strings.Join(afterPlan.ids, ",")
				if spec.mode == "org" {
					d.TargetID = plan.subjectType + ":" + d.TargetID + ":" + plan.identity
				}
			}
			if d.TargetName == "" {
				d.TargetName = after.name
				if d.TargetName == "" {
					d.TargetName = before.name
				}
			}
			if plan.relatedID != "" && d.RelatedID == "" {
				d.RelatedType, d.RelatedID = "workflow_definition_version", plan.relatedID
			}
			// Preserve more precise changes supplied by the business layer.
			seen := make(map[string]bool, len(d.Changes))
			for _, change := range d.Changes {
				seen[change.Field] = true
			}
			for _, change := range changes {
				if !seen[change.Field] {
					d.Changes = append(d.Changes, change)
				}
			}
			d.SnapshotStatus, d.SnapshotReason = auditlog.SnapshotRecorded, ""
			if len(d.Changes) == 0 {
				d.SnapshotStatus, d.SnapshotReason = auditlog.SnapshotUnchanged, "no_collected_changes"
			}
		})
	}
}

func snapshotBusinessSuccess(c *app.RequestContext) bool {
	if c.Response.StatusCode() < 200 || c.Response.StatusCode() >= 300 || c.Response.IsBodyStream() {
		return false
	}
	body := c.Response.Body()
	if len(body) == 0 || len(body) > snapshotMaxBodyBytes {
		return false
	}
	var result struct {
		Code *int `json:"code"`
	}
	return json.Unmarshal(body, &result) == nil && result.Code != nil && *result.Code == 0
}

func snapshotReadState(phase string, err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, errSnapshotLimit) {
		return phase + "_limit"
	}
	return phase + "_read_failed"
}

func snapshotUnavailable(ctx context.Context, c *app.RequestContext, state string) {
	// State is a closed set produced locally, never a database error or input.
	c.Set("audit_snapshot_state", state)
	auditlog.Update(ctx, func(d *auditlog.Details) {
		d.SnapshotStatus, d.SnapshotReason = auditlog.SnapshotUnavailable, state
		if state == "unsupported_setup_key" {
			d.SnapshotStatus = auditlog.SnapshotNotConfigured
		}
	})
}

func snapshotResource(family string) snapshotSpec {
	s := snapshotSpec{family: family, key: "id", mode: "row"}
	var cols string
	switch family {
	case "user":
		s.table, s.name = "users", "user_name"
		cols = "id user_account user_name user_status user_position_id user_admin_enabled user_admin_type user_role_id"
	case "role":
		s.table, s.name = "roles", "role_name"
		cols = "id role_name role_sort role_status role_allow_admin_login role_data_scope"
	case "permission":
		s.table, s.key, s.name = "permissions", "permission_key", "permission_name"
		cols = "id permission_key permission_name permission_platform permission_type permission_parent_key permission_sort permission_status"
	case "department":
		s.table, s.name = "departments", "dept_name"
		cols = "id dept_name dept_parent_id dept_sort dept_status"
	case "position":
		s.table, s.name = "positions", "position_name"
		cols = "id position_name position_sort position_status"
	case "dict_type":
		s.table, s.key, s.name = "sys_dict_types", "dict_type_code", "dict_type_name"
		cols = "dict_type_code dict_type_name dict_type_status"
	case "dict_item":
		s.table, s.name = "sys_dicts", "dict_label"
		cols = "id dict_type_code dict_label dict_value dict_calculation_value dict_sort dict_status"
	case "survey":
		s.table, s.key, s.name = "survey", "survey_id", "survey_title"
		cols = "survey_id survey_title survey_category survey_status survey_visibility survey_allow_multi survey_start_time survey_end_time survey_max_response survey_show_result survey_anonymous survey_dept_ids survey_mode survey_order"
	case "exam":
		s.table, s.key, s.name = "exam", "exam_id", "exam_title"
		cols = "exam_id exam_title exam_category exam_status exam_visibility exam_allow_multi exam_anonymous exam_show_result exam_paper_id exam_start_time exam_end_time exam_duration exam_max_attempts exam_show_score exam_pass_score exam_max_response exam_dept_ids exam_mode exam_publish_dept_ids exam_order"
	case "workflow_definition":
		s.table, s.name = "workflow_definitions", "definition_name"
		cols = "id definition_key definition_name definition_display_name definition_category definition_status definition_current_version definition_start_config_revision definition_field_permissions_revision definition_deleted_at"
	case "settings":
		s.table, s.key, s.mode = "setups", "setup_key", "config"
		cols = "setup_key"
	case "org_approver":
		s.table, s.key, s.mode = "workflow_org_approver_assignments", "subject_id", "org"
		cols = "subject_type subject_id identity_code user_id assignment_sort assignment_status"
	}
	s.columns = strings.Fields(cols)
	return s
}

// This catalog only identifies snapshot strategies. It is deliberately separate
// from the parent's audit coverage/permission catalog and matches router templates,
// not client-supplied paths, table names, or column names.
func snapshotRoute(method, fullPath string) (snapshotSpec, bool) {
	if !strings.HasPrefix(fullPath, "/api/v2/admin/") {
		return snapshotSpec{}, false
	}
	route := method + " " + strings.TrimPrefix(fullPath, "/api/v2/admin")
	family, mode := "", ""
	switch route {
	case "PUT /users/:id", "DELETE /users/:id", "PATCH /users/:id/status", "PUT /managers/:id", "DELETE /managers/:id", "PATCH /managers/:id/status":
		family = "user"
	case "DELETE /users", "DELETE /managers":
		family, mode = "user", "bulk"
	case "PUT /roles/:id", "DELETE /roles/:id":
		family = "role"
	case "DELETE /roles":
		family, mode = "role", "bulk"
	case "PUT /permissions/:key", "DELETE /permissions/:key":
		family = "permission"
	case "PUT /departments/:id", "DELETE /departments/:id":
		family = "department"
	case "PUT /positions/:id", "DELETE /positions/:id":
		family = "position"
	case "PUT /dict/items/:id", "DELETE /dict/items/:id":
		family = "dict_item"
	case "PUT /dict/types/:typeCode", "DELETE /dict/types/:typeCode", "PATCH /dict/types/:typeCode":
		family = "dict_type"
	case "DELETE /dict/types/:typeCode/items":
		family, mode = "dict_item", "dict_items"
	case "PUT /surveys/:id", "DELETE /surveys/:id", "PATCH /surveys/:id/status":
		family = "survey"
	case "PUT /exams/:id", "DELETE /exams/:id", "PATCH /exams/:id/status":
		family = "exam"
	case "PUT /workflow-definitions/:id", "DELETE /workflow-definitions/:id", "PATCH /workflow-definitions/:id/status", "POST /workflow-definitions/:id/publish", "POST /workflow-definitions/:id/versions/:version/rollback":
		family = "workflow_definition"
	case "PUT /settings", "PUT /settings/content":
		family = "settings"
	case "PUT /dingtalk/settings":
		family, mode = "settings", "dingtalk"
	case "PUT /workflow-org-approver-assignments":
		family = "org_approver"
	default:
		return snapshotSpec{}, false
	}
	spec := snapshotResource(family)
	if mode != "" {
		spec.mode = mode
	}
	return spec, true
}

var snapshotDingTalkKeys = strings.Fields(`DINGTALK_H5_CORP_ID DINGTALK_H5_APP_KEY DINGTALK_H5_APP_SECRET DINGTALK_H5_AGENT_ID DINGTALK_H5_UNIFIED_APP_ID DINGTALK_H5_NOTIFY_ENABLED DINGTALK_H5_NOTIFY_MODE DINGTALK_H5_ROBOT_CODE DINGTALK_H5_APP_URL TOKEN_DINGTALK_H5_EXPIRE TOKEN_DINGTALK_H5_REDIS_PREFIX DINGTALK_H5_SINGLE_LOGIN DINGTALK_H5_SELF_BIND_ENABLED DINGTALK_H5_LOGIN_CAPTCHA_ENABLED DINGTALK_H5_APP_NAME DINGTALK_H5_LOGO_TEXT DINGTALK_H5_LOGO_URL`)

func snapshotKnownSetupKey(key string) bool {
	for _, known := range snapshotDingTalkKeys {
		if key == known {
			return true
		}
	}
	switch key {
	case "STATIC_DOMAIN", "SETUP_USER_FORM_FIELDS", "SETUP_CONTENT_AGREEMENT", "SETUP_CONTENT_PRIVACY", "SETUP_CONTENT_ABOUT", "BUILTIN_TEMPLATE_PRESETS", "EXAM_BUILTIN_TEMPLATE_PRESETS", "HOME_PAGE_CONFIG", "TOKEN_USER_EXPIRE", "TOKEN_USER_REDIS_PREFIX", "TOKEN_ADMIN_EXPIRE", "TOKEN_ADMIN_REDIS_PREFIX", "USER_SINGLE_LOGIN", "ADMIN_SINGLE_LOGIN", "ADMIN_LOGIN_CAPTCHA_ENABLED", "ADMIN_FRONTEND_CONFIG":
		return true
	default:
		return false
	}
}

func snapshotRequestPlan(c *app.RequestContext, spec snapshotSpec) (snapshotPlan, string) {
	p := snapshotPlan{spec: spec}
	switch spec.mode {
	case "dingtalk":
		fields, err := snapshotRequestFields(c)
		if err != nil {
			return p, "request_unavailable"
		}
		scope := strings.ToLower(strings.TrimSpace(fields["scope"]))
		switch scope {
		case "", "all":
			p.ids = append([]string(nil), snapshotDingTalkKeys...)
			p.corpConfigs = true
		case "corp":
			p.ids = append([]string(nil), snapshotDingTalkKeys[:9]...)
			p.corpConfigs = true
		case "login":
			p.ids = append([]string(nil), snapshotDingTalkKeys[9:14]...)
		case "app":
			p.ids = append([]string(nil), snapshotDingTalkKeys[14:]...)
		default:
			return p, "invalid_target"
		}
	case "config", "bulk", "org":
		fields, err := snapshotRequestFields(c)
		if err != nil {
			return p, "request_unavailable"
		}
		switch spec.mode {
		case "config":
			key := fields["key"]
			if !snapshotKnownSetupKey(key) {
				return p, "unsupported_setup_key"
			}
			p.ids = []string{key}
		case "bulk":
			parts := strings.Split(fields["ids"], ",")
			if len(parts) > snapshotMaxTargets {
				return p, "target_limit"
			}
			seen := map[string]bool{}
			for _, raw := range parts {
				id, ok := snapshotID(strings.TrimSpace(raw))
				if !ok {
					return p, "invalid_target"
				}
				if !seen[id] {
					p.ids = append(p.ids, id)
					seen[id] = true
				}
			}
		case "org":
			p.subjectType = strings.TrimSpace(fields["subjectType"])
			id := fields["subjectId"]
			if p.subjectType != "user" && p.subjectType != "department" {
				p.subjectType, id = "department", fields["departmentId"]
			}
			id, ok := snapshotID(id)
			p.identity = strings.TrimSpace(fields["identityCode"])
			if !ok || !snapshotIdentifier(p.identity, 80) {
				return p, "invalid_target"
			}
			p.ids = []string{id}
		}
	case "dict_items":
		key := c.Param("typeCode")
		if !snapshotIdentifier(key, 50) {
			return p, "invalid_target"
		}
		p.ids = []string{key}
	default:
		key := c.Param("id")
		if spec.key == "permission_key" {
			key = c.Param("key")
		} else if spec.key == "dict_type_code" {
			key = c.Param("typeCode")
		} else {
			var ok bool
			key, ok = snapshotID(key)
			if !ok {
				return p, "invalid_target"
			}
		}
		if !snapshotIdentifier(key, 160) {
			return p, "invalid_target"
		}
		p.ids = []string{key}
		if spec.family == "permission" && string(c.Method()) == "PUT" {
			fields, err := snapshotRequestFields(c)
			if err != nil {
				return p, "request_unavailable"
			}
			newKey := strings.TrimSpace(fields["key"])
			if !snapshotIdentifier(newKey, 160) {
				return p, "invalid_target"
			}
			p.afterIDs = []string{newKey}
		}
		if c.Param("version") != "" {
			version, ok := snapshotID(c.Param("version"))
			if !ok {
				return p, "invalid_target"
			}
			p.relatedID = key + ":" + version
		}
	}
	return p, ""
}

func snapshotID(s string) (string, bool) {
	if len(s) == 0 || len(s) > 20 {
		return "", false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return "", false
		}
	}
	n, err := strconv.ParseUint(s, 10, 64)
	return strconv.FormatUint(n, 10), err == nil && n > 0
}

func snapshotIdentifier(s string, limit int) bool {
	if s == "" || len(s) > limit {
		return false
	}
	for _, r := range s {
		if r < 32 || r == 127 {
			return false
		}
	}
	return true
}

// Only routing selectors are retained. Never call Bind (which could consume or
// rewrite the handler's request), or parse multipart uploads in this middleware.
func snapshotRequestFields(c *app.RequestContext) (map[string]string, error) {
	if c.Request.IsBodyStream() || c.Request.Header.ContentLength() > snapshotMaxBodyBytes {
		return nil, errSnapshotLimit
	}
	body, err := c.Body()
	if err != nil || len(body) > snapshotMaxBodyBytes {
		return nil, errSnapshotLimit
	}
	result := map[string]string{}
	keys := []string{"key", "ids", "subjectType", "subjectId", "departmentId", "identityCode", "scope"}
	contentType := strings.ToLower(strings.TrimSpace(strings.Split(string(c.Request.Header.ContentType()), ";")[0]))
	switch contentType {
	case "application/x-www-form-urlencoded":
		values, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			result[key] = values.Get(key)
		}
	case "application/json":
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(body, &fields); err != nil {
			return nil, err
		}
		for _, key := range keys {
			raw := fields[key]
			if len(raw) == 0 {
				continue
			}
			var value string
			if json.Unmarshal(raw, &value) == nil {
				result[key] = value
				continue
			}
			if key == "subjectId" || key == "departmentId" {
				if id, ok := snapshotID(string(raw)); ok {
					result[key] = id
				}
			}
		}
	default:
		return nil, errors.New("unsupported snapshot body")
	}
	return result, nil
}

type snapshotQuery struct {
	table, key    string
	columns       []string
	ids           []string
	where         map[string]interface{}
	limit         int
	privateColumn string
	allRows       bool // only the fixed, bounded multi-corp settings collection
}

func snapshotQueries(p snapshotPlan) []snapshotQuery {
	s := p.spec
	q := snapshotQuery{table: s.table, key: s.key, columns: s.columns, ids: p.ids, limit: snapshotMaxTargets}
	switch s.mode {
	case "config", "dingtalk":
		q.privateColumn, q.limit = "setup_value", len(p.ids)
	case "org":
		q.where = map[string]interface{}{"subject_type": p.subjectType, "identity_code": p.identity}
		q.limit = snapshotMaxRelations
	case "dict_items":
		q.key = "dict_type_code"
	}
	queries := []snapshotQuery{q}
	if p.corpConfigs {
		queries = append(queries, snapshotQuery{table: "dingtalk_h5_corp_configs", key: "id", columns: strings.Fields("id corp_id corp_name app_key agent_id unified_app_id app_url notify_enabled notify_mode robot_code enabled"), privateColumn: "app_secret", limit: snapshotMaxTargets, allRows: true})
	}
	if s.family == "user" {
		queries = append(queries,
			snapshotQuery{table: "user_roles", key: "user_role_user_id", columns: strings.Fields("user_role_user_id user_role_role_id user_role_is_primary user_role_status"), ids: p.ids, limit: snapshotMaxRelations},
			snapshotQuery{table: "user_depts", key: "user_dept_user_id", columns: strings.Fields("user_dept_user_id user_dept_dept_id"), ids: p.ids, limit: snapshotMaxRelations},
			snapshotQuery{table: "user_reporting_relations", key: "employee_user_id", columns: strings.Fields("employee_user_id manager_user_id relation_type is_primary relation_sort relation_status effective_from effective_to"), ids: p.ids, limit: snapshotMaxRelations},
		)
	}
	if s.family == "user" || s.family == "role" {
		queries = append(queries, snapshotQuery{table: "permission_grants", key: "grant_subject_id", columns: strings.Fields("grant_subject_id grant_permission_key grant_effect grant_status"), privateColumn: "grant_scope_value", ids: p.ids, where: map[string]interface{}{"grant_subject_type": s.family}, limit: snapshotMaxRelations})
	}
	return queries
}

func snapshotSelect(db *gorm.DB, q snapshotQuery) *gorm.DB {
	// All identifiers and expressions originate in snapshotResource/Queries.
	columns := make([]string, 0, len(q.columns)+2)
	for _, column := range q.columns {
		columns = append(columns, "`"+column+"`")
	}
	if q.privateColumn != "" {
		column := "`" + q.privateColumn + "`"
		columns = append(columns,
			fmt.Sprintf("CASE WHEN OCTET_LENGTH(%s) <= %d THEN %s ELSE NULL END AS snapshot_private", column, snapshotMaxValueBytes, column),
			fmt.Sprintf("CASE WHEN OCTET_LENGTH(%s) > %d THEN 1 ELSE 0 END AS snapshot_oversize", column, snapshotMaxValueBytes),
		)
	}
	query := db.Table(q.table).Select(strings.Join(columns, ", "))
	if !q.allRows {
		query = query.Where(clause.IN{Column: clause.Column{Name: q.key}, Values: snapshotInterfaces(q.ids)})
	}
	if len(q.where) != 0 {
		query = query.Where(q.where)
	}
	for _, column := range q.columns {
		query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: column}})
	}
	return query.Limit(q.limit + 1)
}

func snapshotInterfaces(ids []string) []interface{} {
	values := make([]interface{}, len(ids))
	for i, id := range ids {
		values[i] = id
	}
	return values
}

func readAuditSnapshot(ctx context.Context, plan snapshotPlan) (snapshotData, error) {
	ctx, cancelRead := context.WithTimeout(ctx, snapshotReadTimeout)
	defer cancelRead()
	db, cancelDB := database.WithContext(ctx)
	defer cancelDB()
	if db == nil {
		return snapshotData{}, errors.New("snapshot database unavailable")
	}
	// Avoid ORM error logging of rows/parameters. The collector gets fixed state.
	db = db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
	data := snapshotData{values: map[string]interface{}{}, private: map[string]string{}}
	for i, query := range snapshotQueries(plan) {
		var rows []map[string]interface{}
		if err := snapshotSelect(db, query).Find(&rows).Error; err != nil {
			return snapshotData{}, err
		}
		if len(rows) > query.limit {
			return snapshotData{}, errSnapshotLimit
		}
		if err := snapshotRows(&data, plan, query, rows, i == 0); err != nil {
			return snapshotData{}, err
		}
	}
	return data, nil
}

func snapshotRows(data *snapshotData, p snapshotPlan, q snapshotQuery, rows []map[string]interface{}, primary bool) error {
	if len(rows) > q.limit {
		return errSnapshotLimit
	}
	for _, row := range rows {
		if snapshotString(row["snapshot_oversize"]) == "1" {
			return errSnapshotLimit
		}
	}
	if q.table == "dingtalk_h5_corp_configs" {
		for _, row := range rows {
			id, ok := snapshotID(snapshotString(row["id"]))
			if !ok {
				return errors.New("invalid stored config target")
			}
			prefix := "settings.corp." + id + "."
			for _, column := range q.columns {
				if column == "id" {
					continue
				}
				value := snapshotString(row[column])
				if len(value) > 512 {
					return errSnapshotLimit
				}
				snapshotSettingValue(data, prefix+column, row[column])
			}
			data.private[prefix+"app_secret"] = snapshotString(row["snapshot_private"])
		}
		return nil
	}
	if primary && p.spec.family == "settings" {
		for _, row := range rows {
			key := snapshotString(row[q.key])
			if !snapshotKnownSetupKey(key) {
				return errors.New("unknown setup key")
			}
			snapshotSettingValue(data, "settings."+key, row["snapshot_private"])
		}
		data.count = len(rows)
		return nil
	}
	if primary && p.spec.mode != "org" {
		data.count = len(rows)
		for _, row := range rows {
			identityColumn := p.spec.key
			if p.spec.family == "permission" {
				identityColumn = "id"
			}
			id := snapshotString(row[identityColumn])
			if !snapshotIdentifier(id, 160) {
				return errors.New("invalid stored target")
			}
			prefix := p.spec.family + "." + id + "."
			data.values[prefix+"exists"] = true
			for _, column := range q.columns {
				if column == identityColumn {
					continue
				}
				value := snapshotString(row[column])
				if len(value) > 512 {
					return errSnapshotLimit
				}
				data.values[prefix+column] = value
			}
			if len(rows) == 1 && len(p.ids) == 1 {
				data.name = snapshotString(row[p.spec.name])
			}
		}
		return nil
	}
	// Relation IDs/timestamps change on replacement even when grants are identical;
	// only natural business fields participate, sorted for stable set comparison.
	groups := map[string][]map[string]interface{}{}
	for _, id := range p.ids {
		groups[id] = []map[string]interface{}{}
	}
	for _, row := range rows {
		id := snapshotString(row[q.key])
		if _, ok := groups[id]; !ok {
			return errors.New("unexpected relation target")
		}
		item := map[string]interface{}{}
		for _, column := range q.columns {
			if column == q.key {
				continue
			}
			value := snapshotString(row[column])
			if len(value) > 512 {
				return errSnapshotLimit
			}
			item[column] = value
		}
		if q.table == "permission_grants" {
			scope, err := snapshotScope(snapshotString(row["snapshot_private"]))
			if err != nil {
				return err
			}
			item["scope"] = scope
		}
		groups[id] = append(groups[id], item)
	}
	for id, items := range groups {
		sort.Slice(items, func(i, j int) bool {
			a, _ := json.Marshal(items[i])
			b, _ := json.Marshal(items[j])
			return string(a) < string(b)
		})
		data.values[p.spec.family+"."+id+"."+q.table] = items
	}
	return nil
}

func snapshotSettingValue(data *snapshotData, field string, value interface{}) {
	// SQL drivers may return text as bytes. Preserve other types for validation;
	// stringifying arbitrary values could turn invalid values into public scalars.
	if bytes, ok := value.([]byte); ok {
		value = string(bytes)
	}
	if public, ok := auditlog.PublicSnapshotValue(field, value); ok {
		data.values[field] = public
	} else {
		data.private[field] = snapshotString(value)
	}
}

func snapshotScope(raw string) (interface{}, error) {
	if len(raw) > snapshotMaxValueBytes {
		return nil, errSnapshotLimit
	}
	var scope struct {
		DeptIDs []uint64 `json:"deptIds"`
		UserIDs []uint64 `json:"userIds"`
	}
	if strings.TrimSpace(raw) != "" && json.Unmarshal([]byte(raw), &scope) != nil {
		return nil, errors.New("invalid grant scope")
	}
	if len(scope.DeptIDs)+len(scope.UserIDs) > snapshotMaxTargets*5 {
		return nil, errSnapshotLimit
	}
	return map[string]interface{}{"deptIds": snapshotSortedIDs(scope.DeptIDs), "userIds": snapshotSortedIDs(scope.UserIDs)}, nil
}

func snapshotSortedIDs(ids []uint64) []uint64 {
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	result := []uint64{}
	for _, id := range ids {
		if id > 0 && (len(result) == 0 || result[len(result)-1] != id) {
			result = append(result, id)
		}
	}
	return result
}

func snapshotString(value interface{}) string {
	if value == nil {
		return ""
	}
	if bytes, ok := value.([]byte); ok {
		return string(bytes)
	}
	return fmt.Sprint(value)
}

func snapshotDiff(before, after snapshotData) ([]auditlog.Change, error) {
	keys := map[string]bool{}
	for key := range before.values {
		keys[key] = true
	}
	for key := range after.values {
		keys[key] = true
	}
	for key := range before.private {
		keys[key] = true
	}
	for key := range after.private {
		keys[key] = true
	}
	ordered := make([]string, 0, len(keys))
	for key := range keys {
		ordered = append(ordered, key)
	}
	sort.Strings(ordered)
	var changes []auditlog.Change
	for _, key := range ordered {
		oldPrivate, oldIsPrivate := before.private[key]
		newPrivate, newIsPrivate := after.private[key]
		if oldIsPrivate || newIsPrivate {
			if oldIsPrivate != newIsPrivate || oldPrivate != newPrivate {
				changes = append(changes, auditlog.Change{Field: key, Before: "[已隐藏]", After: "[已变更]"})
			}
		} else if !reflect.DeepEqual(before.values[key], after.values[key]) {
			changes = append(changes, auditlog.Change{Field: key, Before: before.values[key], After: after.values[key]})
		}
		if len(changes) > snapshotMaxChanges {
			return nil, errSnapshotLimit
		}
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Field < changes[j].Field })
	raw, err := json.Marshal(changes)
	if err != nil || len(raw) > snapshotMaxDiffBytes {
		return nil, errSnapshotLimit
	}
	return changes, nil
}
