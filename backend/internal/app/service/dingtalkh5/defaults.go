package dingtalkh5

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/passwordutil"
)

type defaultUser struct {
	Account                string
	Name                   string
	Role                   string
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
		{Label: "卓越", Score: 50},
		{Label: "优秀", Score: 40},
		{Label: "良好", Score: 30},
		{Label: "及格", Score: 20},
		{Label: "较差", Score: 10},
	}
}

func defaultUsers() []defaultUser {
	return []defaultUser{
		{Account: "lip", Name: "Lip", Role: "employee", Position: "Java 工程师", DepartmentLevel1: "M/H业务", DepartmentLevel2: "研发部", DepartmentLevel3: "Java开发一组", ManagerID: "cube", HRBPID: "lucky"},
		{Account: "arthur", Name: "Arthur", Role: "employee", Position: "Android 工程师", DepartmentLevel1: "M/H业务", DepartmentLevel2: "研发部", DepartmentLevel3: "安卓开发一组", ManagerID: "neil", HRBPID: "nick"},
		{Account: "foster", Name: "Foster", Role: "employee", Position: "运维工程师", DepartmentLevel1: "M/H业务", DepartmentLevel2: "研发部", DepartmentLevel3: "运维组", ManagerID: "david", HRBPID: "hrbp"},
		{Account: "rock", Name: "Rock", Role: "employee", Position: "Java 工程师", DepartmentLevel1: "M/H业务", DepartmentLevel2: "研发部", DepartmentLevel3: "Java开发一组", ManagerID: "paul", HRBPID: "nick"},
		{Account: "cube", Name: "Cube", Role: "supervisor", Position: "产品主管", DepartmentLevel1: "M/H业务", DepartmentLevel2: "产品部", DepartmentLevel3: "国内组", HRBPID: "hrbp"},
		{Account: "sherif", Name: "Sherif", Role: "supervisor", Position: "Android 主管", DepartmentLevel1: "M/H业务", DepartmentLevel2: "研发部", DepartmentLevel3: "安卓开发二组", ManagerID: "david", HRBPID: "nick"},
		{Account: "paul", Name: "Paul", Role: "supervisor", Position: "Java 主管", DepartmentLevel1: "M/H业务", DepartmentLevel2: "研发部", DepartmentLevel3: "Java开发一组", ManagerID: "david", HRBPID: "nick"},
		{Account: "neil", Name: "Neil", Role: "supervisor", Position: "Android 主管", DepartmentLevel1: "M/H业务", DepartmentLevel2: "研发部", DepartmentLevel3: "安卓开发一组", ManagerID: "david", HRBPID: "nick"},
		{Account: "david", Name: "David", Role: "manager", Position: "研发经理", DepartmentLevel1: "M/H业务", DepartmentLevel2: "研发部", HRBPID: "nick"},
		{Account: "hrbp", Name: "HRBP", Role: "hrbp", Position: "HRBP", DepartmentLevel1: "M/H业务", ResponsibleDepartments: []string{"研发部", "产品部"}},
		{Account: "nick", Name: "Nick", Role: "admin", Position: "HRBP 管理员", DepartmentLevel1: "M/H业务", ResponsibleDepartments: []string{"研发部", "综合部"}},
		{Account: "monica", Name: "Monica", Role: "hrbp", Position: "HRBP", DepartmentLevel1: "M/H业务", DepartmentLevel2: "综合部", ManagerID: "nick", HRBPID: "nick", ResponsibleDepartments: []string{"产品部"}},
		{Account: "lucky", Name: "Lucky", Role: "hrbp", Position: "HRBP", DepartmentLevel1: "M/H业务", DepartmentLevel2: "综合部", ManagerID: "nick", HRBPID: "nick", ResponsibleDepartments: []string{"产品部"}},
		{Account: "betty", Name: "Betty", Role: "employee", Position: "员工", DepartmentLevel1: "M/H业务线", DepartmentLevel2: "综合部"},
		{Account: "cherry", Name: "Cherry", Role: "admin", Position: "管理员", DepartmentLevel1: "M/H业务", DepartmentLevel2: "综合部"},
		{Account: "amy", Name: "Amy", Role: "supervisor", Position: "主管", DepartmentLevel1: "M/H业务", DepartmentLevel2: "综合部"},
	}
}

func EnsureSeedContext(ctx context.Context) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return nil
	}

	var templateCount int64
	if err := db.Model(&model.DingTalkH5PerfTemplate{}).Count(&templateCount).Error; err != nil {
		return err
	}
	now := database.Now()
	if templateCount == 0 {
		tpl := model.DingTalkH5PerfTemplate{
			Key:      TemplateKeyDefault,
			Payload:  encodeJSON(defaultTemplate()),
			AddTime:  now,
			EditTime: now,
		}
		if err := db.Create(&tpl).Error; err != nil {
			return err
		}
	}

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
		Role:                   item.Role,
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
	if err != gorm.ErrRecordNotFound {
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

func LoadTemplateContext(ctx context.Context) (TemplateDTO, error) {
	if err := EnsureSeedContext(ctx); err != nil {
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
