import { inject, provide } from 'vue'

const performanceContextKey = Symbol('dingtalk-h5-performance')

export function providePerformanceContext(context) {
  provide(performanceContextKey, context)
}

export function usePerformanceContext() {
  const context = inject(performanceContextKey)
  if (!context) {
    throw new Error('dingtalk-h5 performance context is not provided')
  }
  return context
}
