# 后端版本化迁移说明

后端服务启动不再执行 `AutoMigrate`、SQL 迁移或种子数据。首次部署、版本升级或需要补齐基础权限数据时，请在维护窗口单独执行 `backend/init.sh`。

## 命名规则

- 文件名格式：`YYYYMMDDHHMMSS_中文或英文说明.sql`
- 一次迁移只处理一个明确主题，例如新增索引、修改字段类型、补数据。
- SQL 必须可以重复审阅，复杂数据修复需要在文件顶部写清楚前置条件和回滚方式。

## 执行策略

- `backend/init.sh` 会调用 `go run ./cmd/maintenance`。
- 维护命令先执行一次 GORM 模型建表与兼容迁移，再执行基础配置、权限菜单种子。
- 本目录中的 SQL 文件按文件名顺序执行，执行结果写入 `schema_migrations`。
- `schema_migrations.migration_version` 已存在时会跳过该文件；同版本 checksum 变化默认会报错，避免历史迁移被悄悄改写。极少数已执行且确认幂等的历史索引迁移，可在维护执行器中登记“版本 + 当前 checksum”进行一次受控校准。
- 已有 MySQL 单点部署升级时，先参考 [单点 MySQL 部署兼容升级说明](../../docs/SINGLE_NODE_MYSQL_UPGRADE.md) 完成备份、维护窗口迁移和回滚预案。
- 如果迁移涉及 `/api/v2` 响应字段、索引或权限数据，需要同步更新 [API v2 接口说明](../../docs/API_V2.md)、Swagger 文档和相关测试。

常用命令：

```bash
cd backend
bash init.sh
```

如需指定迁移目录：

```bash
WECHECKIN_MIGRATIONS_DIR=/path/to/migrations bash init.sh
```

## 配置读取规则

`backend/init.sh` 会先切换到 `backend` 目录，再执行 `go run ./cmd/maintenance`。`cmd/maintenance` 不维护独立配置，和后端主服务一样调用 `internal/config.LoadConfig`。

默认读取顺序：

1. `backend/config.yaml`
2. `backend/config/config.yaml`

执行时如果传入环境参数，会在默认配置基础上继续合并对应环境配置：

```bash
cd backend
go run ./cmd/maintenance -env prod
```

上面的命令会额外尝试读取：

```text
backend/config.prod.yaml
backend/config/config.prod.yaml
```

环境变量会覆盖配置文件中的同名配置，例如：

```bash
WECHECKIN_DATABASE_HOST=127.0.0.1 \
WECHECKIN_DATABASE_PORT=3306 \
WECHECKIN_DATABASE_USER=wecheckin \
WECHECKIN_DATABASE_PASSWORD='change-me' \
WECHECKIN_DATABASE_DBNAME=wecheckin \
bash init.sh
```

常用覆盖项包括 `WECHECKIN_DATABASE_*`、`WECHECKIN_REDIS_*`、`WECHECKIN_SERVER_*`、`WECHECKIN_TOKEN_*` 和 `WECHECKIN_CORS_*`。

注意：在宿主机直接执行 `bash backend/init.sh` 时，Docker Compose 的 `.env` 文件不会自动加载。如需复用 `.env`，请先手动导入环境变量，或在 Compose 网络内使用专门的维护容器执行迁移。

## 新增迁移检查清单

- [ ] 是否可以在生产数据量下执行。
- [ ] 是否需要先加可空字段，再分批回填，最后收紧约束。
- [ ] 是否会锁表或影响高频查询。
- [ ] 是否补充了对应的模型、索引或查询测试。
- [ ] 是否需要同步种子菜单、权限码或后台权限声明。
- [ ] 是否需要在部署文档中补充执行步骤或回滚说明。
