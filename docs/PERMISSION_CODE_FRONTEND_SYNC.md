# 权限编码与前端同步说明

更新时间：2026-07-30

本文说明当后台或后端调整权限编码时，前端需要同步修改哪些地方。权限编码是前后端共同使用的契约，不建议只在后台“权限管理”页面直接修改内置权限编码。

## 一句话规则

内置权限编码要按“后端声明 -> 数据迁移 -> 前端常量/筛选规则 -> 重新登录验证”的顺序同步。

如果只在数据库或后台页面改了 `permission_key`，可能出现以下问题：

- 后端接口中间件仍按旧编码校验，接口返回“无权限访问”。
- 钉钉 H5 前端仍按旧编码判断按钮和接口能力，按钮不显示或显示后无法调用。
- 角色/用户表单按前缀分组，改坏前缀后权限不会出现在正确的树里。
- 内置权限重新同步时，旧编码可能被代码里的声明再次写回。

## 权限编码来源

| 场景 | 后端声明位置 | 前端是否通常要改 |
| --- | --- | --- |
| 后台菜单/按钮 | `backend/internal/app/support/adminmenuperm/declarations.go` | 通常不用逐个改，后台菜单树来自后端；如果页面按钮还写死了 `hasPerm`，需要同步。 |
| 后台 API | `backend/internal/app/support/adminrouteperm/catalog.go` | 通常不用逐个改；接口权限树来自后端。 |
| 客户端菜单 | `backend/internal/app/support/appmenuperm/catalog.go` | 现阶段管理后台配置页通常不用逐个改，除非前端页面自己写死了菜单 key。 |
| 客户端 API | `backend/internal/app/support/appapiperm/catalog.go` | 管理后台配置页通常不用逐个改，除非调用端写死了权限 key。 |
| 钉钉 H5 菜单/按钮 | `backend/internal/app/support/appmenuperm/catalog.go` | 菜单主要来自 `bootstrap` 返回；按钮 key 在 H5 页面有硬编码，需要同步。 |
| 钉钉 H5 API | `backend/internal/app/support/appapiperm/catalog.go` | 需要同步 H5 页面里所有 `hasApiPermission` 和动作映射。 |

数据库中的 `permissions.permission_key` 是权限定义编码，`permission_grants.grant_permission_key` 是角色或用户实际授权编码。两者必须保持一致。

菜单展示名和图标也来自 `permissions`：

- `permission_name`：后台权限树和钉钉 H5 菜单显示名称。
- `permission_icon`：菜单图标。后台菜单使用 Element Plus 图标名；钉钉 H5 菜单使用 H5 内置图标键，例如 `dashboard`、`performance`、`mine`、`history`、`manager`、`hrbp`、`summary`、`org`、`template`、`account`。
- 修改钉钉 H5 菜单图标通常只需要在后台“权限管理 / 钉钉 H5”中调整图标，不需要改权限编码。

## 标准修改流程

1. 先确定旧编码和新编码，例如：
   - 旧：`dingtalk_h5:api:template:save`
   - 新：`dingtalk_h5:api:template:update`
2. 修改后端内置声明。
   - 菜单/按钮改 `appmenuperm` 或 `adminmenuperm`。
   - API 改 `appapiperm` 或 `adminrouteperm`。
   - API 权限还要确认 `HTTP Method + Path` 对应的新 `PermissionKey` 正确，否则中间件仍会按旧 key 或找不到 key。
3. 增加迁移脚本同步历史数据。
   - 更新 `permissions.permission_key`。
   - 更新子权限的 `permissions.permission_parent_key`。
   - 更新 `permission_grants.grant_permission_key`。
4. 修改前端硬编码权限 key。
   - 管理后台角色/用户权限表单主要读取后端权限树，通常不需要逐个改 key。
   - 钉钉 H5 页面存在按钮和 API 权限常量，需要同步。
5. 发布后清缓存并重新登录。
   - 权限树和钉钉 H5 菜单权限有短 TTL 缓存，权限保存会主动失效，但用户端仍建议重新登录或重新拉取 `bootstrap`。
   - 钉钉 H5 依赖 `bootstrap/login` 返回的 `menus`、`buttonPermissionKeys`、`apiPermissionKeys` 和 `permissionVersion`。

## 管理后台前端怎么改

角色编辑和用户编辑中的“应用权限”配置主要按后端返回的树渲染，不会写死每个具体权限编码。

需要注意的是前缀筛选：

| 前缀 | 用途 |
| --- | --- |
| `client:menu:` | 客户端菜单权限。 |
| `client:api:` | 客户端接口权限。 |
| `dingtalk_h5:menu:` | 钉钉 H5 菜单权限。 |
| `dingtalk_h5:button:` | 钉钉 H5 页面按钮权限。 |
| `dingtalk_h5:api:` | 钉钉 H5 接口权限。 |

如果只是把 `dingtalk_h5:api:template:save` 改成 `dingtalk_h5:api:template:update`，且仍保持 `dingtalk_h5:api:` 前缀，管理后台配置页通常不需要改。

如果连前缀、平台或权限类型一起改了，例如把 `dingtalk_h5:api:` 改成 `h5:api:`，则必须同步修改：

- `admin/src/views/role/index.vue` 中应用权限树的 key 过滤逻辑。
- `admin/src/views/user/index.vue` 中用户额外授权/禁止权限的 key 过滤逻辑。
- 后端权限目录中的 `permission_platform`、`permission_type` 和相关树构造逻辑。

除非是整体重命名平台，否则不要修改这些标准前缀。

## 钉钉 H5 前端怎么改

钉钉 H5 页面会在登录或启动后读取后端返回的权限：

- `menus`：当前用户可见菜单。
- `buttonPermissionKeys`：当前用户可用按钮。
- `apiPermissionKeys`：当前用户可访问接口。
- `permissionVersion`：权限版本，用于感知权限变化。

菜单显示主要使用后端返回的 `menus`，如果只改菜单权限 key、不改菜单 `path`，H5 页面通常不用改路由逻辑。

菜单图标也跟随 `menus[].icon` 返回。后端优先读取 `permissions.permission_icon`，没有配置时使用 `appmenuperm.DingTalkH5MenuDeclarations()` 里的默认图标。H5 前端的实际 SVG 图标集合在 `dingtalk-h5/components/performance/AppShell.vue` 的 `navIconMap` 中；新增图标键时，需要同时更新后台可选项和这个映射。

按钮和接口判断存在硬编码，修改权限编码时必须同步检查：

- `dingtalk-h5/pages/index/index.vue` 中的流程动作 API 映射：`reviewActionApiPermissions`。
- `dingtalk-h5/pages/index/index.vue` 中的流程动作按钮映射：`reviewActionButtonPermissions`。
- `dingtalk-h5/pages/index/index.vue` 中所有 `hasApiPermission('dingtalk_h5:api:*')`。
- `dingtalk-h5/pages/index/index.vue` 中所有 `hasButtonPermission('dingtalk_h5:button:*')`。

示例：如果把模板保存接口权限从 `dingtalk_h5:api:template:save` 改为 `dingtalk_h5:api:template:update`，需要同步：

```js
function canEditTemplate() {
  return hasButtonPermission('dingtalk_h5:button:template:edit') &&
    hasApiPermission('dingtalk_h5:api:template:update')
}
```

同时后端 `appapiperm` 中 `/api/v2/dingtalk/h5/template` 对应的 `PermissionKey` 也要改为 `dingtalk_h5:api:template:update`，并迁移已有授权。

## 后端接口为什么也要改

接口权限不是前端自己决定的。钉钉 H5 请求会经过 `ApiPerm`：

```text
登录态 -> 按 Method + Path 查找接口权限编码 -> permission_grants 判断用户/角色授权 -> 放行或拒绝
```

因此前端显示按钮只代表体验层，真正是否能访问接口取决于后端路由声明里的权限编码和数据库授权是否一致。

如果前端改成新编码，但后端路由仍声明旧编码，会出现：

- 页面看起来有权限。
- 调接口时仍返回 `{"code":1,"msg":"无权限访问"}`。

如果后端改成新编码，但前端仍检查旧编码，会出现：

- 接口权限已经配置成功。
- 页面按钮不显示，或前端提前提示无权限。

## 后台页面直接改权限编码时的影响

后台“权限管理”页面编辑权限编码时，当前服务会同步更新：

- `permissions.permission_key`
- 直接子权限的 `permissions.permission_parent_key`
- `permission_grants.grant_permission_key`

这能保证数据库里的授权关系不断开。

但内置权限仍以代码声明为准。对于后台菜单、后台接口、客户端菜单、钉钉 H5 菜单/按钮/API 这些内置权限，推荐把后台页面编辑视为临时调试能力；正式改名必须同步代码声明和迁移脚本。

## 验证清单

修改权限编码后，手动验证以下几点：

1. 后台权限管理中只看到新编码，不再有旧编码。
2. 角色编辑和用户编辑表单能正常回显已授权的新编码。
3. 钉钉 H5 登录或 `bootstrap` 返回的新权限数组包含新编码。
4. 钉钉 H5 页面按钮显示符合预期。
5. 有授权用户访问对应接口成功。
6. 无授权用户访问对应接口返回“无权限访问”。
7. 数据库 `permission_grants.grant_permission_key` 不再残留旧编码。

## 建议

后续新增或重命名权限时，优先保持编码稳定。能只改权限名称、排序、父级或展示文案时，不要改 `permission_key`。

确实要重命名时，提交内容至少包含：

- 后端权限声明修改。
- 数据迁移脚本。
- 钉钉 H5 或其他调用端硬编码权限 key 修改。
- 本文验证清单的结果。
