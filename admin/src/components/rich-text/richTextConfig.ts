import Quill from 'quill'
import type { Attributor } from 'parchment'
import hljs from 'highlight.js/lib/core'
import bash from 'highlight.js/lib/languages/bash'
import c from 'highlight.js/lib/languages/c'
import cpp from 'highlight.js/lib/languages/cpp'
import csharp from 'highlight.js/lib/languages/csharp'
import go from 'highlight.js/lib/languages/go'
import java from 'highlight.js/lib/languages/java'
import javascript from 'highlight.js/lib/languages/javascript'
import json from 'highlight.js/lib/languages/json'
import python from 'highlight.js/lib/languages/python'
import sql from 'highlight.js/lib/languages/sql'
import typescript from 'highlight.js/lib/languages/typescript'
import xml from 'highlight.js/lib/languages/xml'

export interface RichTextSelectOption {
  label: string
  value: string
}

export const RICH_TEXT_FONT_OPTIONS: RichTextSelectOption[] = [
  { label: '默认字体', value: '' },
  { label: '微软雅黑', value: 'Microsoft YaHei' },
  { label: '宋体', value: 'SimSun' },
  { label: '黑体', value: 'SimHei' },
  { label: '楷体', value: 'KaiTi' },
  { label: 'Arial', value: 'Arial' },
  { label: 'Times New Roman', value: 'Times New Roman' },
  { label: '等宽字体', value: 'monospace' },
]

export const RICH_TEXT_SIZE_OPTIONS: RichTextSelectOption[] = [
  { label: '默认字号', value: '' },
  { label: '12px', value: '12px' },
  { label: '14px', value: '14px' },
  { label: '16px', value: '16px' },
  { label: '18px', value: '18px' },
  { label: '24px', value: '24px' },
  { label: '32px', value: '32px' },
]

export const RICH_TEXT_HEADER_OPTIONS: RichTextSelectOption[] = [
  { label: '正文', value: '' },
  { label: '标题 1', value: '1' },
  { label: '标题 2', value: '2' },
  { label: '标题 3', value: '3' },
]

export const RICH_TEXT_LANGUAGES: RichTextSelectOption[] = [
  { label: '纯文本', value: 'plain' },
  { label: 'JavaScript', value: 'javascript' },
  { label: 'TypeScript', value: 'typescript' },
  { label: 'Java', value: 'java' },
  { label: 'Python', value: 'python' },
  { label: 'C', value: 'c' },
  { label: 'C++', value: 'cpp' },
  { label: 'C#', value: 'csharp' },
  { label: 'Go', value: 'go' },
  { label: 'JSON', value: 'json' },
  { label: 'HTML / XML', value: 'xml' },
  { label: 'Shell', value: 'bash' },
  { label: 'SQL', value: 'sql' },
]

const languageModules = {
  bash,
  c,
  cpp,
  csharp,
  go,
  java,
  javascript,
  json,
  python,
  sql,
  typescript,
  xml,
}

for (const [name, language] of Object.entries(languageModules)) {
  hljs.registerLanguage(name, language)
}

let formatsRegistered = false

export function ensureRichTextFormatsRegistered() {
  if (formatsRegistered)
    return

  const fontStyle = Quill.import('attributors/style/font') as Attributor
  const sizeStyle = Quill.import('attributors/style/size') as Attributor
  fontStyle.whitelist = RICH_TEXT_FONT_OPTIONS.map(item => item.value).filter(Boolean)
  sizeStyle.whitelist = RICH_TEXT_SIZE_OPTIONS.map(item => item.value).filter(Boolean)
  Quill.register(fontStyle, true)
  Quill.register(sizeStyle, true)
  formatsRegistered = true
}

export function createRichTextModules() {
  ensureRichTextFormatsRegistered()
  return {
    toolbar: false,
    table: true,
    history: {
      delay: 800,
      maxStack: 100,
      userOnly: true,
    },
    syntax: {
      hljs,
      languages: RICH_TEXT_LANGUAGES.map(item => ({ key: item.value, label: item.label })),
      interval: 300,
    },
    imageResize: {},
  }
}

