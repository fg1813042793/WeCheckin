import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const root = resolve(import.meta.dirname, '..')

function read(path) {
  const fullPath = resolve(root, path)
  if (!existsSync(fullPath)) {
    throw new Error(`missing required deployment file: ${path}`)
  }
  return readFileSync(fullPath, 'utf8')
}

function requireSnippet(source, snippet, label) {
  if (!source.includes(snippet)) {
    throw new Error(`${label} missing required snippet: ${snippet}`)
  }
}

function forbidSnippet(source, snippet, label) {
  if (source.includes(snippet)) {
    throw new Error(`${label} must not contain obsolete snippet: ${snippet}`)
  }
}

const envExample = read('backend/.env.example')
for (const snippet of [
  'MYSQL_ROOT_PASSWORD=',
  'WECHECKIN_DATABASE_PASSWORD=',
  'WECHECKIN_REDIS_PASSWORD=',
  'WECHECKIN_LOG_MAX_SIZE=',
  'WECHECKIN_LOG_MAX_FILE=',
  'WECHECKIN_SERVER_PORT=8083',
  'WECHECKIN_CORS_ALLOW_ORIGINS='
]) {
  requireSnippet(envExample, snippet, 'backend/.env.example')
}

const compose = read('backend/docker-compose.yml')
forbidSnippet(compose, 'version:', 'backend/docker-compose.yml')
forbidSnippet(compose, './init.sql', 'backend/docker-compose.yml')
for (const snippet of [
  'env_file:',
  'path: .env',
  'mysqladmin ping',
  'redis-cli',
  'condition: service_healthy',
  'x-logging:',
  'logging: *default-logging',
  '"8083:8083"',
  'http://127.0.0.1:8083/health',
  'WECHECKIN_REDIS_PASSWORD=${WECHECKIN_REDIS_PASSWORD'
]) {
  requireSnippet(compose, snippet, 'backend/docker-compose.yml')
}
forbidSnippet(compose, 'WECHECKIN_AUTO_MIGRATE', 'backend/docker-compose.yml')

const backendOnlyEnvExample = read('backend/.env.backend.example')
for (const snippet of [
  'WECHECKIN_DATABASE_HOST=',
  'WECHECKIN_REDIS_HOST=',
  'WECHECKIN_BACKEND_HTTP_PORT=8083',
  'WECHECKIN_CORS_ALLOW_ORIGINS='
]) {
  requireSnippet(backendOnlyEnvExample, snippet, 'backend/.env.backend.example')
}

const backendOnlyCompose = read('backend/docker-compose.backend.yml')
forbidSnippet(backendOnlyCompose, 'version:', 'backend/docker-compose.backend.yml')
forbidSnippet(backendOnlyCompose, 'condition: service_healthy', 'backend/docker-compose.backend.yml')
for (const snippet of [
  'Backend-only Docker Compose configuration',
  'services:',
  'backend:',
  'host.docker.internal:host-gateway',
  '${WECHECKIN_BACKEND_HTTP_PORT:-8083}:${WECHECKIN_SERVER_PORT:-8083}',
  'http://127.0.0.1:${WECHECKIN_SERVER_PORT:-8083}/health',
  'logging: *default-logging'
]) {
  requireSnippet(backendOnlyCompose, snippet, 'backend/docker-compose.backend.yml')
}
for (const forbidden of [
  'mysql:',
  'redis:',
  'nginx:'
]) {
  forbidSnippet(backendOnlyCompose, forbidden, 'backend/docker-compose.backend.yml')
}
forbidSnippet(backendOnlyCompose, 'WECHECKIN_AUTO_MIGRATE', 'backend/docker-compose.backend.yml')

const dingtalkH5EnvExample = read('dingtalk-h5/.env.docker.example')
for (const snippet of [
  'DINGTALK_H5_HTTP_PORT=8086',
  'VITE_API_BASE_URL=',
  'VITE_DINGTALK_CORP_ID=',
  'NGINX_API_PROXY_TARGET=http://host.docker.internal:8083'
]) {
  requireSnippet(dingtalkH5EnvExample, snippet, 'dingtalk-h5/.env.docker.example')
}

const dingtalkH5Compose = read('dingtalk-h5/docker-compose.h5.yml')
forbidSnippet(dingtalkH5Compose, 'version:', 'dingtalk-h5/docker-compose.h5.yml')
for (const snippet of [
  'DingTalk H5 standalone Docker Compose configuration',
  'dingtalk-h5:',
  'logging: *default-logging',
  '${DINGTALK_H5_HTTP_PORT:-8086}:80',
  'host.docker.internal:host-gateway',
  'NGINX_API_PROXY_TARGET'
]) {
  requireSnippet(dingtalkH5Compose, snippet, 'dingtalk-h5/docker-compose.h5.yml')
}
for (const forbidden of [
  'mysql:',
  'redis:',
  'backend:'
]) {
  forbidSnippet(dingtalkH5Compose, forbidden, 'dingtalk-h5/docker-compose.h5.yml')
}

const dingtalkH5Dockerfile = read('dingtalk-h5/Dockerfile')
for (const snippet of [
  'FROM node:20-alpine AS builder',
  'npm run build:h5',
  'FROM nginx:1.27-alpine',
  'dist/build/h5'
]) {
  requireSnippet(dingtalkH5Dockerfile, snippet, 'dingtalk-h5/Dockerfile')
}

const dingtalkH5Nginx = read('dingtalk-h5/docker/nginx.conf.template')
for (const snippet of [
  'try_files $uri $uri/ /index.html',
  'location /api/v2/',
  'proxy_pass ${NGINX_API_PROXY_TARGET}',
  'location /uploads/',
  'location = /health'
]) {
  requireSnippet(dingtalkH5Nginx, snippet, 'dingtalk-h5/docker/nginx.conf.template')
}

const initScript = read('backend/init.sh')
for (const snippet of [
  'go run ./cmd/maintenance',
  'schema_migrations',
  '-migrations'
]) {
  requireSnippet(initScript, snippet, 'backend/init.sh')
}

const backup = read('backend/scripts/docker-backup.sh')
for (const snippet of [
  'set -euo pipefail',
  'docker compose exec -T mysql',
  'mysqldump',
  'backups'
]) {
  requireSnippet(backup, snippet, 'backend/scripts/docker-backup.sh')
}

const restore = read('backend/scripts/docker-restore.sh')
for (const snippet of [
  'set -euo pipefail',
  'TYPE RESTORE TO CONTINUE',
  'docker compose exec -T mysql',
  'mysql -u"$MYSQL_USER"'
]) {
  requireSnippet(restore, snippet, 'backend/scripts/docker-restore.sh')
}

const nginx = read('backend/nginx.conf')
for (const snippet of [
  'request_id=$request_id',
  'add_header X-Request-ID $request_id always',
  'proxy_set_header X-Request-ID $request_id',
  'location /api/',
  'proxy_pass http://backend:8083;',
  'location = /nginx_status',
  'stub_status'
]) {
  requireSnippet(nginx, snippet, 'backend/nginx.conf')
}
forbidSnippet(nginx, 'proxy_pass http://backend:8083/;', 'backend/nginx.conf')

const logs = read('backend/scripts/docker-logs.sh')
for (const snippet of [
  'set -euo pipefail',
  'docker compose logs',
  '--tail',
  'SERVICE="${1:-backend}"'
]) {
  requireSnippet(logs, snippet, 'backend/scripts/docker-logs.sh')
}

const deploymentDoc = read('docs/DEPLOYMENT_TROUBLESHOOTING.md')
for (const snippet of [
  'backend/.env.example',
  'backend/docker-compose.backend.yml',
  '.env.backend.example',
  'dingtalk-h5/.env.docker.example',
  'docker-compose.h5.yml',
  'docker-backup.sh',
  'docker-restore.sh',
  'docker-logs.sh',
  'X-Request-ID',
  '日志轮转',
  'condition: service_healthy'
]) {
  requireSnippet(deploymentDoc, snippet, 'docs/DEPLOYMENT_TROUBLESHOOTING.md')
}
