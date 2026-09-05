import type { UserFeedbackMessage, UserFeedbackStatus } from '@/api/user-feedback'

export type FeedbackDraftMode = 'create' | 'supplement'
export type FeedbackDraftImageError = 'too_large' | 'unsupported_type'
export type FeedbackDraftValidationError
  = ''
    | 'content_required'
    | 'content_too_long'
    | 'content_or_image_required'
    | 'images_too_many'
    | 'image_invalid'

export interface FeedbackDraftImage {
  clientId: string
  filePath: string
  name?: string
  mimeType?: string
  size?: number
  error?: FeedbackDraftImageError
}

export interface FeedbackComponentLifecycle {
  isActive: () => boolean
  invalidate: () => void
}

export interface FeedbackAsyncOperationOptions<T> {
  lifecycle: FeedbackComponentLifecycle
  request: () => Promise<T>
  success: (value: T) => void | Promise<void>
  failure: (error: unknown) => void | Promise<void>
  settled: () => void
}

export interface FeedbackVersionConflictRecoveryOptions {
  notify: (message: string) => void
  refresh: () => Promise<void>
}

const MAX_CONTENT_LENGTH = 5000
const MAX_IMAGE_COUNT = 6
const MAX_IMAGE_SIZE = 10 * 1024 * 1024
const SAFE_IMAGE_MIME_TYPES = new Set(['image/jpeg', 'image/png', 'image/webp'])

export function createFeedbackComponentLifecycle(): FeedbackComponentLifecycle {
  let active = true
  return {
    isActive: () => active,
    invalidate: () => {
      active = false
    },
  }
}

export async function runFeedbackAsyncOperation<T>(options: FeedbackAsyncOperationOptions<T>) {
  try {
    const value = await options.request()
    if (!options.lifecycle.isActive())
      return 'stale' as const
    await options.success(value)
    return 'success' as const
  }
  catch (error) {
    if (!options.lifecycle.isActive())
      return 'stale' as const
    await options.failure(error)
    return 'failure' as const
  }
  finally {
    if (options.lifecycle.isActive())
      options.settled()
  }
}

export function createFeedbackRequestId() {
  const cryptoValue = globalThis.crypto?.randomUUID?.()
  const entropy = cryptoValue
    ? cryptoValue.toLowerCase().replace(/[^a-z0-9-]/g, '')
    : `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 14)}`
  return `feedback-${entropy}`.slice(0, 64)
}

export function createStableFeedbackRequestIdState(
  factory: () => string | undefined = createFeedbackRequestId,
) {
  let value = String(factory() || createFeedbackRequestId()).trim()

  return {
    current: () => value,
    rotate: () => {
      value = String(factory() || createFeedbackRequestId()).trim()
      return value
    },
  }
}

export function feedbackContentLength(content: string) {
  return Array.from(String(content || '')).length
}

export function feedbackDraftHasContent(content: string, images: readonly FeedbackDraftImage[]) {
  return Boolean(String(content || '').trim() || images.length > 0)
}

export function feedbackImageSelectionCount(selectedCount: number, maxCount: number) {
  return {
    selectedCount: Math.max(0, Math.floor(Number(selectedCount) || 0)),
    maxCount: Math.max(1, Math.floor(Number(maxCount) || MAX_IMAGE_COUNT)),
  }
}

export function validFeedbackDraftImages(images: readonly FeedbackDraftImage[]) {
  return images.filter(image => Boolean(image.filePath) && !image.error)
}

export function validateFeedbackDraft(
  mode: FeedbackDraftMode,
  content: string,
  images: readonly FeedbackDraftImage[],
): FeedbackDraftValidationError {
  if (feedbackContentLength(content) > MAX_CONTENT_LENGTH)
    return 'content_too_long'
  if (images.length > MAX_IMAGE_COUNT)
    return 'images_too_many'
  if (images.some(image => Boolean(image.error) || !image.filePath))
    return 'image_invalid'

  const hasText = Boolean(String(content || '').trim())
  const hasImage = validFeedbackDraftImages(images).length > 0
  if (mode === 'create' && !hasText)
    return 'content_required'
  if (mode === 'supplement' && !hasText && !hasImage)
    return 'content_or_image_required'
  return ''
}

function chooseResultFiles(result: unknown) {
  if (!result || typeof result !== 'object' || Array.isArray(result))
    return { files: [] as unknown[], paths: [] as unknown[] }
  const record = result as Record<string, unknown>
  return {
    files: Array.isArray(record.tempFiles) ? record.tempFiles : [],
    paths: Array.isArray(record.tempFilePaths) ? record.tempFilePaths : [],
  }
}

function chosenFileRecord(value: unknown) {
  return value && typeof value === 'object' && !Array.isArray(value)
    ? value as Record<string, unknown>
    : {}
}

export function normalizeFeedbackChosenImages(result: unknown, batchID: string): FeedbackDraftImage[] {
  const { files, paths } = chooseResultFiles(result)
  const count = Math.max(files.length, paths.length)
  const normalized: FeedbackDraftImage[] = []

  for (let index = 0; index < count; index += 1) {
    const file = chosenFileRecord(files[index])
    const filePath = String(file.path || file.tempFilePath || paths[index] || '').trim()
    if (!filePath)
      continue
    const mimeType = String(file.type || file.mimeType || '').trim().toLowerCase().split(';')[0]
    const size = Math.max(0, Number(file.size || 0))
    let error: FeedbackDraftImageError | undefined
    if (size > MAX_IMAGE_SIZE)
      error = 'too_large'
    else if (mimeType && !SAFE_IMAGE_MIME_TYPES.has(mimeType))
      error = 'unsupported_type'

    normalized.push({
      clientId: `${batchID}-${index + 1}`,
      filePath,
      name: String(file.name || '').trim() || undefined,
      mimeType: mimeType || undefined,
      size,
      error,
    })
  }
  return normalized
}

export function canSupplementFeedback(status: UserFeedbackStatus, allowsSupplement: boolean) {
  return allowsSupplement && (status === 'pending' || status === 'processing')
}

function errorStatus(value: unknown) {
  if (!value || typeof value !== 'object' || Array.isArray(value))
    return 0
  const record = value as Record<string, unknown>
  const directStatus = Number(record.statusCode || record.status || 0)
  if (directStatus)
    return directStatus
  return errorStatus(record.response)
}

function isFeedbackNotFoundEnvelope(value: unknown) {
  if (!value || typeof value !== 'object' || Array.isArray(value))
    return false
  const record = value as Record<string, unknown>
  const data = record.data
  if (!data || typeof data !== 'object' || Array.isArray(data))
    return false
  const envelope = data as Record<string, unknown>
  const code = Number(envelope.code || 0)
  return Number.isFinite(code)
    && code !== 0
    && String(envelope.msg || '').trim() === '反馈不存在'
}

function feedbackBusinessError(value: unknown): { code: number, message: string } | null {
  if (!value || typeof value !== 'object' || Array.isArray(value))
    return null
  const record = value as Record<string, unknown>
  const code = Number(record.code || 0)
  const message = String(record.msg || record.message || '').trim()
  if (Number.isFinite(code) && code !== 0 && message)
    return { code, message }
  return feedbackBusinessError(record.data)
}

export function feedbackSupplementVersionConflictMessage(error: unknown) {
  const businessError = feedbackBusinessError(error)
  return businessError?.message === '反馈已更新，请刷新后重试'
    ? businessError.message
    : ''
}

export async function recoverFeedbackSupplementVersionConflict(
  error: unknown,
  options: FeedbackVersionConflictRecoveryOptions,
) {
  const message = feedbackSupplementVersionConflictMessage(error)
  if (!message)
    return false
  options.notify(message)
  await options.refresh()
  return true
}

export function feedbackDetailIsInaccessible(error: unknown) {
  const status = errorStatus(error)
  return status === 403 || status === 404 || isFeedbackNotFoundEnvelope(error)
}

export function sortFeedbackMessages<T extends Pick<UserFeedbackMessage, 'createdAt'>>(messages: readonly T[]) {
  return messages
    .map((message, index) => ({ index, message }))
    .sort((left, right) => (
      Number(left.message.createdAt || 0) - Number(right.message.createdAt || 0)
      || left.index - right.index
    ))
    .map(item => item.message)
}
