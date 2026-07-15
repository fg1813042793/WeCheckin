# P0-P2 加固和优化 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 完成 P0 到 P2 的后续优化：密码散列升级、部署配置统一、健康检查、Docker/Nginx 可用化、管理后台类型和构建体积优化、移动端 formkit 逻辑补齐，以及项目级检查增强。

**Architecture:** 后端安全改动采用兼容式升级：新增密码工具包，登录时兼容旧 MD5 并在验证成功后写回 bcrypt，新密码写入路径全部使用 bcrypt。部署相关改动统一以 `8083` 为后端端口，补 `/health` 和 `/ready`，修正 Docker/Nginx 示例和文档；前端与管理后台优化只做边界清晰、可验证的小步收口。

**Tech Stack:** Go 1.24、Hertz、GORM、MySQL、Redis、bcrypt、Vue 3、Vite、uni-app、Node 静态检查、Nginx、Docker Compose。

---

## File Structure

- Create: `backend/pkg/passwordutil/passwordutil.go`
  - 负责 bcrypt 哈希、旧 MD5 兼容校验、是否需要升级判断。
- Create: `backend/pkg/passwordutil/passwordutil_test.go`
  - 覆盖 bcrypt 哈希、旧 MD5 兼容、错误密码拒绝。
- Modify: `backend/internal/app/service/admin.go`
  - 管理员登录、新增、编辑密码、修改密码改用 `passwordutil`，登录成功后升级旧哈希。
- Modify: `backend/internal/app/service/passport.go`
  - 用户注册默认密码、密码登录改用 `passwordutil`，登录成功后升级旧哈希。
- Modify: `backend/internal/app/service/admin_password_safety_test.go`
  - 静态防回归，确保登录不再把密码哈希写进 SQL 条件。
- Modify: `backend/internal/config/config.go`
  - 默认端口改为 `8083`。
- Modify: `backend/cmd/main.go`
  - Swagger 注解 host 改为 `localhost:8083`。
- Modify: `backend/cmd/routes.go`
  - 增加 `/health` 和 `/ready`。
- Modify: `backend/cmd/main_safety_test.go`
  - 增加端口/健康检查/启动入口防回归。
- Modify: `backend/Dockerfile`
  - 暴露 `8083`，复制配置目录，包级构建入口。
- Modify: `backend/docker-compose.yml`
  - 后端端口、环境变量、Redis 密码和 Nginx 代理路径对齐当前后端。
- Modify: `backend/nginx.conf`
  - server 块放入 http，代理到 `backend:8083`，保留上传和健康检查。
- Modify: `backend/internal/app/service/database.go`
  - 启动迁移支持环境变量开关，默认保持现状，生产可关闭 AutoMigrate。
- Modify: `backend/internal/app/service/database_safety_test.go`
  - 增加迁移开关防回归。
- Modify: `backend/internal/app/service/home.go`
  - 默认静态域名改为 `http://localhost:8083`。
- Modify: `backend/docs/swagger/*`
  - 同步 Swagger host。
- Modify: `admin/src/utils/request.ts`
  - 提供类型化 `ApiResponse` 和 request 实例导出，减少调用端 `as any`。
- Modify: `admin/vite.config.ts`
  - 配置 manualChunks 和 chunk warning 阈值。
- Modify: `frontend/components/formkit/FormRender.vue`
  - 接入已有 `frontend/utils/logicEngine.js`，支持 show/hide/required 基础逻辑。
- Modify: `frontend/scripts/*.mjs` / `admin/scripts/*.mjs` / `scripts/check.sh`
  - 增加端口、Docker、健康检查、formkit 逻辑和构建配置静态检查。
- Modify: `README.md` and `docs/DEPLOYMENT_TROUBLESHOOTING.md`
  - 更新已完成的部署、健康检查、Docker 和密码升级说明。
- Modify: `docs/superpowers/plans/2026-07-14-full-stack-optimization-roadmap.md`
  - 同步 P0-P2 完成进度。

---

## Tasks

- [x] P0：新增密码工具包红灯测试。
- [x] P0：实现 bcrypt + 旧 MD5 兼容校验。
- [x] P0：改造用户和管理员密码写入、登录验证、旧哈希升级路径。
- [x] P0：统一后端默认端口、Swagger、静态域名、Docker、compose、Nginx 到 `8083`。
- [x] P1：新增 `/health` 和 `/ready`，并加入项目检查。
- [x] P1：为启动 AutoMigrate 增加可关闭开关和安全测试。
- [x] P1：管理后台 request 返回类型收口，并移除当前已知响应处 `as any`。
- [x] P1：管理后台 Vite manualChunks 和构建体积 warning 收口。
- [x] P2：移动端 FormRender 接入基础逻辑规则。
- [x] P2：移动端 FormRender 接入计算字段即时回填。
- [x] P2：扩展项目级静态检查，覆盖 Docker/Nginx/端口/健康检查/formkit 逻辑。
- [x] 文档：更新 README、部署排障文档和路线图。
- [x] 验证：运行专项测试、`bash scripts/check.sh`、`CHECK_ADMIN_BUILD=1 bash scripts/check.sh`。
