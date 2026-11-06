import DOMPurify from 'dompurify'

const FORBID_TAGS = ['script', 'style', 'iframe', 'object', 'embed', 'form', 'input', 'button', 'meta', 'link']
const FORBID_ATTR = [
  'srcdoc',
  'onerror',
  'onload',
  'onclick',
  'onfocus',
  'onmouseover',
  'onmouseenter',
  'onmouseleave',
  'onanimationstart',
]

export function sanitizeRichText(value: string) {
  return DOMPurify.sanitize(String(value || ''), {
    USE_PROFILES: { html: true },
    FORBID_TAGS,
    FORBID_ATTR,
    ADD_ATTR: ['target', 'rel', 'data-language', 'data-row'],
    ADD_DATA_URI_TAGS: ['img'],
    ALLOW_DATA_ATTR: true,
  })
}

export function serializeRichText(value: string) {
  const template = document.createElement('template')
  template.innerHTML = String(value || '')
  template.content.querySelectorAll('.ql-ui').forEach(node => node.remove())
  return template.innerHTML
}

