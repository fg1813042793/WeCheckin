# 后端全量优化续作计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans and superpowers:test-driven-development. 本计划承接 `2026-07-16-backend-p0-p2-optimization.md`，用于继续完成剩余 P1-P3 优化。

**目标：** 在不推翻现有接口和业务行为的前提下，继续收敛后端响应契约、上下文链路、事务一致性、迁移可观测性、权限审计、服务测试与结构约束。

**执行原则：**
- 先加结构测试或单元测试，再修改实现。
- 高频接口优先 DTO 化，低频工具接口保留动态结构但加白名单说明。
- DB/Redis 调用优先迁移业务链路和中间件链路；长期仓储化通过结构测试持续压降。
- 多表写入必须进入事务，失败时不能留下半完成数据。
- 自动迁移继续保守运行，但必须有日志和测试覆盖，后续再迁到版本化 migration。

---

## Task 1：扩大高频接口 DTO 约束

- [x] 后台角色、管理员、新闻、活动、报名列表响应使用 DTO。
- [x] 问卷题库、通知、渠道、统计响应补 DTO。
- [x] 客户端问卷、考试、新闻、通行证核心响应补 DTO。
- [x] 结构测试覆盖新的高频文件，防止重新出现裸 `response.JSON(c, map[string]interface{})`。
- [x] FormKit 工具/报表接口保留动态响应，但限定在专用文件中。

## Task 2：服务层返回结构收敛

- [x] `role` 列表返回结构化 DTO。
- [x] `adminmgr` 列表和详情返回结构化 DTO。
- [x] `adminauth` 登录返回结构化 DTO。
- [x] `passport` 登录、注册、个人详情返回结构化 DTO。
- [x] `news` 列表和分类返回结构化 DTO。
- [x] `dict` 类型列表返回结构化 DTO。
- [x] 仍需动态 JSON 的问卷/表单统计接口保留动态结构，但用命名 DTO 包裹顶层字段。

## Task 3：上下文与事务一致性

- [x] `role`、`adminmgr`、`adminauth`、`passport`、`dict`、`news` 迁移到 `database.WithContext`。
- [x] 客户端问卷/考试核心 handler 查询迁移到 `database.WithContext`。
- [x] 角色新增/编辑/删除、管理员新增/编辑/删除改为事务。
- [x] 权限缓存失效与角色/菜单写入保持同一业务路径可观测。
- [x] 支持部门关联保存提供 Context/Tx helper。

## Task 4：迁移可靠性

- [x] 启动迁移的 `ALTER TABLE` 必须记录执行结果，不能静默吞错。
- [x] 增加迁移执行器测试，覆盖执行成功、执行失败、失败步骤命名。
- [x] 增加版本化 migration 的目录与中文说明，后续新增 DDL 必须进入版本文件。

## Task 5：权限缓存与审计增强

- [x] 权限拒绝审计输出结构化字段。
- [x] 权限缓存增加可测试的失效版本号，便于后续多实例 Redis 广播。
- [x] 结构测试确保后台路由权限声明继续收敛到显式白名单。

## Task 6：核心服务测试补齐

- [x] `adminauth`、`adminmgr`、`role` 覆盖结构约束或纯逻辑测试。
- [x] `news`、`online`、`poststat` 覆盖解析、构造和边界逻辑。
- [x] `media` 覆盖数据库未初始化时的 fallback。
- [x] 新增测试不依赖外部网络或真实 Redis。

## Task 7：P3 结构化演进

- [x] FormKit/报表/工具类接口保留动态能力，但通过结构测试限定到专用文件。
- [x] 增加 API DTO 与 handler 结构保护文档。
- [x] 补充中文迁移、权限、DTO 规范，方便后续继续执行。

## Verification

- [x] `GOCACHE=$PWD/.cache/go-build go test ./backend/...`
- [x] `git diff --check`
