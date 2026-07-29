import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))

const pageConfigs = [
  {
    label: '通知列表',
    path: '../pages/news/news_index.vue',
    imageClass: 'news-img-inner',
    loadMethod: 'loadData',
    hasTabs: false,
  },
  {
    label: '打卡任务',
    path: '../pages/enroll/enroll_index.vue',
    imageClass: 'card-img-inner',
    loadMethod: 'loadData',
    hasTabs: true,
  },
  {
    label: '赛事活动',
    path: '../pages/event/event_index.vue',
    imageClass: 'card-img-inner',
    loadMethod: 'loadData',
    hasTabs: true,
  },
  {
    label: '问卷列表',
    path: '../pages/survey/index.vue',
    imageClass: '',
    loadMethod: 'load',
    hasTabs: false,
    requiresLimitMerge: true,
    responseMapName: 'myRespMap',
  },
  {
    label: '考试列表',
    path: '../pages/exam/index.vue',
    imageClass: '',
    loadMethod: 'load',
    hasTabs: false,
    requiresLimitMerge: true,
    responseMapName: 'myRecordMap',
  },
]

function assertIncludes(label, source, snippet) {
  if (!source.includes(snippet)) {
    throw new Error(`${label} missing list performance guard: ${snippet}`)
  }
}

function assertNotIncludes(label, source, snippet) {
  if (source.includes(snippet)) {
    throw new Error(`${label} must not use eager or full-page list loading: ${snippet}`)
  }
}

function assertLazyImage(label, source, imageClass) {
  if (!imageClass) return
  const imageTags = source.match(/<image\b[^>]*>/g) || []
  const cardImage = imageTags.find(tag => tag.includes(`class="${imageClass}"`) || tag.includes(`class='${imageClass}'`))
  if (!cardImage) {
    throw new Error(`${label} missing card image tag: ${imageClass}`)
  }
  if (!cardImage.includes('lazy-load')) {
    throw new Error(`${label} card image must use lazy-load`)
  }
}

for (const config of pageConfigs) {
  const file = resolve(currentDir, config.path)
  const source = readFileSync(file, 'utf8')

  assertLazyImage(config.label, source, config.imageClass)
  assertIncludes(config.label, source, 'page: 1')
  assertIncludes(config.label, source, 'pageSize: 10')
  assertIncludes(config.label, source, 'hasMore: true')
  assertIncludes(config.label, source, 'loading: false')
  assertIncludes(config.label, source, 'onReachBottom()')
  assertIncludes(config.label, source, 'this.loadMore()')
  assertIncludes(config.label, source, 'if ((!this.hasMore && this.page > 1) || this.loading) return')
  assertIncludes(config.label, source, 'this.loading = true')
  assertIncludes(config.label, source, 'this.loading = false')
  assertIncludes(config.label, source, 'this.page === 1')
  assertIncludes(config.label, source, 'this.list = [...this.list, ...data]')
  assertIncludes(config.label, source, 'data.length >= this.pageSize')
  assertIncludes(config.label, source, 'if (this.hasMore && !this.loading)')
  assertIncludes(config.label, source, 'this.page++')
  assertIncludes(config.label, source, 'page: this.page')
  assertIncludes(config.label, source, 'pageSize: this.pageSize')
  assertNotIncludes(config.label, source, 'pageSize: 50')

  if (config.hasTabs) {
    assertIncludes(config.label, source, 'if (this.cur === tab) return')
  }

  if (config.requiresLimitMerge) {
    assertIncludes(config.label, source, '...this.limitsMap')
  }

  if (config.responseMapName) {
    assertIncludes(config.label, source, `if (!this.isLogged && this.page === 1)`)
    assertNotIncludes(config.label, source, `} else {\n          this.${config.responseMapName} = {}`)
  }

  for (const eagerTrigger of [
    '@input="load',
    '@input="loadData',
    '@input="handleSearch',
  ]) {
    assertNotIncludes(config.label, source, eagerTrigger)
  }
}
