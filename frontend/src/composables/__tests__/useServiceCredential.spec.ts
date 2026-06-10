import { beforeEach, describe, expect, it, vi } from 'vitest'

const { list, create, copyToClipboard, showError, showSuccess } = vi.hoisted(() => ({
  list: vi.fn(),
  create: vi.fn(),
  copyToClipboard: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/keys', () => ({
  keysAPI: {
    list,
    create,
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
  })

  it('reuses an existing active key for copy', async () => {
    list.mockResolvedValue({
      items: [{ id: 1, key: 'sk-existing', group_id: 7, status: 'active' }],
    })

    const credential = useServiceCredential()

    await credential.copyCredential(group)

    expect(create).not.toHaveBeenCalled()
    expect(copyToClipboard).toHaveBeenCalledWith('sk-existing', 'services.copySuccess')
  })

  it('creates a key lazily when the service has no active key', async () => {
    list.mockResolvedValue({ items: [] })

    const credential = useServiceCredential()

    await credential.copyCredential(group)

    expect(create).toHaveBeenCalledWith('Claude services.connectionSuffix', 7)
    expect(copyToClipboard).toHaveBeenCalledWith('sk-created', 'services.copySuccess')
  })

  it('ensures a credential without downloading a bootstrap script', async () => {
    list.mockResolvedValue({
      items: [{ id: 1, key: 'sk-existing', group_id: 7, status: 'active' }],
    })

    const credential = useServiceCredential()

    const key = await credential.ensureCredential(group)

    expect(key.key).toBe('sk-existing')
    expect(showSuccess).not.toHaveBeenCalled()
  })
})
