package dingtalkh5

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/passwordutil"
)

type defaultUser struct {
	Account                string
	Name                   string
	Position               string
	DepartmentLevel1       string
	DepartmentLevel2       string
	DepartmentLevel3       string
	ManagerID              string
	HRBPID                 string
	ResponsibleDepartments []string
}

func defaultTemplate() TemplateDTO {
	return TemplateDTO{
		ObjectiveDefaults: []NextObjective{
			{ID: "tpl-1", Target: "及时高效完成客户端版本迭代和其他人员所需的服务端支持", Weight: 40},
			{ID: "tpl-2", Target: "保证线上服务稳定，线上无重大事故，轻微影响使用的 bug 不超过 3 个", Weight: 30},
			{ID: "tpl-3", Target: "结合聊天上下文提升对话效果，提高使用 3 次用户占比到 50%", Weight: 30},
		},
		NextObjectiveDefaults: []NextObjective{
			{ID: "next-1", Target: "及时高效完成下月版本迭代和服务端支持", Weight: 40},
			{ID: "next-2", Target: "保证线上服务稳定，持续降低线上问题数量", Weight: 30},
			{ID: "next-3", Target: "分析竞品并积极调研，从技术角度提出 3 个优化需求", Weight: 30},
		},
		GradeLevels: []GradeLevel{
			{Label: "优秀", Grade: "A+", Coefficient: 1.5},
			{Label: "优秀", Grade: "A-", Coefficient: 1.4},
			{Label: "良好", Grade: "B+", Coefficient: 1.3},
			{Label: "良好", Grade: "B-", Coefficient: 1.2},
			{Label: "及格", Grade: "C+", Coefficient: 1.1},
			{Label: "及格", Grade: "C", Coefficient: 1},
			{Label: "及格", Grade: "C-", Coefficient: 0.9},
			{Label: "较差", Grade: "D+", Coefficient: 0.8},
			{Label: "较差", Grade: "D-", Coefficient: 0.7},
			{Label: "糟糕", Grade: "E+", Coefficient: 0.6},
			{Label: "糟糕", Grade: "E-", Coefficient: 0.5},
		},
		Values: []ValueTemplate{
			{ID: "team", Name: "团结一心", Definition: "从不抱怨、拥抱变化、积极主动、换位思考、自我批评、火车头精神", Rubric: defaultValueRubric()},
			{ID: "innovation", Name: "开拓创新", Definition: "要事第一、开疆拓土、迭代创新、唯快不破", Rubric: defaultValueRubric()},
			{ID: "grit", Name: "坚韧不拔", Definition: "思想坚定、不畏困难、死磕到底、永不放弃", Rubric: defaultValueRubric()},
		},
	}
}

func defaultValueRubric() []ValueRubric {
	return []ValueRubric{
		{Label: "卓越", Score: 50, Description: "持续超出要求，对团队或业务产生明显正向影响"},
		{Label: "优秀", Score: 40, Description: "高质量完成要求，表现稳定且有主动贡献"},
		{Label: "良好", Score: 30, Description: "符合岗位要求，能够稳定完成相关表现"},
		{Label: "及格", Score: 20, Description: "基本达到要求，但仍有明显提升空间"},
		{Label: "较差", Score: 10, Description: "未达到要求，需要重点改进"},
	}
}

func defaultUsers() []defaultUser {
	return []defaultUser{
		{Account: "lip", Name: "Lip", Position: "Java 工程师", DepartmentLevel1: "M/H业务", DepartmentLevel2: "研发部", DepartmentLevel3: "Java开发一组", ManagerID: "cube", HRBPID: "lucky"},
		{Account: "arthur", Name: "Arthur", Position: "Android 工程师", DepartmentLevel1: "M/H业务", DepartmentLevel2: "研发部", DepartmentLevel3: "安卓开发一组", ManagerID: "neil", HRBPID: "nick"},
		{Account: "foster", Name: "Foster", Position: "运维工程师", DepartmentLevel1: "M/H业务", DepartmentLevel2: "研发部", DepartmentLevel3: "运维组", ManagerID: "david", HRBPID: "hrbp"},
		{Account: "rock", Name: "Rock", Position: "Java 工程师", DepartmentLevel1: "M/H业务", DepartmentLevel2: "研发部", DepartmentLevel3: "Java开发一组", ManagerID: "paul", HRBPID: "nick"},
		{Account: "cube", Name: "Cube", Position: "产品主管", DepartmentLevel1: "M/H业务", DepartmentLevel2: "产品部", DepartmentLevel3: "国内组", HRBPID: "hrbp"},
		{Account: "sherif", Name: "Sherif", Position: "Android 主管", DepartmentLevel1: "M/H业务", DepartmentLevel2: "研发部", DepartmentLevel3: "安卓开发二组", ManagerID: "david", HRBPID: "nick"},
		{Account: "paul", Name: "Paul", Position: "Java 主管", DepartmentLevel1: "M/H业务", DepartmentLevel2: "研发部", DepartmentLevel3: "Java开发一组", ManagerID: "david", HRBPID: "nick"},
		{Account: "neil", Name: "Neil", Position: "Android 主管", DepartmentLevel1: "M/H业务", DepartmentLevel2: "研发部", DepartmentLevel3: "安卓开发一组", ManagerID: "david", HRBPID: "nick"},
		{Account: "david", Name: "David", Position: "研发经理", DepartmentLevel1: "M/H业务", DepartmentLevel2: "研发部", HRBPID: "nick"},
		{Account: "hrbp", Name: "HRBP", Position: "HRBP", DepartmentLevel1: "M/H业务", ResponsibleDepartments: []string{"研发部", "产品部"}},
		{Account: "nick", Name: "Nick", Position: "HRBP 管理员", DepartmentLevel1: "M/H业务", ResponsibleDepartments: []string{"研发部", "综合部"}},
		{Account: "monica", Name: "Monica", Position: "HRBP", DepartmentLevel1: "M/H业务", DepartmentLevel2: "综合部", ManagerID: "nick", HRBPID: "nick", ResponsibleDepartments: []string{"产品部"}},
		{Account: "lucky", Name: "Lucky", Position: "HRBP", DepartmentLevel1: "M/H业务", DepartmentLevel2: "综合部", ManagerID: "nick", HRBPID: "nick", ResponsibleDepartments: []string{"产品部"}},
		{Account: "betty", Name: "Betty", Position: "员工", DepartmentLevel1: "M/H业务线", DepartmentLevel2: "综合部"},
		{Account: "cherry", Name: "Cherry", Position: "管理员", DepartmentLevel1: "M/H业务", DepartmentLevel2: "综合部"},
		{Account: "amy", Name: "Amy", Position: "主管", DepartmentLevel1: "M/H业务", DepartmentLevel2: "综合部"},
	}
}

func EnsureSeedContext(ctx context.Context) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return nil
	}

	if err := ensureDefaultTemplateDB(db); err != nil {
		return err
	}
	now := database.Now()

	hash, err := passwordutil.Hash("123456")
	if err != nil {
		return err
	}
	for _, item := range defaultUsers() {
		if err := upsertDefaultPerfUser(db, item, hash, now); err != nil {
			return err
		}
	}
	return nil
}

func upsertDefaultPerfUser(db *gorm.DB, item defaultUser, passwordHash string, now int64) error {
	user := model.DingTalkH5PerfUser{
		Account:                item.Account,
		Name:                   item.Name,
		Password:               passwordHash,
		Position:               item.Position,
		Department:             departmentText(item.DepartmentLevel1, item.DepartmentLevel2, item.DepartmentLevel3),
		DepartmentLevel1:       item.DepartmentLevel1,
		DepartmentLevel2:       item.DepartmentLevel2,
		DepartmentLevel3:       item.DepartmentLevel3,
		ManagerAccount:         item.ManagerID,
		HRBPAccount:            item.HRBPID,
		ResponsibleDepartments: encodeJSON(item.ResponsibleDepartments),
		Status:                 1,
		AddTime:                now,
		EditTime:               now,
	}
	var existing model.User
	err := db.Where("`user_mini_openid` = ?", item.Account).First(&existing).Error
	if err == nil {
		updates := map[string]interface{}{
			"user_name":      item.Name,
			"user_obj":       encodePerfUserObj(existing.Obj, user),
			"user_edit_time": now,
		}
		if strings.TrimSpace(existing.Password) == "" {
			updates["user_password"] = passwordHash
		}
		return db.Model(&model.User{}).Where("`id` = ?", existing.ID).Updates(updates).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	baseUser := model.User{
		MiniOpenID: item.Account,
		Name:       item.Name,
		Password:   passwordHash,
		Pic:        "/static/default-avatar.png",
		Forms:      "[]",
		Obj:        encodePerfUserObj("", user),
		Status:     1,
		AddTime:    now,
		AddIP:      "127.0.0.1",
		EditTime:   now,
		EditIP:     "127.0.0.1",
	}
	return db.Create(&baseUser).Error
}

func ensureDefaultTemplateContext(ctx context.Context) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return ensureDefaultTemplateDB(db)
}

func ensureDefaultTemplateDB(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	var tpl model.DingTalkH5PerfTemplate
	err := db.Where("template_key = ?", TemplateKeyDefault).First(&tpl).Error
	if err == nil {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	now := database.Now()
	tpl = model.DingTalkH5PerfTemplate{
		Key:      TemplateKeyDefault,
		Payload:  encodeJSON(defaultTemplate()),
		AddTime:  now,
		EditTime: now,
	}
	applyDingTalkH5CreateAudit(&tpl.DingTalkH5AuditFields, dingtalkH5AuditMetaForUserContext(context.Background(), db, nil, now))
	return db.Create(&tpl).Error
}

func LoadTemplateContext(ctx context.Context) (TemplateDTO, error) {
	if err := ensureDefaultTemplateContext(ctx); err != nil {
		return TemplateDTO{}, err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var tpl model.DingTalkH5PerfTemplate
	if err := db.Where("template_key = ?", TemplateKeyDefault).First(&tpl).Error; err != nil {
		return TemplateDTO{}, err
	}
	var result TemplateDTO
	if err := jsonUnmarshal(tpl.Payload, &result); err != nil {
		return TemplateDTO{}, err
	}
	return result, nil
}

func SaveTemplateContext(ctx context.Context, user *model.DingTalkH5PerfUser, payload TemplateDTO) (TemplateDTO, error) {
	result, err := sanitizeTemplate(payload)
	if err != nil {
		return TemplateDTO{}, err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	now := database.Now()
	audit := dingtalkH5AuditMetaForUserContext(ctx, db, user, now)
	raw := encodeJSON(result)
	var tpl model.DingTalkH5PerfTemplate
	err = db.Where("template_key = ?", TemplateKeyDefault).First(&tpl).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		tpl = model.DingTalkH5PerfTemplate{
			Key:      TemplateKeyDefault,
			Payload:  raw,
			AddTime:  now,
			EditTime: now,
		}
		applyDingTalkH5CreateAudit(&tpl.DingTalkH5AuditFields, audit)
		return result, db.Create(&tpl).Error
	}
	if err != nil {
		return TemplateDTO{}, err
	}
	applyDingTalkH5UpdateAudit(&tpl.DingTalkH5AuditFields, audit)
	updates := dingtalkH5AuditUpdateValues(audit)
	updates["payload"] = raw
	updates["template_key"] = TemplateKeyDefault
	if tpl.CreateBy == 0 {
		for key, value := range dingtalkH5CreateAuditValues(audit) {
			updates[key] = value
		}
	}
	if err := db.Model(&tpl).Updates(updates).Error; err != nil {
		return TemplateDTO{}, err
	}
	return result, nil
}

func sanitizeTemplate(input TemplateDTO) (TemplateDTO, error) {
	result := TemplateDTO{
		ObjectiveDefaults:     sanitizeTemplateObjectives(input.ObjectiveDefaults, "objective"),
		NextObjectiveDefaults: sanitizeTemplateObjectives(input.NextObjectiveDefaults, "next"),
		GradeLevels:           sanitizeGradeLevels(input.GradeLevels),
		Values:                sanitizeValueTemplates(input.Values),
	}
	if len(result.ObjectiveDefaults) == 0 && len(result.NextObjectiveDefaults) == 0 && len(result.GradeLevels) == 0 && len(result.Values) == 0 {
		return TemplateDTO{}, errors.New("模板内容不能为空")
	}
	return result, nil
}

func sanitizeTemplateObjectives(items []NextObjective, prefix string) []NextObjective {
	result := make([]NextObjective, 0, len(items))
	for _, item := range items {
		target := strings.TrimSpace(item.Target)
		if target == "" {
			continue
		}
		id := strings.TrimSpace(item.ID)
		if id == "" {
			id = fmt.Sprintf("%s-%d", prefix, len(result)+1)
		}
		result = append(result, NextObjective{
			ID:     id,
			Target: target,
			Weight: clampTemplateNumber(item.Weight, 0, 100),
		})
	}
	return result
}

func sanitizeGradeLevels(items []GradeLevel) []GradeLevel {
	result := make([]GradeLevel, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Label)
		grade := strings.TrimSpace(item.Grade)
		if label == "" && grade == "" {
			continue
		}
		result = append(result, GradeLevel{
			Label:       label,
			Grade:       grade,
			Coefficient: clampTemplateNumber(item.Coefficient, 0, 10),
		})
	}
	return result
}

func sanitizeValueTemplates(items []ValueTemplate) []ValueTemplate {
	result := make([]ValueTemplate, 0, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item.Name)
		definition := strings.TrimSpace(item.Definition)
		if name == "" && definition == "" {
			continue
		}
		id := strings.TrimSpace(item.ID)
		if id == "" {
			id = fmt.Sprintf("value-%d", len(result)+1)
		}
		result = append(result, ValueTemplate{
			ID:         id,
			Name:       name,
			Definition: definition,
			Rubric:     sanitizeValueRubric(item.Rubric),
		})
	}
	return result
}

func sanitizeValueRubric(items []ValueRubric) []ValueRubric {
	result := make([]ValueRubric, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Label)
		if label == "" {
			continue
		}
		result = append(result, ValueRubric{
			Label:       label,
			Score:       clampTemplateNumber(item.Score, 0, 100),
			Description: strings.TrimSpace(item.Description),
		})
	}
	return result
}

func clampTemplateNumber(value, min, max float64) float64 {
	if value < min {
		value = min
	}
	if value > max {
		value = max
	}
	return math.Round(value*10) / 10
}

func departmentText(parts ...string) string {
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	return strings.Join(items, " / ")
}
