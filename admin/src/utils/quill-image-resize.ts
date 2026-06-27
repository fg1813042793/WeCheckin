import Quill from 'quill'

class ImageResizeModule {
  quill: any
  overlay: HTMLDivElement | null = null
  handle: HTMLDivElement | null = null
  activeImg: HTMLImageElement | null = null
  startX = 0
  startY = 0
  startW = 0
  startH = 0

  constructor(quill: any, _options: any = {}) {
    this.quill = quill
    this.quill.root.addEventListener('click', this.onClick.bind(this))
    this.quill.root.addEventListener('blur', this.onBlur.bind(this))
  }

  onClick(e: Event) {
    const target = e.target as HTMLElement
    if (target.tagName === 'IMG') {
      this.selectImage(target as HTMLImageElement)
    } else {
      this.removeOverlay()
    }
  }

  onBlur() {
    setTimeout(() => this.removeOverlay(), 200)
  }

  selectImage(img: HTMLImageElement) {
    this.removeOverlay()
    this.activeImg = img

    const rect = img.getBoundingClientRect()
    const editorRect = this.quill.root.getBoundingClientRect()

    this.overlay = document.createElement('div')
    this.overlay.className = 'ql-image-overlay'
    Object.assign(this.overlay.style, {
      position: 'absolute',
      top: `${rect.top - editorRect.top + this.quill.root.scrollTop - 2}px`,
      left: `${rect.left - editorRect.left - 2}px`,
      width: `${rect.width + 4}px`,
      height: `${rect.height + 4}px`,
      border: '2px solid #3873f6',
      pointerEvents: 'none',
      boxSizing: 'border-box',
      zIndex: '10'
    })

    this.handle = document.createElement('div')
    Object.assign(this.handle.style, {
      position: 'absolute',
      right: '-6px',
      bottom: '-6px',
      width: '12px',
      height: '12px',
      background: '#3873f6',
      border: '2px solid #fff',
      borderRadius: '50%',
      cursor: 'nwse-resize',
      pointerEvents: 'all',
      zIndex: '11'
    })
    this.handle.addEventListener('mousedown', this.onResizeStart.bind(this))
    this.overlay.appendChild(this.handle)
    this.quill.root.parentElement?.appendChild(this.overlay)
  }

  onResizeStart(e: MouseEvent) {
    e.preventDefault()
    e.stopPropagation()
    this.startX = e.clientX
    this.startY = e.clientY
    this.startW = this.activeImg?.width || 0
    this.startH = this.activeImg?.height || 0
    document.addEventListener('mousemove', this.onResizeMove)
    document.addEventListener('mouseup', this.onResizeEnd)
  }

  onResizeMove = (e: MouseEvent) => {
    if (!this.activeImg || !this.overlay) return
    const dx = e.clientX - this.startX
    const dy = e.clientY - this.startY
    const ratio = this.startW / this.startH
    let newW = this.startW + dx
    let newH = newW / ratio

    if (newW < 20) { newW = 20; newH = newW / ratio }
    if (newH < 20) { newH = 20; newW = newH * ratio }

    this.activeImg.width = Math.round(newW)
    this.activeImg.height = Math.round(newH)

    const rect = this.activeImg.getBoundingClientRect()
    const editorRect = this.quill.root.getBoundingClientRect()
    this.overlay.style.top = `${rect.top - editorRect.top + this.quill.root.scrollTop - 2}px`
    this.overlay.style.left = `${rect.left - editorRect.left - 2}px`
    this.overlay.style.width = `${rect.width + 4}px`
    this.overlay.style.height = `${rect.height + 4}px`
  }

  onResizeEnd = () => {
    document.removeEventListener('mousemove', this.onResizeMove)
    document.removeEventListener('mouseup', this.onResizeEnd)
  }

  removeOverlay() {
    if (this.overlay) { this.overlay.remove(); this.overlay = null }
    this.activeImg = null
    this.handle = null
  }
}

Quill.register('modules/imageResize', ImageResizeModule)
