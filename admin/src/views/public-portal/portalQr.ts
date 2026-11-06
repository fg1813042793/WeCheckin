import { safeTargetUrl } from './portalModel'

export const QR_OPTIONS = {
  width: 1024,
  margin: 4,
  errorCorrectionLevel: 'M' as const,
  color: { dark: '#000000ff', light: '#ffffffff' },
}

export async function generateActivityQr(url: string): Promise<string> {
  const safe = safeTargetUrl(url)
  if (!safe) throw new Error('活动链接无效，无法生成二维码')
  const { toDataURL } = await import('qrcode')
  const dataUrl = await toDataURL(safe, QR_OPTIONS)
  if (!dataUrl.startsWith('data:image/png;base64,')) throw new Error('二维码生成失败，请重试')
  return dataUrl
}

export function downloadActivityQr(dataUrl: string, filename: string): void {
  if (!dataUrl.startsWith('data:image/png;base64,')) throw new Error('二维码尚未生成，请重试')
  const anchor = document.createElement('a')
  anchor.href = dataUrl
  anchor.download = filename
  document.body.appendChild(anchor)
  try { anchor.click() } finally { anchor.remove() }
}
