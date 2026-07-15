# 中文部署和排障文档 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 提供一份可按步骤执行的中文部署和排障文档，覆盖后端、管理后台、uni-app 客户端、环境变量、验证命令和常见问题。

**Architecture:** 新增独立文档 `docs/DEPLOYMENT_TROUBLESHOOTING.md` 承载完整流程，README 只保留入口链接和简短部署提示。文档以当前代码实际配置为准，明确后端默认端口、环境变量覆盖、构建命令、静态资源部署和 Docker 示例注意事项。

**Tech Stack:** Go 1.24、Hertz、MySQL、Redis、Vue 3、Vite、uni-app、Nginx、Docker 示例。

---

## File Structure

- Create: `docs/DEPLOYMENT_TROUBLESHOOTING.md`
  - 中文部署步骤、配置清单、上线检查、常见问题和排障命令。
- Modify: `README.md`
  - 在文档列表和部署章节中链接新的中文部署排障文档。
- Modify: `docs/superpowers/plans/2026-07-14-full-stack-optimization-roadmap.md`
  - 同步 P2/P3 执行进度。

---

## Tasks

- [x] 核对 README、后端配置、Docker 示例、前后端 package 脚本和检查脚本，确认文档依据。
- [x] 新增 `docs/DEPLOYMENT_TROUBLESHOOTING.md`。
- [x] 更新 README 文档入口和部署章节。
- [x] 更新总路线图 P2/P3 进度。
- [x] 运行文档关键字检查。
- [x] 运行 `bash scripts/check.sh`。

## 验证记录

- 文档关键字检查：README 和 `docs/DEPLOYMENT_TROUBLESHOOTING.md` 已覆盖部署入口、环境变量、Nginx、Docker、管理后台、uni-app、排障和构建检查关键词。
- `bash scripts/check.sh`：通过，覆盖后端测试、uni-app 静态检查、管理后台 request/navigation 检查。
