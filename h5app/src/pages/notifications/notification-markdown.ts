import MarkdownIt from 'markdown-it'

function createNotificationMarkdown(html: boolean) {
  const markdown = new MarkdownIt({
    html,
    linkify: true,
    breaks: true,
    typographer: false,
  })
  const defaultLinkOpen = markdown.renderer.rules.link_open
  markdown.renderer.rules.link_open = (tokens, index, options, env, self) => {
    tokens[index].attrSet('target', '_blank')
    tokens[index].attrSet('rel', 'noopener noreferrer')
    return defaultLinkOpen
      ? defaultLinkOpen(tokens, index, options, env, self)
      : self.renderToken(tokens, index, options)
  }
  return markdown
}

const markdownWithHtml = createNotificationMarkdown(true)
const markdownWithoutHtml = createNotificationMarkdown(false)

export function renderNotificationMarkdownSource(content: string) {
  return markdownWithHtml.render(String(content || '').trim() || '暂无内容')
}

export function renderNotificationMarkdownWithoutHtml(content: string) {
  return markdownWithoutHtml.render(String(content || '').trim() || '暂无内容')
}
