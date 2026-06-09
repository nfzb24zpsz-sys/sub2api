<template>
  <AppLayout>
    <div class="mx-auto max-w-5xl px-4 py-6 sm:px-6 lg:px-8">
      <header class="mb-5 flex flex-col gap-3 border-b border-gray-200 pb-5 dark:border-dark-700 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">
            {{ t('services.title') }}
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ t('services.description') }}
          </p>
        </div>
        <div class="inline-flex w-fit items-center rounded-full border border-blue-200 bg-blue-50 px-4 py-2 text-sm font-semibold text-blue-700 dark:border-blue-500/30 dark:bg-blue-500/10 dark:text-blue-200">
          {{ t('services.balancePill', { balance: formatMoney(balance) }) }}
        </div>
      </header>

      <div v-if="loading" class="flex min-h-[280px] items-center justify-center">
        <LoadingSpinner size="lg" />
      </div>

      <div v-else-if="loadError" class="rounded-2xl border border-red-200 bg-red-50 p-6 text-center dark:border-red-500/30 dark:bg-red-500/10">
        <p class="font-medium text-red-700 dark:text-red-200">{{ loadError }}</p>
        <button class="btn btn-secondary mt-4" @click="loadServices">
          {{ t('common.retry') }}
        </button>
      </div>

      <div v-else-if="serviceRows.length === 0" class="rounded-2xl border border-dashed border-gray-300 p-10 text-center dark:border-dark-600">
        <p class="font-medium text-gray-800 dark:text-gray-100">{{ t('services.emptyTitle') }}</p>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('services.emptyDescription') }}</p>
      </div>

      <div v-else class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <div
          v-for="row in serviceRows"
          :key="row.group.id"
          class="grid gap-4 border-b border-gray-100 p-4 transition last:border-b-0 dark:border-dark-700 sm:grid-cols-[1fr_auto] sm:items-center sm:p-5"
          :class="row.available ? 'bg-white dark:bg-dark-800' : 'bg-gray-50 opacity-60 dark:bg-dark-900'"
        >
          <div class="flex min-w-0 gap-4">
            <div
              class="mt-0.5 flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-2xl border"
              :class="row.available ? 'border-gray-200 bg-gray-50 text-gray-700 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-100' : 'border-gray-200 bg-gray-100 text-gray-400 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-400'"
            >
              <PlatformIcon :platform="row.group.platform" size="lg" />
            </div>
            <div class="min-w-0">
              <h2 class="truncate text-base font-bold text-gray-950 dark:text-white">
                {{ row.group.name }}
              </h2>
              <p class="mt-1 text-sm font-medium text-gray-600 dark:text-gray-300">
                {{ row.statusText }}
              </p>
              <p v-if="row.valueText" class="mt-0.5 text-xs text-gray-400 dark:text-gray-500">
                {{ row.valueText }}
              </p>
            </div>
          </div>

          <div class="flex flex-col gap-2 sm:items-end">
            <div class="flex flex-wrap gap-2 sm:justify-end">
              <button
                class="btn btn-primary"
                :disabled="!row.available || credential.isGroupLoading(row.group.id)"
                @click="handleOpen(row)"
              >
                {{ credential.isGroupLoading(row.group.id) ? t('services.processing') : t('services.openNow') }}
              </button>
              <button
                class="btn btn-secondary"
                :disabled="!row.available || credential.isGroupLoading(row.group.id)"
                @click="credential.copyCredential(row.group)"
              >
                {{ t('services.copy') }}
              </button>
              <button class="btn btn-secondary" @click="goToPurchase(row)">
                {{ row.ctaText }}
              </button>
            </div>
            <p class="text-xs text-gray-400 dark:text-gray-500">
              {{ t('services.openHint') }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <BaseDialog
      :show="showAntigravityClientDialog"
      :title="t('services.selectClientTitle')"
      width="narrow"
      @close="showAntigravityClientDialog = false"
    >
      <div class="space-y-3">
        <p class="text-sm text-gray-500 dark:text-gray-400">
          {{ t('services.selectClientDescription') }}
        </p>
        <div class="grid gap-3 sm:grid-cols-2">
          <button class="rounded-xl border border-gray-200 p-4 text-left transition hover:border-blue-300 hover:bg-blue-50 dark:border-dark-600 dark:hover:border-blue-500/40 dark:hover:bg-blue-500/10" @click="confirmAntigravityClient('claude')">
            <span class="font-semibold text-gray-900 dark:text-white">{{ t('services.clientClaude') }}</span>
            <span class="mt-1 block text-xs text-gray-500 dark:text-gray-400">{{ t('services.clientClaudeHint') }}</span>
          </button>
          <button class="rounded-xl border border-gray-200 p-4 text-left transition hover:border-blue-300 hover:bg-blue-50 dark:border-dark-600 dark:hover:border-blue-500/40 dark:hover:bg-blue-500/10" @click="confirmAntigravityClient('gemini')">
            <span class="font-semibold text-gray-900 dark:text-white">{{ t('services.clientGemini') }}</span>
            <span class="mt-1 block text-xs text-gray-500 dark:text-gray-400">{{ t('services.clientGeminiHint') }}</span>
          </button>
        </div>
      </div>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { userGroupsAPI } from '@/api/groups'
import subscriptionsAPI from '@/api/subscriptions'
import { useServiceCredential } from '@/composables/useServiceCredential'
import { useAuthStore, useAppStore } from '@/stores'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import type { CcSwitchClientType } from '@/utils/ccswitchImport'
import type { Group, UserSubscription } from '@/types'

interface ServiceRow {
  group: Group
  available: boolean
  statusText: string
  valueText: string
  ctaText: string
}

const { t } = useI18n()
const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()

const groups = ref<Group[]>([])
const availableGroupIds = ref<Set<number>>(new Set())
const userGroupRates = ref<Record<number, number>>({})
const subscriptions = ref<UserSubscription[]>([])
const loading = ref(true)
const loadError = ref('')
const showAntigravityClientDialog = ref(false)
const pendingAntigravityGroup = ref<Group | null>(null)

const balance = computed(() => authStore.user?.balance ?? 0)

const credential = useServiceCredential({
  getBaseUrl: () => appStore.apiBaseUrl || window.location.origin,
  getProviderName: () => (appStore.siteName || 'sub2api').trim() || 'sub2api',
})

const serviceRows = computed<ServiceRow[]>(() => {
  return groups.value
    .slice()
    .sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0) || a.id - b.id)
    .map((group) => buildServiceRow(group))
})

function buildServiceRow(group: Group): ServiceRow {
  if (group.subscription_type === 'subscription') {
    const subscription = findLatestSubscription(group.id)
    const daysRemaining = getDaysRemaining(subscription)
    const isActive = subscription?.status === 'active' && daysRemaining > 0

    return {
      group,
      available: isActive && availableGroupIds.value.has(group.id),
      statusText: formatSubscriptionStatus(subscription, daysRemaining),
      valueText: formatSubscriptionLimit(group),
      ctaText: t('services.openSubscription'),
    }
  }

  const hasAccess = availableGroupIds.value.has(group.id)
  const canUse = balance.value > 0 && hasAccess
  return {
    group,
    available: canUse,
    statusText: !hasAccess
      ? t('services.status.standardUnavailable')
      : canUse
      ? t('services.status.standardActive', { balance: formatMoney(balance.value) })
      : t('services.status.standardInsufficient'),
    valueText: formatStandardValue(group),
    ctaText: t('services.recharge'),
  }
}

function formatSubscriptionStatus(subscription: UserSubscription | undefined, daysRemaining: number): string {
  if (!subscription) {
    return t('services.status.subscriptionMissing')
  }
  if (subscription.status === 'active' && daysRemaining > 0) {
    return t('services.status.subscriptionActive', { days: daysRemaining })
  }
  if (subscription.status === 'expired' || subscription.status === 'active') {
    return t('services.status.subscriptionExpired')
  }
  if (subscription.status === 'suspended') {
    return t('services.status.subscriptionSuspended')
  }
  return t('services.status.subscriptionUnavailable')
}

function findLatestSubscription(groupId: number): UserSubscription | undefined {
  return subscriptions.value
    .filter((subscription) => subscription.group_id === groupId)
    .sort((a, b) => Date.parse(b.expires_at) - Date.parse(a.expires_at))[0]
}

function getDaysRemaining(subscription?: UserSubscription): number {
  if (!subscription) {
    return 0
  }
  const remainingMs = Date.parse(subscription.expires_at) - Date.now()
  return remainingMs > 0 ? Math.ceil(remainingMs / 86400000) : 0
}

function formatMoney(value: number): string {
  return value.toFixed(2)
}

function formatTokenValue(value: number): string {
  if (!Number.isFinite(value)) {
    return '0'
  }
  return Number.isInteger(value) ? String(value) : value.toFixed(2).replace(/\.?0+$/, '')
}

function formatStandardValue(group: Group): string {
  const effectiveRate = userGroupRates.value[group.id] ?? group.rate_multiplier
  if (!Number.isFinite(effectiveRate) || effectiveRate <= 0) {
    return t('services.value.standardUnlimited')
  }
  return t('services.value.standard', { value: formatTokenValue(1 / effectiveRate) })
}

function formatSubscriptionLimit(group: Group): string {
  if (group.daily_limit_usd && group.daily_limit_usd > 0) {
    return t('services.value.daily', { value: formatTokenValue(group.daily_limit_usd) })
  }
  if (group.weekly_limit_usd && group.weekly_limit_usd > 0) {
    return t('services.value.weekly', { value: formatTokenValue(group.weekly_limit_usd) })
  }
  if (group.monthly_limit_usd && group.monthly_limit_usd > 0) {
    return t('services.value.monthly', { value: formatTokenValue(group.monthly_limit_usd) })
  }
  return ''
}

function handleOpen(row: ServiceRow) {
  if (row.group.platform === 'antigravity') {
    pendingAntigravityGroup.value = row.group
    showAntigravityClientDialog.value = true
    return
  }
  credential.importToCcSwitch(row.group)
}

function confirmAntigravityClient(clientType: CcSwitchClientType) {
  const group = pendingAntigravityGroup.value
  showAntigravityClientDialog.value = false
  pendingAntigravityGroup.value = null
  if (group) {
    credential.importToCcSwitch(group, clientType)
  }
}

function goToPurchase(row: ServiceRow) {
  router.push({
    path: '/purchase',
    query: row.group.subscription_type === 'subscription'
      ? { tab: 'subscription', group: String(row.group.id) }
      : { tab: 'recharge', group: String(row.group.id) },
  })
}

async function loadServices() {
  loading.value = true
  loadError.value = ''
  try {
    const [catalog, available, rates, mySubscriptions] = await Promise.all([
      userGroupsAPI.getCatalog(),
      userGroupsAPI.getAvailable(),
      userGroupsAPI.getUserGroupRates(),
      subscriptionsAPI.getMySubscriptions(),
      authStore.refreshUser(),
      appStore.fetchPublicSettings(),
    ])

    groups.value = catalog
    availableGroupIds.value = new Set(available.map((group) => group.id))
    userGroupRates.value = rates
    subscriptions.value = mySubscriptions
  } catch (error: any) {
    loadError.value = error?.response?.data?.detail || t('services.loadFailed')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadServices()
})
</script>
