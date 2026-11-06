export const CHUNK_RELOAD_COOLDOWN_MS = 60_000

const chunkReloadStorageKey = 'wecheckin:admin:chunk-reload-at'

interface EventTargetLike {
  addEventListener: (type: string, listener: EventListener) => void
  removeEventListener: (type: string, listener: EventListener) => void
}

interface StorageLike {
  getItem: (key: string) => string | null
  setItem: (key: string, value: string) => void
}

interface ChunkLoadRecoveryOptions {
  target?: EventTargetLike
  storage?: StorageLike
  now?: () => number
  reload?: () => void
}

export function installVitePreloadErrorRecovery(options: ChunkLoadRecoveryOptions = {}) {
  const target = options.target ?? window
  const storage = options.storage ?? window.sessionStorage
  const now = options.now ?? Date.now
  const reload = options.reload ?? (() => window.location.reload())

  const onPreloadError: EventListener = (event) => {
    const currentTime = now()
    let lastReloadAt = 0
    try {
      lastReloadAt = Number(storage.getItem(chunkReloadStorageKey) || 0)
    }
    catch {
      return
    }
    const elapsed = currentTime - lastReloadAt
    if (lastReloadAt > 0 && elapsed >= 0 && elapsed <= CHUNK_RELOAD_COOLDOWN_MS)
      return

    try {
      storage.setItem(chunkReloadStorageKey, String(currentTime))
    }
    catch {
      return
    }
    event.preventDefault()
    reload()
  }

  target.addEventListener('vite:preloadError', onPreloadError)
  return () => target.removeEventListener('vite:preloadError', onPreloadError)
}
