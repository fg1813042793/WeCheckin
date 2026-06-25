# Implementation Tasks

## 1. 更新 cmd/main.go 的 tag.name 定义
- [ ] `<backend/cmd/main.go>`: 改行 9 `@tag.name 管理端 API` → `@tag.name PC端`
- [ ] `<backend/cmd/main.go>`: 改行 11 `@tag.name 客户端 API` → `@tag.name 客户端`

## 2. 替换所有客户端 Tags（"赛事活动-客户端"/"考试-客户端"/"问卷-客户端" → 去掉"-客户端"后缀）
- [ ] `<backend/internal/api/client/event.go>`: `赛事活动-客户端, 客户端 API` → `客户端, 赛事活动`
- [ ] `<backend/internal/exam/api/client_exam.go>`: `考试-客户端, 客户端 API` → `客户端, 考试`
- [ ] `<backend/internal/survey/api/client_survey.go>`: `问卷-客户端, 客户端 API` → `客户端, 问卷`

## 3. 替换其他客户端 Tags
- [ ] `<backend/internal/api/client/passport.go>`: `通行证, 客户端 API` → `客户端, 通行证`
- [ ] `<backend/internal/api/client/geo.go>`: `地理编码, 客户端 API` → `客户端, 地理编码`
- [ ] `<backend/internal/api/client/fav.go>`: `收藏, 客户端 API` → `客户端, 收藏`
- [ ] `<backend/internal/api/client/news.go>`: `新闻, 客户端 API` → `客户端, 新闻`
- [ ] `<backend/internal/api/client/home.go>`: `首页, 客户端 API` → `客户端, 首页`
- [ ] `<backend/internal/api/client/enroll.go>`: `@Tags 报名` → `@Tags 客户端, 报名`
- [ ] `<backend/internal/survey/api/client_survey.go>`: `@Tags 表单工具` → `@Tags 客户端, 表单工具` (仅无分类的 2 处)

## 4. 替换所有管理端 Tags
- [ ] `<backend/internal/api/admin/admin_user.go>`: `用户管理, 管理端 API` → `PC端, 用户管理`；`在线用户, 管理端 API` → `PC端, 在线用户`
- [ ] `<backend/internal/api/admin/admin_news.go>`: `新闻管理, 管理端 API` → `PC端, 新闻管理`
- [ ] `<backend/internal/api/admin/admin_event.go>`: `赛事活动管理, 管理端 API` → `PC端, 赛事活动管理`
- [ ] `<backend/internal/api/admin/admin_enroll.go>`: `打卡管理, 管理端 API` → `PC端, 打卡管理`
- [ ] `<backend/internal/api/admin/admin_menu.go>`: `菜单管理, 管理端 API` → `PC端, 菜单管理`
- [ ] `<backend/internal/api/admin/admin_role.go>`: `角色管理, 管理端 API` → `PC端, 角色管理`
- [ ] `<backend/internal/api/admin/admin_dict.go>`: `字典管理, 管理端 API` → `PC端, 字典管理`
- [ ] `<backend/internal/api/admin/admin_department.go>`: `部门管理, 管理端 API` → `PC端, 部门管理`
- [ ] `<backend/internal/api/admin/admin_mgr.go>`: `管理员管理, 管理端 API` → `PC端, 管理员管理`；`在线管理员, 管理端 API` → `PC端, 在线管理员`
- [ ] `<backend/internal/api/admin/admin_setup.go>`: `系统设置, 管理端 API` → `PC端, 系统设置`
- [ ] `<backend/internal/api/admin/admin_home.go>`: `管理后台首页, 管理端 API` → `PC端, 管理后台首页`
- [ ] `<backend/internal/survey/api/admin_survey.go>`: `问卷管理, 管理端 API` → `PC端, 问卷管理`
- [ ] `<backend/internal/survey/api/admin_formkit.go>`: `表单工具, 管理端 API` → `PC端, 表单工具`
- [ ] `<backend/internal/exam/api/admin_exam.go>`: `考试管理, 管理端 API` → `PC端, 考试管理`

## 5. 重新生成 Swagger 文档
- [ ] 在 `backend/` 目录执行 `swag init`

## 6. 验证
- [ ] 检查 swagger.yaml 中 tags 和 paths 是否正确更新
