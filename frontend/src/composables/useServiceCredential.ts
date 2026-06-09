import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { keysAPI } from '@/api/keys'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores/app'
import {
  buildCcSwitchImportDeeplink,
  type CcSwitchClientType,
} from '@/utils/ccswitchImport'
import type { ApiKey, Group } from '@/types'

interface UseServiceCredentialOptions {
  getBaseUrl: () => string
  getProviderName: () => string
}

const usageScript = `({
  request: {
    url: "{{baseUrl}}/v1/usage",
    method: "GET",
    headers: { "Authorization": "Bearer {{apiKey}}" }
  },
  extractor: function(response) {
    const remaining = response?.remaining ?? response?.quota?.remaining ?? response?.balance;
    const unit = response?.unit ?? response?.quota?.unit ?? "USD";
    return {
      isValid: response?.is_active ?? response?.isValid ?? true,
      remaining,
      unit
    };
  }
})`

function isReusableKey(key: ApiKey, groupId: number): boolean {
  return key.group_id === groupId && key.status === 'active'
}

export function useServiceCredential(options: UseServiceCredentialOptions) {
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

  async function copyCredential(group: Group): Promise<void> {
    setGroupLoading(group.id, true)
    try {
      const key = await getOrCreateCredential(group)
      await copyToClipboard(key.key, t('services.copySuccess'))
    } catch (error: any) {
      appStore.showError(error?.response?.data?.detail || t('services.copyFailed'))
    } finally {
      setGroupLoading(group.id, false)
    }
  }

  async function importToCcSwitch(group: Group, clientType?: CcSwitchClientType): Promise<void> {
    const resolvedClientType: CcSwitchClientType =
      clientType || (group.platform === 'gemini' ? 'gemini' : 'claude')

    setGroupLoading(group.id, true)
    try {
      const key = await getOrCreateCredential(group)
      const deeplink = buildCcSwitchImportDeeplink({
        baseUrl: options.getBaseUrl(),
        platform: group.platform,
        clientType: resolvedClientType,
        providerName: options.getProviderName(),
        apiKey: key.key,
        usageScript,
      })

      window.open(deeplink, '_self')
      setTimeout(() => {
        if (document.hasFocus()) {
          appStore.showError(t('services.ccSwitchNotInstalled'))
        }
      }, 100)
    } catch (error: any) {
      appStore.showError(error?.response?.data?.detail || t('services.openFailed'))
    } finally {
      setGroupLoading(group.id, false)
    }
  }

  function isGroupLoading(groupId: number): boolean {
    return loadingGroupIds.value.has(groupId)
  }

  return {
    copyCredential,
    getOrCreateCredential,
    importToCcSwitch,
    isGroupLoading,
    loadingGroupIds,
  }
}
