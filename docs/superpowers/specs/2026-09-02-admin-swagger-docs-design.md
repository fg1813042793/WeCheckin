# Admin 接口文档菜单设计

## 1. 目标

在 WeCheckin Admin 后台增加“接口文档”菜单，管理员可在现有后台页签内查看后端 Swagger UI，并可刷新或在新窗口打开文档。

## 2. 菜单与权限

- 新增顶级菜单权限 `admin:menu:swagger`，菜单名称为“接口文档”，路径为 `/swagger-docs`，图标为 `Document`，兼容权限标识为 `swagger:view`。
- 菜单排序位于当前“定时任务”之后，不新增只有一个子项的目录。
- 保留现有权限模型：保留超级管理员自动读取全部后台菜单，普通角色通过权限管理显式授权。
- 本次权限只控制 Admin 菜单可见性，不改变后端 `/swagger` 当前公开访问策略，也不新增伪装成业务 API 的权限路由。
- 新旧数据库都通过版本化 SQL 新增菜单定义；代码权限目录同步增加相同声明。

## 3. Admin 页面

- 新增 `/swagger-docs` 路由和独立 Vue 页面。
- 页面使用同源 `/swagger/index.html` 作为 iframe 地址，避免在前端硬编码后端主机和端口。
- 加载前请求 `/swagger/doc.json` 检查文档是否可用；失败时显示错误状态和重试操作。
- 工具栏提供刷新和新窗口打开两个图标按钮，并为图标提供 title 与 aria-label。
- iframe 使用稳定的全高容器，跟随 Admin 工作区尺寸变化，不在内容外增加嵌套卡片。

## 4. 代理与部署

- Vite 开发服务器增加 `/swagger` 到本地后端的代理，使开发环境 iframe 与 Admin 保持同源。
- `admin/nginx/default.conf.template` 与根 `nginx/default.conf.template` 同步代理 `/swagger` 到 `NGINX_API_UPSTREAM`。
- Swagger UI 内部访问的 `/swagger/doc.json` 和静态资源继续由后端现有 `RegisterSwagger` 路由提供。

## 5. 错误处理与验证

- 文档探测失败时不渲染空白 iframe，显示加载失败状态；用户可重试或在新窗口打开后端文档。
- 验证覆盖后台菜单声明、历史数据库迁移、Admin 路由与页面结构、Vite/Nginx 代理配置、TypeScript 构建和后端全量测试。
- 按用户此前要求不执行浏览器自动化，实际显示效果由用户启动环境后验证。
