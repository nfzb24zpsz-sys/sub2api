import { describe, expect, it, vi } from 'vitest'
import { detectServiceBootstrapTargetOS, downloadBlobFile } from '@/utils/serviceBootstrap'

describe('serviceBootstrap utils', () => {
  it('defaults to unix outside the browser', () => {
    expect(detectServiceBootstrapTargetOS()).toBe('unix')
  })

  it('downloads a blob with the provided filename', () => {
    URL.createObjectURL = vi.fn(() => 'blob:test')
    URL.revokeObjectURL = vi.fn()
    const click = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})

    downloadBlobFile('setup.sh', new Blob(['x']))

    expect(URL.createObjectURL).toHaveBeenCalled()
    expect(click).toHaveBeenCalled()
    expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:test')
    click.mockRestore()
  })
})
