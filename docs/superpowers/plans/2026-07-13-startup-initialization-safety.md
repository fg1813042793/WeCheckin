# 启动初始化安全优化实施计划

> **给 agentic workers：** 必须使用子技能：推荐使用 `superpowers:subagent-driven-development`，或使用 `superpowers:executing-plans` 按任务逐项执行。本计划使用复选框语法跟踪执行状态。

**目标：** 移除后端启动流程中的无条件删表 SQL，让启动初始化只做自动迁移和幂等种子初始化。

**架构：** 保持 `service.InitBusiness(enableExam bool)` 的公开接口不变。通过源码级安全护栏测试防止 `backend/internal/app/service/database.go` 再次出现 `DROP TABLE` 或 `user_form_fields` 清理逻辑，然后删除当前危险 SQL，并补充中文文档说明用户扩展表单字段的现行存储位置。

**技术栈：** Go 1.24、Go testing、GORM 启动初始化代码、Markdown 中文文档。

---

### 任务 1：新增启动初始化安全护栏测试

**文件：**
- 新建：`backend/internal/app/service/database_safety_test.go`
- 读取：`backend/internal/app/service/database.go`

- [ ] **步骤 1：写入失败测试**

创建 `backend/internal/app/service/database_safety_test.go`，内容如下：

```go
package service

import (
	"os"
	"strings"
	"testing"
)

func TestStartupInitializationDoesNotDropLegacyUserFormFields(t *testing.T) {
	src, err := os.ReadFile("database.go")
	if err != nil {
		t.Fatalf("read database.go: %v", err)
	}
	text := strings.ToLower(string(src))
	if strings.Contains(text, "drop table") {
		t.Fatalf("startup initialization must not contain DROP TABLE statements")
	}
	if strings.Contains(text, "user_form_fields") {
		t.Fatalf("startup initialization must not clean legacy user_form_fields table")
	}
}
```

- [ ] **步骤 2：运行测试确认 RED**

运行：

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/service -run TestStartupInitializationDoesNotDropLegacyUserFormFields -count=1
```

预期：测试失败，失败原因包含 `DROP TABLE` 或 `user_form_fields`。

### 任务 2：移除启动流程中的危险 SQL

**文件：**
- 修改：`backend/internal/app/service/database.go`
- 测试：`backend/internal/app/service/database_safety_test.go`

- [ ] **步骤 1：删除无条件删表语句**

从 `InitBusiness` 中删除这一行：

```go
database.DB.Exec("DROP TABLE IF EXISTS `user_form_fields`")
```

保留函数顺序：

```go
func InitBusiness(enableExam bool) {
	if err := autoMigrate(); err != nil {
		log.Printf("Migration warning: %v (continuing)", err)
	}

	seedSetups()
	seedMenus(enableExam)
}
```

- [ ] **步骤 2：运行聚焦测试确认 GREEN**

运行：

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/service -run TestStartupInitializationDoesNotDropLegacyUserFormFields -count=1
```

预期：测试通过。

### 任务 3：补充中文 README 说明

**文件：**
- 修改：`README.md`

- [ ] **步骤 1：追加用户扩展表单字段说明**

在配置说明段落中追加：

```markdown
用户扩展表单字段通过 `setups` 表中的 `SETUP_USER_FORM_FIELDS` 配置项保存。后端启动不会清理旧 `user_form_fields` 表；如需迁移历史数据，应使用单独迁移脚本处理。
```

### 任务 4：最终验证和清理

**文件：**
- 读取：工作区状态

- [ ] **步骤 1：运行完整范围验证**

运行：

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/service ./backend/internal/config ./backend/internal/app/formkit/...
```

预期：所有列出的包通过。

- [ ] **步骤 2：删除测试缓存目录**

运行：

```bash
rm -rf .cache
```

预期：`.cache/` 不存在。

- [ ] **步骤 3：检查最终 diff**

运行：

```bash
git diff -- backend/internal/app/service/database.go backend/internal/app/service/database_safety_test.go README.md docs/superpowers/plans/2026-07-13-startup-initialization-safety.md
```

预期：只包含本轮启动初始化安全优化改动。
