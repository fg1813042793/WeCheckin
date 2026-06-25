## 1. 准备与基础设施

- [ ] 1.1 创建所有模块目录（`internal/{passport,user,news,event,enroll,exam,survey,department,role,menu,dict,admin,setup,fav,geo,home}/`）
- [ ] 1.2 `git rm` 清理构建产物（`main.exe`、`wecheckin-server`、`stderr.txt`、`stdout.txt`）
- [ ] 1.3 `git rm -r` 清理备份目录（`mcloud_bak/`）

## 2. 迁移核心模块 handler + service（`internal/api/` + `internal/service/` → 各模块）

- [ ] 2.1 `geo`：`internal/api/client/geo.go` → `internal/geo/handler.go`
- [ ] 2.2 `passport`：`internal/api/client/passport.go` + `internal/service/passport.go` → `internal/passport/{handler,service}.go`
- [ ] 2.3 `home`：`internal/api/admin/admin_home.go` + `internal/api/client/home.go` + `internal/service/home.go` → `internal/home/{handler,service}.go`
- [ ] 2.4 `news`：`internal/api/admin/admin_news.go` + `internal/api/client/news.go` + `internal/service/news.go` → `internal/news/{handler,service}.go`
- [ ] 2.5 `event`：`internal/api/admin/admin_event.go` + `internal/api/client/event.go` + `internal/service/event.go` → `internal/event/{handler,service}.go`
- [ ] 2.6 `enroll`：`internal/api/admin/admin_enroll.go` + `internal/api/client/enroll.go` + `internal/service/enroll.go` → `internal/enroll/{handler,service}.go`
- [ ] 2.7 `fav`：`internal/api/client/fav.go` + `internal/service/fav.go` → `internal/fav/{handler,service}.go`
- [ ] 2.8 `user`：`internal/api/admin/admin_user.go` + `internal/service/admin.go`（用户相关部分）→ `internal/user/{handler,service}.go`
- [ ] 2.9 `department`：`internal/api/admin/admin_department.go` + `internal/service/admin.go`（部门相关部分）→ `internal/department/{handler,service}.go`
- [ ] 2.10 `role`：`internal/api/admin/admin_role.go` + `internal/service/admin.go`（角色相关部分）→ `internal/role/{handler,service}.go`
- [ ] 2.11 `menu`：`internal/api/admin/admin_menu.go` + `internal/service/admin.go`（菜单相关部分）→ `internal/menu/{handler,service}.go`
- [ ] 2.12 `dict`：`internal/api/admin/admin_dict.go` + `internal/service/admin.go`（字典相关部分）→ `internal/dict/{handler,service}.go`
- [ ] 2.13 `admin`（管理员管理）：`internal/api/admin/admin_mgr.go` + `internal/service/admin.go`（管理员相关部分）→ `internal/admin/{handler,service}.go`
- [ ] 2.14 `setup`：`internal/api/admin/admin_setup.go` + `internal/service/home.go`（设置相关部分）→ `internal/setup/{handler,service}.go`

## 3. 迁移 exam 模块

- [ ] 3.1 合并 `internal/exam/api/admin_exam.go` + `internal/exam/api/client_exam.go` → `internal/exam/`（移除 api/ 子目录）
- [ ] 3.2 合并 `internal/exam/service/` → `internal/exam/`（移除 service/ 子目录）

## 4. 迁移 survey 模块

- [ ] 4.1 合并 `internal/survey/api/admin_survey.go` + `internal/survey/api/client_survey.go` + `internal/survey/api/admin_formkit.go` → `internal/survey/`（移除 api/ 子目录）
- [ ] 4.2 合并 `internal/survey/service/` → `internal/survey/`（移除 service/ 子目录）

## 5. 更新 model 和共享代码

- [ ] 5.1 评估 `internal/model/models.go` 中的模型：拆分模块专用 model 到各模块，共享模型保留 `internal/model/`
- [ ] 5.2 更新 `pkg/response/` 等 import 路径（如无需调整则跳过）

## 6. 更新路由注册

- [ ] 6.1 更新 `cmd/main.go` 中所有 import 路径为新模块路径
- [ ] 6.2 更新 `cmd/main.go` 中所有 handler 初始化引用新 package 名
- [ ] 6.3 验证路由路径未改变

## 7. 编译验证

- [ ] 7.1 执行 `go build ./...` 确认编译通过
- [ ] 7.2 修复所有编译错误
