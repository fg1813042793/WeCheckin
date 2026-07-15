import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const source = readFileSync(resolve(currentDir, '../src/utils/request.ts'), 'utf8')

const requiredSnippets = [
  "import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'",
  'export interface ApiResponse<T = any>',
  'export type ApiRequest',
  'const request = axiosInstance as ApiRequest',
  "import { clearPerms } from './permission'",
  'const LOGIN_EXPIRED_MESSAGES = new Set',
  'function clearAdminSession()',
  'clearPerms()',
  'function redirectToLogin()',
  'function encodeFormBody',
  'data instanceof FormData',
  'headers.delete',
]

for (const snippet of requiredSnippets) {
  if (!source.includes(snippet)) {
    throw new Error(`admin request layer missing: ${snippet}`)
  }
}

const forbiddenResponseCasts = [
  'adminApi.surveyTemplatePresetsGet() as any',
  "request.get('/home/setup_get', { params: { key: 'BUILTIN_TEMPLATE_PRESETS' } }) as any"
]

const surveyDesignerSource = readFileSync(resolve(currentDir, '../src/views/survey/SurveyDesigner.vue'), 'utf8')
for (const snippet of forbiddenResponseCasts) {
  if (surveyDesignerSource.includes(snippet)) {
    throw new Error(`admin response handling must not cast typed API response to any: ${snippet}`)
  }
}
