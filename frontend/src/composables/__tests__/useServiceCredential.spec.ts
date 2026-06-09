import { beforeEach, describe, expect, it, vi } from 'vitest'

const { list, create, downloadBootstrapScript, copyToClipboard, showError, showSuccess, downloadBlobFile } = vi.hoisted(() => ({
  list: vi.fn(),
  create: vi.fn(),
  downloadBootstrapScript: vi.fn(),
  copyToClipboard: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
  downloadBlobFile: vi.fn(),
}))

vi.mock('@/api/keys', () => ({
  keysAPI: {
    list,
    create,
    downloadBootstrapScript,
  },
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard,
  }),
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

vi.mock('@/utils/serviceBootstrap', async () => {
  const actual = await vi.importActual<typeof import('@/utils/serviceBootstrap')>('@/utils/serviceBootstrap')
  return {
    ...actual,
    detectServiceBootstrapTargetOS: () => 'unix',
    downloadBlobFile,
  }
})

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

import { useServiceCredential } from '@/composables/useServiceCredential'
import type { Group } from '@/types'

const group = {
  id: 7,
  name: 'Claude',
  platform: 'anthropic',
  rate_multiplier: 0.2,
  subscription_type: 'standard',
  status: 'active',
} as Group

describe('useServiceCredential', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    copyToClipboard.mockResolvedValue(true)
    create.mockResolvedValue({ id: 2, key: 'sk-created', group_id: 7, status: 'active' })
    downloadBootstrapScript.mockResolvedValue({ blob: new Blob(['setup']), filename: 'erqishi-claude-setup.zip' })
  })

  it('reuses an existing active key for copy', async () => {
    list.mockResolvedValue({
      items: [{ id: 1, key: 'sk-existing', group_id: 7, status: 'active' }],
    })

    const credential = useServiceCredential({
      getBaseUrl: () => 'https://api.example.com',
    })

    await credential.copyCredential(group)

    expect(create).not.toHaveBeenCalled()
    expect(copyToClipboard).toHaveBeenCalledWith('sk-existing', 'services.copySuccess')
  })

  it('creates a key lazily when the service has no active key', async () => {
    list.mockResolvedValue({ items: [] })

    const credential = useServiceCredential({
      getBaseUrl: () => 'https://api.example.com',
    })

    await credential.copyCredential(group)

    expect(create).toHaveBeenCalledWith('Claude services.connectionSuffix', 7)
    expect(copyToClipboard).toHaveBeenCalledWith('sk-created', 'services.copySuccess')
  })

  it('downloads a bootstrap script with the service credential', async () => {
    list.mockResolvedValue({
      items: [{ id: 1, key: 'sk-existing', group_id: 7, status: 'active' }],
    })

    const credential = useServiceCredential({
      getBaseUrl: () => 'https://api.example.com',
    })

    await credential.downloadBootstrapScript(group)

    expect(downloadBootstrapScript).toHaveBeenCalledWith({
      group_id: 7,
      os: 'unix',
      base_url: 'https://api.example.com',
    })
    expect(downloadBlobFile).toHaveBeenCalled()
    const [filename] = downloadBlobFile.mock.calls[0]
    expect(filename).toBe('erqishi-claude-setup.zip')
    expect(showSuccess).toHaveBeenCalledWith('services.bootstrapDownloaded')
  })
})
