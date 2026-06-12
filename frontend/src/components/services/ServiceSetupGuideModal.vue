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
      </section>

      <section v-if="phase === 'select'" class="space-y-3">
        <div class="flex items-center justify-between gap-3">
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ t('services.guide.chooseClientTitle') }}
          </h3>
          <p class="text-xs text-gray-500 dark:text-gray-400">
            {{ t('services.guide.chooseClientHint') }}
          </p>
        </div>

        <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          <article
            v-for="client in clients"
            :key="client.id"
            class="flex min-h-[184px] cursor-pointer flex-col gap-3 rounded-2xl border border-gray-200 bg-white p-4 transition hover:border-primary-300 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-primary-500/50 dark:border-dark-600 dark:bg-dark-800 dark:hover:border-primary-500/60 dark:hover:bg-dark-700"
            role="button"
            tabindex="0"
            @click="beginSetup(client.id)"
            @keydown.enter.prevent="beginSetup(client.id)"
            @keydown.space.prevent="beginSetup(client.id)"
          >
            <div class="flex items-center justify-between gap-3">
              <div class="flex h-14 w-14 items-center justify-center overflow-hidden rounded-xl border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-900">
                <img :src="client.icon" :alt="client.label" class="h-full w-full object-contain p-1" />
              </div>
            </div>

            <div class="space-y-1">
              <h4 class="text-base font-semibold text-gray-950 dark:text-white">
                {{ client.label }}
              </h4>
              <p class="text-sm leading-6 text-gray-500 dark:text-gray-400">
                {{ client.description }}
              </p>
            </div>

            <div class="mt-auto flex items-center justify-between gap-2">
              <span class="text-xs text-gray-400 dark:text-gray-500">
                {{ client.installHint }}
              </span>
              <button class="btn btn-primary btn-sm" @click.stop="beginSetup(client.id)">
                {{ t('services.guide.configure') }}
              </button>
            </div>
          </article>
        </div>
      </section>

      <section v-else-if="selectedClient" class="space-y-3">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div class="flex min-w-0 items-center gap-3">
            <div class="flex h-11 w-11 items-center justify-center overflow-hidden rounded-xl border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-900">
              <img :src="selectedClient.icon" :alt="selectedClient.label" class="h-full w-full object-contain p-1" />
            </div>
            <div class="min-w-0">
              <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ t('services.guide.stepsTitle', { client: selectedClient.label }) }}
              </h3>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                {{ selectedClient.description }}
              </p>
            </div>
          </div>

          <div class="flex items-center gap-2">
            <button class="btn btn-secondary btn-sm" @click="backToSelect">
              <Icon name="chevronLeft" size="sm" />
              {{ t('common.back') }}
            </button>
            <button
              class="btn btn-secondary btn-sm"
              :disabled="currentStep === 0"
              @click="currentStep = Math.max(0, currentStep - 1)"
            >
              <Icon name="chevronLeft" size="sm" />
              {{ t('services.guide.previousStep') }}
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
            <div>
              <div>
                <p class="text-xs font-medium uppercase tracking-wide text-gray-400 dark:text-gray-500">
                  {{ t('services.guide.currentStep', { index: currentStep + 1, total: selectedClient.steps.length }) }}
                </p>
                <h4 class="text-lg font-semibold text-gray-950 dark:text-white">
                  {{ currentStepStep.title }}
                </h4>
              </div>
            </div>

            <p class="text-sm leading-6 text-gray-600 dark:text-gray-300">
              {{ currentStepStep.description }}
            </p>

            <div
              v-if="currentStepStep.ccSwitchDownload"
              class="space-y-4 rounded-xl border border-primary-200 bg-primary-50 p-4 dark:border-primary-500/30 dark:bg-primary-500/10"
            >
              <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div class="min-w-0 space-y-1">
                  <div class="flex items-center gap-2">
                    <Icon name="download" size="sm" class="text-primary-600 dark:text-primary-300" />
                    <h5 class="text-sm font-semibold text-primary-900 dark:text-primary-100">
                      {{ t('services.guide.ccSwitchDownload.title') }}
                    </h5>
                  </div>
                  <p class="text-sm leading-6 text-primary-700 dark:text-primary-200">
                    {{ t('services.guide.ccSwitchDownload.description') }}
                  </p>
                  <p class="text-xs text-primary-600/80 dark:text-primary-200/80">
                    {{ t('services.guide.ccSwitchDownload.detected', { os: detectedCcSwitchDownload.osLabel }) }}
                  </p>
                </div>
                <a
                  class="btn btn-primary btn-lg w-full justify-center sm:w-auto"
                  :href="detectedCcSwitchDownload.href"
                  target="_blank"
                  rel="noreferrer"
                >
                  <Icon name="download" size="sm" />
                  {{ detectedCcSwitchDownload.label }}
                </a>
              </div>

              <div
                v-if="showCcSwitchAlternativeLinks"
                class="flex flex-wrap gap-2 border-t border-primary-200/70 pt-3 dark:border-primary-400/20"
              >
                <a
                  v-for="link in ccSwitchDownloadLinks"
                  :key="link.os"
                  class="btn btn-secondary btn-sm"
                  :href="link.href"
                  target="_blank"
                  rel="noreferrer"
                >
                  {{ link.label }}
                </a>
              </div>

              <div class="rounded-xl border border-white/70 bg-white/80 p-4 shadow-sm dark:border-primary-400/20 dark:bg-dark-800/80">
                <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div class="min-w-0 space-y-1">
                    <div class="flex items-center gap-2">
                      <Icon name="link" size="sm" class="text-primary-600 dark:text-primary-300" />
                      <h5 class="text-sm font-semibold text-gray-950 dark:text-white">
                        {{ t('services.guide.ccSwitchImport.title') }}
                      </h5>
                    </div>
                    <p class="text-sm leading-6 text-gray-600 dark:text-gray-300">
                      {{ t('services.guide.ccSwitchImport.description', { model: ccSwitchImportModel }) }}
                    </p>
                  </div>
                  <button class="btn btn-primary btn-lg w-full justify-center sm:w-auto" @click="importCurrentServiceToCcSwitch">
                    <Icon name="link" size="sm" />
                    {{ t('services.guide.ccSwitchImport.button') }}
                  </button>
                </div>
              </div>

              <div class="grid gap-3 md:grid-cols-3">
                <article
                  v-for="(guideImage, index) in ccSwitchGuideImages"
                  :key="guideImage.title"
                  class="overflow-hidden rounded-xl border border-white/80 bg-white shadow-sm dark:border-primary-400/20 dark:bg-dark-800"
                >
                  <div class="flex items-start gap-2 border-b border-gray-100 px-3 py-2 dark:border-dark-600">
                    <span class="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-primary-600 text-xs font-semibold text-white">
                      {{ index + 1 }}
                    </span>
                    <div class="min-w-0">
                      <h6 class="text-sm font-semibold text-gray-900 dark:text-white">{{ guideImage.title }}</h6>
                      <p class="text-xs leading-5 text-gray-500 dark:text-gray-400">{{ guideImage.description }}</p>
                    </div>
                  </div>
                  <img
                    :src="guideImage.src"
                    :alt="guideImage.alt"
                    class="block aspect-[16/10] w-full bg-gray-50 object-contain dark:bg-dark-900"
                  />
                </article>
              </div>
            </div>

            <div
              v-if="currentStepStep.links?.length"
              class="flex flex-wrap gap-2"
              :class="currentStepStep.primaryAction ? 'pt-1' : ''"
            >
              <a
                v-for="link in currentStepStep.links"
                :key="link.label"
                class="btn btn-primary"
                :class="currentStepStep.primaryAction ? 'btn-lg w-full sm:w-auto' : 'btn-sm'"
                :href="link.href"
                target="_blank"
                rel="noreferrer"
              >
                <Icon name="externalLink" size="sm" />
                {{ link.label }}
              </a>
            </div>

            <div v-if="currentStepStep.heroImage" class="overflow-hidden rounded-xl border border-gray-200 bg-gray-50 dark:border-dark-600 dark:bg-dark-900">
              <img
                :src="currentStepStep.heroImage"
                :alt="currentStepStep.heroImageAlt || currentStepStep.title"
                class="block w-full object-cover"
              />
            </div>

            <details
              v-if="currentStepStep.manualConfigFiles?.length"
              class="group rounded-xl border border-gray-200 bg-gray-50 dark:border-dark-600 dark:bg-dark-900"
            >
              <summary class="flex cursor-pointer list-none items-center justify-between gap-3 px-4 py-3 text-sm font-medium text-gray-700 dark:text-gray-200">
                <span>{{ t('services.guide.manualConfig.summary') }}</span>
                <Icon name="chevronDown" size="sm" class="transition group-open:rotate-180" />
              </summary>
              <div class="space-y-4 border-t border-gray-200 px-4 py-4 dark:border-dark-600">
                <div class="grid gap-3 md:grid-cols-2">
                  <div class="rounded-lg border border-gray-200 bg-white p-3 dark:border-dark-600 dark:bg-dark-800">
                    <p class="text-xs font-semibold uppercase tracking-wide text-gray-400 dark:text-gray-500">
                      {{ t('services.guide.manualConfig.macosTitle') }}
                    </p>
                    <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
                      {{ t('services.guide.manualConfig.macosGuide') }}
                    </p>
                  </div>
                  <div class="rounded-lg border border-gray-200 bg-white p-3 dark:border-dark-600 dark:bg-dark-800">
                    <p class="text-xs font-semibold uppercase tracking-wide text-gray-400 dark:text-gray-500">
                      {{ t('services.guide.manualConfig.windowsTitle') }}
                    </p>
                    <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
                      {{ t('services.guide.manualConfig.windowsGuide') }}
                    </p>
                  </div>
                </div>

                <div
                  v-for="file in currentStepStep.manualConfigFiles"
                  :key="file.path"
                  class="overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800"
                >
                  <div class="border-b border-gray-200 px-3 py-2 dark:border-dark-600">
                    <p class="font-mono text-xs text-gray-500 dark:text-gray-400">{{ file.path }}</p>
                    <p v-if="file.hint" class="mt-1 text-xs text-amber-600 dark:text-amber-400">{{ file.hint }}</p>
                  </div>
                  <pre class="overflow-x-auto p-3 text-xs leading-6 text-gray-800 dark:text-gray-200"><code>{{ file.content }}</code></pre>
                </div>

                <ol class="list-decimal space-y-2 pl-5 text-sm leading-6 text-gray-600 dark:text-gray-300">
                  <li>{{ t('services.guide.manualConfig.stepCreateDir') }}</li>
                  <li>{{ t('services.guide.manualConfig.stepWriteFiles') }}</li>
                  <li>{{ t('services.guide.manualConfig.stepStart') }}</li>
                </ol>
              </div>
            </details>

            <div
              v-if="currentStepStep.config"
              class="space-y-2 rounded-xl border border-gray-200 bg-gray-50 p-3 dark:border-dark-600 dark:bg-dark-900"
            >
              <p class="text-xs font-semibold uppercase tracking-wide text-gray-400 dark:text-gray-500">
                {{ t('services.guide.configTitle') }}
              </p>
              <pre class="overflow-x-auto text-xs leading-6 text-gray-800 dark:text-gray-200"><code>{{ currentStepStep.config }}</code></pre>
            </div>
          </div>
        </div>
      </section>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import codexIcon from '@/assets/client-icons/codex-color.png'
import cursorIcon from '@/assets/client-icons/cursor.png'
import openCodeIcon from '@/assets/client-icons/opencode.png'
import qoderIcon from '@/assets/client-icons/qoder.png'
import codexDownloadImage from '@/assets/service-guide/codex_download.png'
import ccSwitchOpenImage from '@/assets/service-guide/ccswitch_open.png'
import ccSwitchImportImage from '@/assets/service-guide/ccswitch_import.png'
import ccSwitchEnableImage from '@/assets/service-guide/ccswitch_enable.png'
import {
  buildCcSwitchImportDeeplink,
  OPENAI_CC_SWITCH_CODEX_MODEL,
} from '@/utils/ccswitchImport'
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
  config?: string
  links?: Array<{ label: string; href: string }>
  primaryAction?: boolean
  heroImage?: string
  heroImageAlt?: string
  ccSwitchDownload?: boolean
  manualConfigFiles?: Array<{ path: string; content: string; hint?: string }>
}

interface ClientOption {
  id: string
  label: string
  description: string
  installHint: string
  icon: string
  steps: ClientStep[]
}

type Phase = 'select' | 'setup'

const props = defineProps<Props>()
const emit = defineEmits<Emits>()
const { t } = useI18n()

const phase = ref<Phase>('select')
const selectedClientId = ref('codex')
const currentStep = ref(0)
const ccSwitchImportModel = OPENAI_CC_SWITCH_CODEX_MODEL

watch(
  () => props.show,
  (visible) => {
    if (visible) {
      phase.value = 'select'
      selectedClientId.value = 'codex'
      currentStep.value = 0
    }
  },
  { immediate: true }
)

const title = computed(() => t('services.guide.title'))
const dialogTitle = computed(() => props.group ? `${props.group.name} · ${title.value}` : title.value)
const description = computed(() => t('services.guide.description'))
const baseRoot = computed(() => (props.baseUrl || window.location.origin).replace(/\/v1\/?$/, '').replace(/\/+$/, ''))
const openAIBase = computed(() => `${baseRoot.value}/v1`)
const anthropicBase = computed(() => baseRoot.value)
const geminiBase = computed(() => `${baseRoot.value}/v1beta`)
const antigravityBase = computed(() => `${baseRoot.value}/antigravity/v1`)
const ccSwitchDownloadLinks = [
  {
    os: 'windows',
    label: 'Windows',
    osLabel: 'Windows',
    href: 'https://gh-proxy.org/https://github.com/farion1231/cc-switch/releases/download/v3.16.2/CC-Switch-v3.16.2-Windows.msi',
  },
  {
    os: 'macos',
    label: 'macOS',
    osLabel: 'macOS',
    href: 'https://gh-proxy.org/https://github.com/farion1231/cc-switch/releases/download/v3.16.2/CC-Switch-v3.16.2-macOS.dmg',
  },
  {
    os: 'linux',
    label: 'Linux',
    osLabel: 'Linux',
    href: 'https://ccswitch.io/zh/',
  },
]
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
const detectedOs = computed(() => {
  if (typeof navigator === 'undefined') return 'macos'
  const platform = `${navigator.userAgent || ''} ${navigator.platform || ''}`.toLowerCase()
  if (platform.includes('win')) return 'windows'
  if (platform.includes('mac')) return 'macos'
  if (platform.includes('linux')) return 'linux'
  return 'unknown'
})
const detectedCcSwitchDownload = computed(() => {
  const download = ccSwitchDownloadLinks.find((link) => link.os === detectedOs.value) || ccSwitchDownloadLinks[2]
  return {
    ...download,
    label: t('services.guide.ccSwitchDownload.button', { os: download.label }),
    osLabel: detectedOs.value === 'unknown' ? t('services.guide.ccSwitchDownload.unknownOs') : download.osLabel,
  }
})
const showCcSwitchAlternativeLinks = computed(() => detectedOs.value === 'unknown')
const ccSwitchGuideImages = computed(() => [
  {
    title: t('services.guide.ccSwitchImport.guide.open.title'),
    description: t('services.guide.ccSwitchImport.guide.open.description'),
    alt: t('services.guide.ccSwitchImport.guide.open.alt'),
    src: ccSwitchOpenImage,
  },
  {
    title: t('services.guide.ccSwitchImport.guide.import.title'),
    description: t('services.guide.ccSwitchImport.guide.import.description'),
    alt: t('services.guide.ccSwitchImport.guide.import.alt'),
    src: ccSwitchImportImage,
  },
  {
    title: t('services.guide.ccSwitchImport.guide.enable.title'),
    description: t('services.guide.ccSwitchImport.guide.enable.description'),
    alt: t('services.guide.ccSwitchImport.guide.enable.alt'),
    src: ccSwitchEnableImage,
  },
])

const clients = computed<ClientOption[]>(() => [
  {
    id: 'codex',
    label: 'Codex',
    description: t('services.guide.clients.codex.description'),
    installHint: t('services.guide.clients.codex.installHint'),
    icon: codexIcon,
    steps: buildCodexSteps(),
  },
  {
    id: 'cursor',
    label: 'Cursor',
    description: t('services.guide.clients.cursor.description'),
    installHint: t('services.guide.clients.cursor.installHint'),
    icon: cursorIcon,
    steps: buildCursorSteps(),
  },
  {
    id: 'opencode',
    label: 'OpenCode',
    description: t('services.guide.clients.opencode.description'),
    installHint: t('services.guide.clients.opencode.installHint'),
    icon: openCodeIcon,
    steps: buildOpenCodeSteps(),
  },
  {
    id: 'qoder',
    label: 'Qoder',
    description: t('services.guide.clients.qoder.description'),
    installHint: t('services.guide.clients.qoder.installHint'),
    icon: qoderIcon,
    steps: buildQoderSteps(),
  },
])

const selectedClient = computed(() => clients.value.find((client) => client.id === selectedClientId.value) || clients.value[0])
const currentStepStep = computed(() => selectedClient.value.steps[Math.min(currentStep.value, selectedClient.value.steps.length - 1)])

function beginSetup(clientId: string) {
  selectedClientId.value = clientId
  currentStep.value = 0
  phase.value = 'setup'
}

function backToSelect() {
  phase.value = 'select'
}

function importCurrentServiceToCcSwitch() {
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

  const deeplink = buildCcSwitchImportDeeplink({
    baseUrl: baseRoot.value,
    platform: 'openai',
    clientType: 'claude',
    providerName: props.group?.name || t('services.guide.ccSwitchImport.defaultProviderName'),
    apiKey: props.apiKey,
    usageScript,
  })

  try {
    window.open(deeplink, '_self')
  } catch {
    window.alert(t('services.guide.ccSwitchImport.openFailed'))
  }
}

function buildCodexSteps(): ClientStep[] {
  const manualConfigFiles = buildCodexManualConfigFiles()

  return [
    {
      title: t('services.guide.stepTitles.download'),
      summary: t('services.guide.stepSummaries.codex.download'),
      description: t('services.guide.clients.codex.steps.download'),
      links: [{ label: t('services.guide.openDownloadPage'), href: 'https://openai.com/codex' }],
      primaryAction: true,
      heroImage: codexDownloadImage,
      heroImageAlt: t('services.guide.clients.codex.downloadImageAlt'),
    },
    {
      title: t('services.guide.stepTitles.configure'),
      summary: t('services.guide.stepSummaries.codex.configure'),
      description: t('services.guide.clients.codex.steps.configure'),
      ccSwitchDownload: true,
      manualConfigFiles,
    },
    {
      title: t('services.guide.stepTitles.start'),
      summary: t('services.guide.stepSummaries.start'),
      description: t('services.guide.clients.codex.steps.start'),
    },
  ]
}

function buildCodexManualConfigFiles() {
  return [
    {
      path: '~/.codex/config.toml / %userprofile%\\.codex\\config.toml',
      hint: t('services.guide.manualConfig.configTomlHint'),
      content: `model_provider = "OpenAI"
model = "gpt-5.5"
review_model = "gpt-5.5"
model_reasoning_effort = "xhigh"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true

[model_providers.OpenAI]
name = "OpenAI"
base_url = "${openAIBase.value}"
wire_api = "responses"
requires_openai_auth = true

[features]
goals = true`,
    },
    {
      path: '~/.codex/auth.json / %userprofile%\\.codex\\auth.json',
      content: `{
  "OPENAI_API_KEY": "${props.apiKey}"
}`,
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
</script>
