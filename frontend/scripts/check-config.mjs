import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const root = resolve(currentDir, '..')
const configPath = resolve(root, 'config/index.js')
const envExamplePath = resolve(root, '.env.example')

const configSource = readFileSync(configPath, 'utf8')
if (!existsSync(envExamplePath)) {
  throw new Error('frontend .env.example must exist')
}
const envExample = readFileSync(envExamplePath, 'utf8')

if (/192\.168\.\d+\.\d+/.test(configSource)) {
  throw new Error('frontend config must not hardcode LAN API addresses')
}

if (!configSource.includes('VITE_API_BASE_URL')) {
  throw new Error('frontend config must read VITE_API_BASE_URL')
}

if (!envExample.includes('VITE_API_BASE_URL=http://localhost:8083')) {
  throw new Error('frontend .env.example must document VITE_API_BASE_URL')
}
