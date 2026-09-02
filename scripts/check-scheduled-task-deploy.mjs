import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const root = resolve(import.meta.dirname, '..')
const read = path => {
  const fullPath = resolve(root, path)
  if (!existsSync(fullPath)) throw new Error(`missing ${path}`)
  return readFileSync(fullPath, 'utf8')
}

const dockerfile = read('backend/Dockerfile')
for (const snippet of ['-o taskd ./cmd/taskd', 'COPY --from=builder /app/taskd .', 'tzdata']) {
  if (!dockerfile.includes(snippet)) throw new Error(`backend Dockerfile missing ${snippet}`)
}

const compose = read('backend/docker-compose.yml')
const taskdStart = compose.indexOf('\n  taskd:')
const nginxStart = compose.indexOf('\n  nginx:', taskdStart)
if (taskdStart < 0 || nginxStart < 0) throw new Error('backend compose missing standalone taskd service')
const taskd = compose.slice(taskdStart, nginxStart)
for (const snippet of ['command: ["./taskd", "--role=all"]', 'condition: service_healthy', 'restart: unless-stopped']) {
  if (!taskd.includes(snippet)) throw new Error(`taskd compose service missing ${snippet}`)
}
if (taskd.includes('ports:')) throw new Error('taskd must not publish or reserve HTTP ports')

const docs = read('docs/SCHEDULED_TASKS.md')
for (const snippet of ['taskd --role=all', 'scheduled-task.cleanup', 'workflow.notification.dispatch_due', '先执行数据库迁移', '至少一次投递']) {
  if (!docs.includes(snippet)) throw new Error(`scheduled task operations doc missing ${snippet}`)
}
