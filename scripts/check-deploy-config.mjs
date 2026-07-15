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
  'WECHECKIN_AUTO_MIGRATE=',
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
  'location = /nginx_status',
  'stub_status'
]) {
  requireSnippet(nginx, snippet, 'backend/nginx.conf')
}

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
  'docker-backup.sh',
  'docker-restore.sh',
  'docker-logs.sh',
  'X-Request-ID',
  '日志轮转',
  'condition: service_healthy'
]) {
  requireSnippet(deploymentDoc, snippet, 'docs/DEPLOYMENT_TROUBLESHOOTING.md')
}
