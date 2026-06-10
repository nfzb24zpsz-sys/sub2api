<template>
  <BaseDialog :show="show" :title="dialogTitle" width="full" @close="emit('close')">
    <div class="space-y-6">
      <section class="space-y-3">
        <div class="flex items-center gap-3">
          <div class="flex h-11 w-11 items-center justify-center rounded-xl border border-gray-200 bg-gray-50 dark:border-dark-600 dark:bg-dark-700">
            <PlatformIcon :platform="group?.platform || 'anthropic'" size="lg" />
          </div>
          <div class="min-w-0">
            <h2 class="text-lg font-semibold text-gray-950 dark:text-white">
              {{ title }}
            </h2>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              {{ description }}
            </p>
          </div>
        </div>

        <div class="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-100">
          {{ detectionNote }}
        </div>
      </section>

      <section class="space-y-3">
        <div class="flex items-center justify-between gap-3">
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ t('services.guide.chooseClientTitle') }}
          </h3>
          <p class="text-xs text-gray-500 dark:text-gray-400">
            {{ t('services.guide.chooseClientHint') }}
          </p>
        </div>

        <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          <div
            v-for="client in clients"
            :key="client.id"
            role="button"
            tabindex="0"
            class="group flex h-full min-h-[148px] flex-col gap-3 rounded-2xl border p-4 text-left transition"
            :class="selectedClientId === client.id
              ? 'border-primary-500 bg-primary-50 dark:border-primary-500 dark:bg-primary-500/10'
              : 'border-gray-200 bg-white hover:border-primary-300 hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-800 dark:hover:border-primary-500/60 dark:hover:bg-dark-700'"
            @click="selectClient(client.id)"
            @keydown.enter.prevent="selectClient(client.id)"
            @keydown.space.prevent="selectClient(client.id)"
          >
            <div class="flex items-center justify-between gap-3">
              <div class="flex h-11 w-11 items-center justify-center rounded-xl border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-900">
                <component :is="client.icon" class="h-6 w-6" />
              </div>
              <span
                class="inline-flex items-center rounded-full px-2.5 py-1 text-xs font-medium"
                :class="getClientStatusClass(client.id)"
              >
                {{ getClientStatusText(client.id) }}
              </span>
            </div>

            <div class="space-y-1">
              <div class="flex items-center gap-2">
                <h4 class="text-base font-semibold text-gray-950 dark:text-white">
                  {{ client.label }}
                </h4>
                <Icon v-if="selectedClientId === client.id" name="checkCircle" size="sm" class="text-primary-500" />
              </div>
              <p class="text-sm leading-6 text-gray-500 dark:text-gray-400">
                {{ client.description }}
              </p>
            </div>

            <div class="mt-auto flex items-center justify-between gap-2">
              <span class="text-xs text-gray-400 dark:text-gray-500">
                {{ client.installHint }}
              </span>
              <button
                v-if="client.downloadUrl"
                class="inline-flex items-center gap-1.5 text-xs font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400"
                @click.stop="openUrl(client.downloadUrl)"
              >
                <Icon name="externalLink" size="sm" />
                {{ t('services.guide.download') }}
              </button>
            </div>
          </div>
        </div>
      </section>

      <section v-if="selectedClient" class="space-y-3">
        <div class="flex items-center justify-between gap-3">
          <div>
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ t('services.guide.stepsTitle', { client: selectedClient.label }) }}
            </h3>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ selectedClient.description }}
            </p>
          </div>
          <div class="flex items-center gap-2">
            <button
              class="btn btn-secondary btn-sm"
              :disabled="currentStep === 0"
              @click="currentStep = Math.max(0, currentStep - 1)"
            >
              <Icon name="chevronLeft" size="sm" />
              {{ t('common.back') }}
            </button>
            <button
              class="btn btn-primary btn-sm"
              :disabled="currentStep >= selectedClient.steps.length - 1"
              @click="currentStep = Math.min(selectedClient.steps.length - 1, currentStep + 1)"
            >
              {{ t('common.next') }}
              <Icon name="chevronRight" size="sm" />
            </button>
          </div>
        </div>

        <div class="grid gap-4 lg:grid-cols-[220px_minmax(0,1fr)]">
          <div class="space-y-2 rounded-2xl border border-gray-200 bg-white p-3 dark:border-dark-600 dark:bg-dark-800">
            <button
              v-for="(step, index) in selectedClient.steps"
              :key="step.title"
              class="flex w-full items-start gap-3 rounded-xl px-3 py-2 text-left transition"
              :class="index === currentStep
                ? 'bg-primary-50 text-primary-700 dark:bg-primary-500/10 dark:text-primary-200'
                : 'text-gray-500 hover:bg-gray-50 dark:text-dark-300 dark:hover:bg-dark-700'"
              @click="currentStep = index"
            >
              <span
                class="mt-0.5 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-xs font-semibold"
                :class="index <= currentStep
                  ? 'bg-primary-600 text-white'
                  : 'bg-gray-200 text-gray-500 dark:bg-dark-600 dark:text-dark-300'"
              >
                {{ index + 1 }}
              </span>
              <span class="min-w-0">
                <span class="block text-sm font-medium">{{ step.title }}</span>
                <span class="block text-xs leading-5 opacity-80">{{ step.summary }}</span>
              </span>
            </button>
          </div>

          <div class="space-y-4 rounded-2xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
            <div class="flex items-center justify-between gap-3">
              <div>
                <p class="text-xs font-medium uppercase tracking-wide text-gray-400 dark:text-gray-500">
                  {{ t('services.guide.currentStep', { index: currentStep + 1, total: selectedClient.steps.length }) }}
                </p>
                <h4 class="text-lg font-semibold text-gray-950 dark:text-white">
                  {{ currentStepStep.title }}
                </h4>
              </div>
              <button
                class="btn btn-secondary btn-sm"
                @click="markInstalled(selectedClient.id)"
              >
                {{ t('services.guide.haveInstalled') }}
              </button>
            </div>

            <p class="text-sm leading-6 text-gray-600 dark:text-gray-300">
              {{ currentStepStep.description }}
            </p>

            <div v-if="currentStepStep.links?.length" class="flex flex-wrap gap-2">
              <a
                v-for="link in currentStepStep.links"
                :key="link.label"
                class="btn btn-primary btn-sm"
                :href="link.href"
                target="_blank"
                rel="noreferrer"
              >
                <Icon name="externalLink" size="sm" />
                {{ link.label }}
              </a>
            </div>

            <div class="space-y-2 rounded-xl border border-gray-200 bg-gray-50 p-3 dark:border-dark-600 dark:bg-dark-900">
              <p class="text-xs font-semibold uppercase tracking-wide text-gray-400 dark:text-gray-500">
                {{ t('services.guide.configTitle') }}
              </p>
              <pre class="overflow-x-auto text-xs leading-6 text-gray-800 dark:text-gray-200"><code>{{ currentStepStep.config }}</code></pre>
            </div>

            <div class="rounded-xl border border-dashed border-gray-200 p-3 text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400">
              {{ t('services.guide.installStatusHint') }}
            </div>
          </div>
        </div>
      </section>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, h, ref, watch, type Component } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import type { Group } from '@/types'

interface Props {
  show: boolean
  group: Group | null
  apiKey: string
  baseUrl: string
}

interface Emits {
  (e: 'close'): void
}

interface ClientStep {
  title: string
  summary: string
  description: string
  config: string
  links?: Array<{ label: string; href: string }>
}

interface ClientOption {
  id: string
  label: string
  description: string
  installHint: string
  icon: Component
  downloadUrl?: string
  steps: ClientStep[]
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()
const { t } = useI18n()

const selectedClientId = ref('codex')
const currentStep = ref(0)
const confirmedInstalledClientIds = ref<Set<string>>(new Set())

watch(
  () => props.show,
  (visible) => {
    if (visible) {
      selectedClientId.value = 'codex'
      currentStep.value = 0
      confirmedInstalledClientIds.value = new Set(
        ['codex', 'cursor', 'opencode', 'qoder'].filter((id) => localStorage.getItem(`service-guide-installed:${id}`) === '1')
      )
    }
  },
  { immediate: true }
)

const title = computed(() => t('services.guide.title'))
const dialogTitle = computed(() => props.group ? `${props.group.name} · ${title.value}` : title.value)
const description = computed(() => t('services.guide.description'))
const detectionNote = computed(() => t('services.guide.detectionNote'))
const baseRoot = computed(() => (props.baseUrl || window.location.origin).replace(/\/v1\/?$/, '').replace(/\/+$/, ''))
const openAIBase = computed(() => `${baseRoot.value}/v1`)
const anthropicBase = computed(() => baseRoot.value)
const geminiBase = computed(() => `${baseRoot.value}/v1beta`)
const antigravityBase = computed(() => `${baseRoot.value}/antigravity/v1`)
const preferredBase = computed(() => {
  switch (props.group?.platform) {
    case 'anthropic':
      return anthropicBase.value
    case 'gemini':
      return geminiBase.value
    case 'antigravity':
      return antigravityBase.value
    default:
      return openAIBase.value
  }
})

const clients = computed<ClientOption[]>(() => [
  {
    id: 'codex',
    label: 'Codex',
    description: t('services.guide.clients.codex.description'),
    installHint: t('services.guide.clients.codex.installHint'),
    icon: terminalIcon,
    downloadUrl: 'https://openai.com/codex',
    steps: buildCodexSteps(),
  },
  {
    id: 'cursor',
    label: 'Cursor',
    description: t('services.guide.clients.cursor.description'),
    installHint: t('services.guide.clients.cursor.installHint'),
    icon: cursorIcon,
    downloadUrl: 'https://cursor.com',
    steps: buildCursorSteps(),
  },
  {
    id: 'opencode',
    label: 'OpenCode',
    description: t('services.guide.clients.opencode.description'),
    installHint: t('services.guide.clients.opencode.installHint'),
    icon: openCodeIcon,
    downloadUrl: 'https://opencode.ai',
    steps: buildOpenCodeSteps(),
  },
  {
    id: 'qoder',
    label: 'Qoder',
    description: t('services.guide.clients.qoder.description'),
    installHint: t('services.guide.clients.qoder.installHint'),
    icon: qoderIcon,
    downloadUrl: 'https://qoder.com',
    steps: buildQoderSteps(),
  },
])

const selectedClient = computed(() => clients.value.find((client) => client.id === selectedClientId.value) || clients.value[0])
const currentStepStep = computed(() => selectedClient.value.steps[Math.min(currentStep.value, selectedClient.value.steps.length - 1)])

function selectClient(clientId: string) {
  selectedClientId.value = clientId
  currentStep.value = 0
}

function markInstalled(clientId: string) {
  localStorage.setItem(`service-guide-installed:${clientId}`, '1')
  confirmedInstalledClientIds.value = new Set([...confirmedInstalledClientIds.value, clientId])
}

function openUrl(url: string) {
  window.open(url, '_blank', 'noopener,noreferrer')
}

function getClientStatusText(clientId: string): string {
  return confirmedInstalledClientIds.value.has(clientId)
    ? t('services.guide.status.confirmed')
    : t('services.guide.status.unknown')
}

function getClientStatusClass(clientId: string): string {
  return confirmedInstalledClientIds.value.has(clientId)
    ? 'bg-green-100 text-green-700 dark:bg-green-500/10 dark:text-green-300'
    : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-200'
}

function buildCodexSteps(): ClientStep[] {
  return [
    {
      title: t('services.guide.stepTitles.download'),
      summary: t('services.guide.stepSummaries.codex.download'),
      description: t('services.guide.clients.codex.steps.download'),
      config: 'npm install -g @openai/codex',
      links: [{ label: t('services.guide.openDownloadPage'), href: 'https://openai.com/codex' }],
    },
    {
      title: t('services.guide.stepTitles.configure'),
      summary: t('services.guide.stepSummaries.codex.configure'),
      description: t('services.guide.clients.codex.steps.configure'),
      config: `model_provider = "OpenAI"
model = "gpt-5.5"

[model_providers.OpenAI]
name = "OpenAI"
base_url = "${openAIBase.value}"
wire_api = "responses"
requires_openai_auth = true

auth.json:
{
  "OPENAI_API_KEY": "${props.apiKey}"
}`,
    },
    {
      title: t('services.guide.stepTitles.start'),
      summary: t('services.guide.stepSummaries.start'),
      description: t('services.guide.clients.codex.steps.start'),
      config: 'codex chat',
    },
  ]
}

function buildCursorSteps(): ClientStep[] {
  return [
    {
      title: t('services.guide.stepTitles.download'),
      summary: t('services.guide.stepSummaries.cursor.download'),
      description: t('services.guide.clients.cursor.steps.download'),
      config: 'Open Cursor Settings > Models > Add new API key',
      links: [{ label: t('services.guide.openDownloadPage'), href: 'https://cursor.com' }],
    },
    {
      title: t('services.guide.stepTitles.configure'),
      summary: t('services.guide.stepSummaries.cursor.configure'),
      description: t('services.guide.clients.cursor.steps.configure'),
      config: `Provider: OpenAI compatible
Base URL: ${preferredBase.value}
API Key: ${props.apiKey}`,
    },
    {
      title: t('services.guide.stepTitles.verify'),
      summary: t('services.guide.stepSummaries.verify'),
      description: t('services.guide.clients.cursor.steps.verify'),
      config: 'Create a new chat and send a test prompt.',
    },
  ]
}

function buildOpenCodeSteps(): ClientStep[] {
  return [
    {
      title: t('services.guide.stepTitles.download'),
      summary: t('services.guide.stepSummaries.opencode.download'),
      description: t('services.guide.clients.opencode.steps.download'),
      config: 'npm install -g opencode',
      links: [{ label: t('services.guide.openDownloadPage'), href: 'https://opencode.ai' }],
    },
    {
      title: t('services.guide.stepTitles.configure'),
      summary: t('services.guide.stepSummaries.opencode.configure'),
      description: t('services.guide.clients.opencode.steps.configure'),
      config: JSON.stringify({
        provider: {
          openai: {
            options: {
              baseURL: preferredBase.value,
              apiKey: props.apiKey
            }
          }
        }
      }, null, 2),
    },
    {
      title: t('services.guide.stepTitles.start'),
      summary: t('services.guide.stepSummaries.start'),
      description: t('services.guide.clients.opencode.steps.start'),
      config: 'opencode',
    },
  ]
}

function buildQoderSteps(): ClientStep[] {
  return [
    {
      title: t('services.guide.stepTitles.download'),
      summary: t('services.guide.stepSummaries.qoder.download'),
      description: t('services.guide.clients.qoder.steps.download'),
      config: 'Install Qoder from the official download page.',
      links: [{ label: t('services.guide.openDownloadPage'), href: 'https://qoder.com' }],
    },
    {
      title: t('services.guide.stepTitles.configure'),
      summary: t('services.guide.stepSummaries.qoder.configure'),
      description: t('services.guide.clients.qoder.steps.configure'),
      config: `Provider: OpenAI compatible
Base URL: ${preferredBase.value}
API Key: ${props.apiKey}`,
    },
    {
      title: t('services.guide.stepTitles.verify'),
      summary: t('services.guide.stepSummaries.verify'),
      description: t('services.guide.clients.qoder.steps.verify'),
      config: 'Open a new project and run a small prompt test.',
    },
  ]
}

const terminalIcon = {
  render() {
    return h('svg', { viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '1.5' }, [
      h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', d: 'm6.75 7.5 3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0 0 21 18V6A2.25 2.25 0 0 0 18.75 3.75H5.25A2.25 2.25 0 0 0 3 6v12A2.25 2.25 0 0 0 5.25 21.75Z' })
    ])
  }
}

const cursorIcon = {
  render() {
    return h('svg', { viewBox: '0 0 24 24', fill: 'currentColor' }, [
      h('path', { d: 'M4 3.5l15 8.5-6.5 1.6L11 20l-1.8-.7L4 3.5z' })
    ])
  }
}

const openCodeIcon = {
  render() {
    return h('svg', { viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '1.5' }, [
      h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', d: 'M7.5 8.25 3.75 12l3.75 3.75M16.5 8.25 20.25 12l-3.75 3.75M14.25 4.5 9.75 19.5' })
    ])
  }
}

const qoderIcon = {
  render() {
    return h('svg', { viewBox: '0 0 24 24', fill: 'currentColor' }, [
      h('path', { d: 'M12 2.5 21 7.5v9L12 21.5 3 16.5v-9L12 2.5zm0 2.2L5 8.3v7.4l7 3.9 7-3.9V8.3l-7-3.6z' })
    ])
  }
}
</script>
