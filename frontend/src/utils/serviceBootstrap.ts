export type ServiceBootstrapTargetOS = 'unix' | 'windows'

export function detectServiceBootstrapTargetOS(): ServiceBootstrapTargetOS {
  if (typeof navigator === 'undefined') {
    return 'unix'
  }
  const platform = [
    navigator.userAgent,
    (navigator as Navigator & { userAgentData?: { platform?: string } }).userAgentData?.platform,
    navigator.platform,
  ].join(' ')
  return /windows|win32|win64/i.test(platform) ? 'windows' : 'unix'
}

export function downloadBlobFile(filename: string, blob: Blob): void {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}
