import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { keysAPI } from '@/api/keys'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores/app'
import type { ApiKey, Group } from '@/types'

function isReusableKey(key: ApiKey, groupId: number): boolean {
  return key.group_id === groupId && key.status === 'active'
}

export function useServiceCredential() {
  const { t } = useI18n()
  const appStore = useAppStore()
  const { copyToClipboard } = useClipboard()
  const loadingGroupIds = ref<Set<number>>(new Set())
  const keyCache = new Map<number, ApiKey>()

  function setGroupLoading(groupId: number, loading: boolean) {
    const next = new Set(loadingGroupIds.value)
    if (loading) {
      next.add(groupId)
    } else {
      next.delete(groupId)
    }
    loadingGroupIds.value = next
  }

  async function getOrCreateCredential(group: Group): Promise<ApiKey> {
    const cached = keyCache.get(group.id)
    if (cached && isReusableKey(cached, group.id)) {
      return cached
    }

    const existing = await keysAPI.list(1, 100, {
      group_id: group.id,
      status: 'active',
      sort_by: 'created_at',
      sort_order: 'asc',
    })
    const reusable = existing.items.find((key) => isReusableKey(key, group.id))
    if (reusable) {
      keyCache.set(group.id, reusable)
      return reusable
    }

    const created = await keysAPI.create(`${group.name} ${t('services.connectionSuffix')}`, group.id)
    keyCache.set(group.id, created)
    return created
  }

  async function ensureCredential(group: Group): Promise<ApiKey> {
    setGroupLoading(group.id, true)
    try {
      return await getOrCreateCredential(group)
    } finally {
      setGroupLoading(group.id, false)
    }
  }

  async function copyCredential(group: Group): Promise<void> {
    try {
      const key = await ensureCredential(group)
      await copyToClipboard(key.key, t('services.copySuccess'))
    } catch (error: any) {
      appStore.showError(error?.response?.data?.detail || t('services.copyFailed'))
    }
  }

  function isGroupLoading(groupId: number): boolean {
    return loadingGroupIds.value.has(groupId)
  }

  return {
    copyCredential,
    getOrCreateCredential,
    ensureCredential,
    isGroupLoading,
    loadingGroupIds,
  }
}
