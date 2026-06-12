<template>
  <BaseDialog :show="show" :title="dialogTitle" width="full" @close="emit('close')">
    <div class="space-y-8">
      <section class="space-y-3">
        <div class="flex items-center gap-4">
          <div class="flex h-12 w-12 items-center justify-center rounded-2xl border border-gray-200 bg-gradient-to-br from-white to-gray-100 shadow-sm dark:border-dark-600 dark:from-dark-700 dark:to-dark-800">
            <PlatformIcon :platform="group?.platform || 'anthropic'" size="lg" />
          </div>
          <div class="min-w-0">
            <h2 class="text-xl font-semibold tracking-tight text-gray-950 dark:text-white">
              {{ title }}
            </h2>
            <p class="text-sm leading-6 text-gray-500 dark:text-gray-400">
              {{ description }}
            </p>
          </div>
        </div>
      </section>

      <section v-if="phase === 'select'" class="space-y-4">
        <div class="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h3 class="text-base font-semibold text-gray-950 dark:text-white">
              {{ t('services.guide.chooseClientTitle') }}
            </h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              {{ t('services.guide.chooseClientHint') }}
            </p>
          </div>
        </div>

        <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <article
            v-for="client in clients"
            :key="client.id"
            class="group flex min-h-[216px] cursor-pointer flex-col gap-4 rounded-[28px] border border-gray-200 bg-white p-5 shadow-sm transition duration-200 hover:-translate-y-0.5 hover:border-primary-300 hover:shadow-lg hover:shadow-primary-100/50 focus:outline-none focus:ring-2 focus:ring-primary-500/40 dark:border-dark-600 dark:bg-dark-800 dark:hover:border-primary-500/50 dark:hover:shadow-none"
            role="button"
            tabindex="0"
            @click="beginSetup(client.id)"
            @keydown.enter.prevent="beginSetup(client.id)"
            @keydown.space.prevent="beginSetup(client.id)"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="flex h-14 w-14 items-center justify-center overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-600 dark:bg-dark-900">
                <img :src="client.icon" :alt="client.label" class="h-full w-full object-contain p-1.5" />
              </div>
              <span class="rounded-full bg-gray-100 px-3 py-1 text-xs font-medium text-gray-500 dark:bg-dark-700 dark:text-dark-300">
                {{ client.installHint }}
              </span>
            </div>

            <div class="space-y-2">
              <h4 class="text-lg font-semibold text-gray-950 dark:text-white">
                {{ client.label }}
              </h4>
              <p class="text-sm leading-6 text-gray-500 dark:text-gray-400">
                {{ client.description }}
              </p>
            </div>

            <div class="mt-auto pt-2">
              <button class="btn btn-primary w-full justify-center">
                {{ t('services.guide.configure') }}
              </button>
            </div>
          </article>
        </div>
      </section>

      <section v-else-if="selectedClient" class="space-y-5">
        <div>
          <button class="btn btn-secondary btn-sm" @click="backToSelect">
            <Icon name="arrowLeft" size="sm" />
            {{ t('common.back') }}
          </button>
        </div>

        <div class="grid gap-5 xl:grid-cols-[240px_minmax(0,1fr)]">
          <aside class="space-y-4 rounded-[28px] border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="flex items-center gap-3">
              <div class="flex h-11 w-11 items-center justify-center overflow-hidden rounded-2xl border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-900">
                <img :src="selectedClient.icon" :alt="selectedClient.label" class="h-full w-full object-contain p-1.5" />
              </div>
              <div class="min-w-0">
                <p class="text-sm font-semibold text-gray-950 dark:text-white">
                  {{ selectedClient.label }}
                </p>
                <p class="text-xs leading-5 text-gray-500 dark:text-gray-400">
                  {{ selectedClient.description }}
                </p>
              </div>
            </div>

            <div class="rounded-2xl bg-gray-50 p-3 dark:bg-dark-900">
              <p class="text-xs font-semibold uppercase tracking-[0.2em] text-gray-400 dark:text-gray-500">
                {{ t('services.guide.progressTitle') }}
              </p>
              <div class="mt-3 space-y-2">
                <div
                  v-for="(step, index) in selectedClient.steps"
                  :key="step.id"
                  class="rounded-2xl border px-3 py-3 transition"
                  :class="stepCardClass(index)"
                >
                  <div class="flex items-start gap-3">
                    <span
                      class="mt-0.5 flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full text-xs font-semibold"
                      :class="stepIndexClass(index)"
                    >
                      {{ index + 1 }}
                    </span>
                    <div class="min-w-0">
                      <p class="text-sm font-semibold">
                        {{ step.title }}
                      </p>
                      <p class="mt-1 text-xs leading-5 opacity-80">
                        {{ step.summary }}
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </aside>

          <div class="overflow-hidden rounded-[32px] border border-gray-200 bg-white shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="space-y-6 p-6 md:p-8">
              <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                <div class="space-y-3">
                  <p class="text-xs font-semibold uppercase tracking-[0.22em] text-gray-400 dark:text-gray-500">
                    {{ t('services.guide.currentStep', { index: currentStep + 1, total: selectedClient.steps.length }) }}
                  </p>
                  <div class="space-y-2">
                    <div class="flex flex-wrap items-center gap-2">
                      <h3 class="text-2xl font-semibold tracking-tight text-gray-950 dark:text-white">
                        {{ currentStepStep.title }}
                      </h3>
                      <span
                        v-if="currentStepBadge"
                        class="rounded-full border border-primary-200 bg-primary-50 px-3 py-1 text-xs font-medium text-primary-700 dark:border-primary-500/30 dark:bg-primary-500/10 dark:text-primary-200"
                      >
                        {{ currentStepBadge }}
                      </span>
                    </div>
                    <p class="max-w-3xl text-sm leading-7 text-gray-600 dark:text-gray-300">
                      {{ currentStepStep.description }}
                    </p>
                  </div>
                </div>
              </div>

              <div
                v-if="currentStepStep.links?.length && currentStepStep.centerLinks"
                class="flex justify-center pt-2"
              >
                <div class="flex w-full max-w-xl flex-col items-center gap-3 sm:flex-row sm:justify-center">
                  <a
                    v-for="link in currentStepStep.links"
                    :key="link.label"
                    class="btn btn-primary btn-lg w-full justify-center sm:w-auto"
                    :href="link.href"
                    target="_blank"
                    rel="noreferrer"
                  >
                    <Icon :name="resolveLinkIcon(link)" size="sm" />
                    {{ link.label }}
                  </a>
                </div>
              </div>

              <div
                v-if="currentStepStep.heroImage"
                class="overflow-hidden rounded-[28px] border border-gray-200 bg-gray-50 shadow-inner dark:border-dark-600 dark:bg-dark-900"
              >
                <img
                  :src="currentStepStep.heroImage"
                  :alt="currentStepStep.heroImageAlt || currentStepStep.title"
                  class="block w-full object-cover"
                />
              </div>

              <div
                v-if="currentStepStep.ccSwitchDownload"
                class="rounded-[28px] border border-primary-200 bg-gradient-to-br from-primary-50 via-white to-primary-50/70 p-5 dark:border-primary-500/30 dark:from-primary-500/10 dark:via-dark-800 dark:to-dark-800"
              >
                <div class="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
                  <div class="min-w-0 space-y-2">
                    <div class="flex items-center gap-2">
                      <Icon name="download" size="sm" class="text-primary-600 dark:text-primary-300" />
                      <h4 class="text-base font-semibold text-primary-900 dark:text-primary-100">
                        {{ t('services.guide.ccSwitchDownload.title') }}
                      </h4>
                    </div>
                    <p class="text-sm leading-7 text-primary-700 dark:text-primary-200">
                      {{ t('services.guide.ccSwitchDownload.description') }}
                    </p>
                    <p class="text-xs text-primary-600/80 dark:text-primary-200/80">
                      {{ t('services.guide.ccSwitchDownload.detected', { os: detectedCcSwitchDownload.osLabel }) }}
                    </p>
                  </div>

                  <div class="flex w-full justify-center lg:w-auto lg:justify-end">
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
                </div>

                <div
                  v-if="showCcSwitchAlternativeLinks"
                  class="mt-4 flex flex-wrap gap-2 border-t border-primary-200/70 pt-4 dark:border-primary-400/20"
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
              </div>

              <div
                v-if="currentStepStep.ccSwitchImport"
                class="space-y-5 rounded-[28px] border border-primary-200 bg-primary-50/70 p-5 dark:border-primary-500/30 dark:bg-primary-500/10"
              >
                <div class="flex items-center gap-2">
                  <Icon name="link" size="sm" class="text-primary-600 dark:text-primary-300" />
                  <h4 class="text-base font-semibold text-primary-900 dark:text-primary-100">
                    {{ t('services.guide.ccSwitchImport.title') }}
                  </h4>
                </div>
                <p class="text-sm leading-7 text-primary-800 dark:text-primary-100/90">
                  {{ t('services.guide.ccSwitchImport.description', { model: ccSwitchImportModel }) }}
                </p>

                <div class="grid gap-3 md:grid-cols-3">
                  <article
                    v-for="(guideImage, index) in ccSwitchGuideImages"
                    :key="guideImage.title"
                    class="overflow-hidden rounded-2xl border border-white/80 bg-white shadow-sm dark:border-primary-400/20 dark:bg-dark-800"
                  >
                    <div class="flex items-start gap-2 border-b border-gray-100 px-3 py-3 dark:border-dark-600">
                      <span class="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-primary-600 text-xs font-semibold text-white">
                        {{ index + 1 }}
                      </span>
                      <div class="min-w-0">
                        <h5 class="text-sm font-semibold text-gray-900 dark:text-white">{{ guideImage.title }}</h5>
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
                v-if="currentStepStep.callout"
                class="rounded-[28px] border p-5"
                :class="currentStepStep.callout.tone === 'warning'
                  ? 'border-amber-200 bg-amber-50 dark:border-amber-500/30 dark:bg-amber-500/10'
                  : 'border-emerald-200 bg-emerald-50 dark:border-emerald-500/30 dark:bg-emerald-500/10'"
              >
                <div class="flex items-start gap-3">
                  <div
                    class="mt-0.5 flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full"
                    :class="currentStepStep.callout.tone === 'warning'
                      ? 'bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300'
                      : 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300'"
                  >
                    <Icon :name="currentStepStep.callout.tone === 'warning' ? 'exclamationTriangle' : 'checkCircle'" size="sm" />
                  </div>
                  <div class="min-w-0">
                    <h4
                      class="text-base font-semibold"
                      :class="currentStepStep.callout.tone === 'warning'
                        ? 'text-amber-900 dark:text-amber-100'
                        : 'text-emerald-900 dark:text-emerald-100'"
                    >
                      {{ currentStepStep.callout.title }}
                    </h4>
                    <p
                      class="mt-2 text-sm leading-7"
                      :class="currentStepStep.callout.tone === 'warning'
                        ? 'text-amber-800 dark:text-amber-100/90'
                        : 'text-emerald-800 dark:text-emerald-100/90'"
                    >
                      {{ currentStepStep.callout.description }}
                    </p>
                  </div>
                </div>
              </div>

              <div
                v-if="currentStepStep.checklist?.length"
                class="rounded-[28px] border border-gray-200 bg-gray-50 p-5 dark:border-dark-600 dark:bg-dark-900"
              >
                <h4 class="text-base font-semibold text-gray-950 dark:text-white">
                  {{ t('services.guide.stepChecklistTitle') }}
                </h4>
                <ol class="mt-4 space-y-3">
                  <li
                    v-for="(item, index) in currentStepStep.checklist"
                    :key="item"
                    class="flex items-start gap-3"
                  >
                    <span class="flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full bg-primary-600 text-xs font-semibold text-white">
                      {{ index + 1 }}
                    </span>
                    <p class="pt-0.5 text-sm leading-7 text-gray-700 dark:text-gray-200">
                      {{ item }}
                    </p>
                  </li>
                </ol>
              </div>

              <div
                v-if="currentStepStep.gallery?.length"
                class="grid gap-4 lg:grid-cols-2"
              >
                <article
                  v-for="image in currentStepStep.gallery"
                  :key="image.title"
                  class="overflow-hidden rounded-[28px] border border-gray-200 bg-white shadow-sm dark:border-dark-600 dark:bg-dark-800"
                >
                  <div class="space-y-2 border-b border-gray-100 px-5 py-4 dark:border-dark-600">
                    <h4 class="text-base font-semibold text-gray-950 dark:text-white">
                      {{ image.title }}
                    </h4>
                    <p class="text-sm leading-6 text-gray-500 dark:text-gray-400">
                      {{ image.description }}
                    </p>
                  </div>
                  <img
                    :src="image.src"
                    :alt="image.alt"
                    class="block w-full bg-gray-50 object-contain dark:bg-dark-900"
                  />
                </article>
              </div>

              <div
                v-if="currentStepStep.links?.length && !currentStepStep.centerLinks"
                class="flex flex-wrap gap-3"
              >
                <a
                  v-for="link in currentStepStep.links"
                  :key="link.label"
                  class="btn btn-secondary"
                  :href="link.href"
                  target="_blank"
                  rel="noreferrer"
                >
                  <Icon :name="resolveLinkIcon(link)" size="sm" />
                  {{ link.label }}
                </a>
              </div>

              <div
                v-if="currentStepStep.manualConfigFiles?.length"
                class="space-y-5 rounded-[28px] border border-gray-200 bg-gray-50 p-5 dark:border-dark-600 dark:bg-dark-900"
              >
                <div class="space-y-4">
                  <div class="flex items-center gap-2">
                    <Icon name="link" size="sm" class="text-gray-500 dark:text-gray-400" />
                    <h4 class="text-base font-semibold text-gray-950 dark:text-white">
                      {{ t('services.guide.manualConfig.summary') }}
                    </h4>
                  </div>

                  <div class="grid gap-3 md:grid-cols-2">
                    <div class="rounded-2xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
                      <p class="text-xs font-semibold uppercase tracking-wide text-gray-400 dark:text-gray-500">
                        {{ t('services.guide.manualConfig.macosTitle') }}
                      </p>
                      <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
                        {{ t('services.guide.manualConfig.macosGuide') }}
                      </p>
                    </div>
                    <div class="rounded-2xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
                      <p class="text-xs font-semibold uppercase tracking-wide text-gray-400 dark:text-gray-500">
                        {{ t('services.guide.manualConfig.windowsTitle') }}
                      </p>
                      <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
                        {{ t('services.guide.manualConfig.windowsGuide') }}
                      </p>
                    </div>
                  </div>
                </div>

                <div
                  v-for="file in currentStepStep.manualConfigFiles"
                  :key="file.path"
                  class="overflow-hidden rounded-2xl border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800"
                >
                  <div class="border-b border-gray-200 px-4 py-3 dark:border-dark-600">
                    <p class="font-mono text-xs text-gray-500 dark:text-gray-400">{{ file.path }}</p>
                    <p v-if="file.hint" class="mt-1 text-xs text-amber-600 dark:text-amber-400">{{ file.hint }}</p>
                  </div>
                  <pre class="overflow-x-auto p-4 text-xs leading-6 text-gray-800 dark:text-gray-200"><code>{{ file.content }}</code></pre>
                </div>

                <ol class="list-decimal space-y-2 pl-5 text-sm leading-6 text-gray-600 dark:text-gray-300">
                  <li>{{ t('services.guide.manualConfig.stepCreateDir') }}</li>
                  <li>{{ t('services.guide.manualConfig.stepWriteFiles') }}</li>
                  <li>{{ t('services.guide.manualConfig.stepStart') }}</li>
                </ol>
              </div>

              <div
                v-if="currentStepStep.config"
                class="space-y-2 rounded-[28px] border border-gray-200 bg-gray-50 p-5 dark:border-dark-600 dark:bg-dark-900"
              >
                <p class="text-xs font-semibold uppercase tracking-wide text-gray-400 dark:text-gray-500">
                  {{ t('services.guide.configTitle') }}
                </p>
                <pre class="overflow-x-auto text-xs leading-6 text-gray-800 dark:text-gray-200"><code>{{ currentStepStep.config }}</code></pre>
              </div>
            </div>

            <div class="border-t border-gray-200 bg-gray-50/80 p-4 dark:border-dark-600 dark:bg-dark-900/80 md:p-6">
              <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                <div class="flex flex-col gap-3 sm:flex-row">
                  <template v-for="action in footerActions.left" :key="action.id">
                    <a
                      v-if="action.href"
                      class="btn"
                      :class="action.variant === 'primary' ? 'btn-primary' : 'btn-secondary'"
                      :href="action.href"
                      target="_blank"
                      rel="noreferrer"
                    >
                      <Icon v-if="action.icon" :name="action.icon" size="sm" />
                      {{ action.label }}
                    </a>
                    <button
                      v-else
                      class="btn"
                      :class="action.variant === 'primary' ? 'btn-primary' : 'btn-secondary'"
                      @click="action.onClick?.()"
                    >
                      <Icon v-if="action.icon" :name="action.icon" size="sm" />
                      {{ action.label }}
                    </button>
                  </template>
                </div>

                <div class="flex flex-col gap-3 sm:flex-row sm:justify-end">
                  <template v-for="action in footerActions.right" :key="action.id">
                    <a
                      v-if="action.href"
                      class="btn"
                      :class="action.variant === 'primary' ? 'btn-primary' : 'btn-secondary'"
                      :href="action.href"
                      target="_blank"
                      rel="noreferrer"
                    >
                      <Icon v-if="action.icon" :name="action.icon" size="sm" />
                      {{ action.label }}
                    </a>
                    <button
                      v-else
                      class="btn"
                      :class="action.variant === 'primary' ? 'btn-primary' : 'btn-secondary'"
                      @click="action.onClick?.()"
                    >
                      <Icon v-if="action.icon" :name="action.icon" size="sm" />
                      {{ action.label }}
                    </button>
                  </template>
                </div>
              </div>
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
import restartCodexImage from '@/assets/service-guide/restart-codex.png'
import codexHiCheckImage from '@/assets/service-guide/codex-hi-check.png'
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

type GuideIcon = 'download' | 'externalLink' | 'link'

interface StepLink {
  label: string
  href: string
  icon?: GuideIcon
}

interface ClientStep {
  id: string
  title: string
  summary: string
  description: string
  config?: string
  links?: StepLink[]
  centerLinks?: boolean
  heroImage?: string
  heroImageAlt?: string
  ccSwitchDownload?: boolean
  ccSwitchImport?: boolean
  checklist?: string[]
  callout?: {
    tone: 'warning' | 'success'
    title: string
    description: string
  }
  gallery?: Array<{
    title: string
    description: string
    src: string
    alt: string
  }>
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

interface FooterAction {
  id: string
  label: string
  variant: 'primary' | 'secondary'
  icon?: GuideIcon
  href?: string
  onClick?: () => void
}

type Phase = 'select' | 'setup'
type CodexSetupMode = 'cc-switch' | 'manual' | null

const props = defineProps<Props>()
const emit = defineEmits<Emits>()
const { t } = useI18n()

const phase = ref<Phase>('select')
const selectedClientId = ref('codex')
const currentStep = ref(0)
const codexSetupMode = ref<CodexSetupMode>(null)
const ccSwitchImportModel = OPENAI_CC_SWITCH_CODEX_MODEL

watch(
  () => props.show,
  (visible) => {
    if (visible) {
      resetGuide()
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
] as const

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
const currentStepStep = computed(() => {
  const steps = selectedClient.value?.steps || []
  const safeIndex = Math.min(currentStep.value, Math.max(steps.length - 1, 0))
  return steps[safeIndex]
})

const currentStepBadge = computed(() => {
  if (selectedClientId.value !== 'codex' || currentStepStep.value?.id !== 'configureService') {
    return ''
  }

  if (codexSetupMode.value === 'manual') {
    return t('services.guide.modeBadges.manual')
  }

  if (codexSetupMode.value === 'cc-switch') {
    return t('services.guide.modeBadges.ccSwitch')
  }

  return ''
})

const footerActions = computed(() => {
  if (!selectedClient.value || !currentStepStep.value) {
    return { left: [] as FooterAction[], right: [] as FooterAction[] }
  }

  if (selectedClientId.value === 'codex') {
    return buildCodexFooterActions(currentStepStep.value.id)
  }

  return buildGenericFooterActions()
})

function resetGuide() {
  phase.value = 'select'
  selectedClientId.value = 'codex'
  currentStep.value = 0
  codexSetupMode.value = null
}

function beginSetup(clientId: string) {
  selectedClientId.value = clientId
  currentStep.value = 0
  codexSetupMode.value = null
  phase.value = 'setup'
}

function backToSelect() {
  resetGuide()
}

function goToNextStep() {
  const steps = selectedClient.value?.steps || []
  currentStep.value = Math.min(currentStep.value + 1, Math.max(steps.length - 1, 0))
}

function setCodexSetupMode(mode: Exclude<CodexSetupMode, null>) {
  codexSetupMode.value = mode
  currentStep.value = 2
}

function finishGuide() {
  emit('close')
}

function resolveLinkIcon(link: StepLink): GuideIcon {
  return link.icon || 'externalLink'
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

function stepCardClass(index: number) {
  if (index === currentStep.value) {
    return 'border-primary-200 bg-primary-50/80 text-primary-800 dark:border-primary-500/30 dark:bg-primary-500/10 dark:text-primary-100'
  }

  if (index < currentStep.value) {
    return 'border-emerald-200 bg-emerald-50/70 text-emerald-800 dark:border-emerald-500/30 dark:bg-emerald-500/10 dark:text-emerald-200'
  }

  return 'border-transparent bg-white text-gray-500 dark:bg-dark-800 dark:text-dark-300'
}

function stepIndexClass(index: number) {
  if (index === currentStep.value) {
    return 'bg-primary-600 text-white'
  }

  if (index < currentStep.value) {
    return 'bg-emerald-600 text-white'
  }

  return 'bg-gray-200 text-gray-500 dark:bg-dark-600 dark:text-dark-300'
}

function buildCodexFooterActions(stepId: string) {
  const left: FooterAction[] = []
  const right: FooterAction[] = []

  if (stepId === 'downloadCodex') {
    right.push({
      id: 'codex-next',
      label: t('services.guide.actions.downloadedCodexNext'),
      variant: 'primary',
      onClick: goToNextStep,
    })
    return { left, right }
  }

  if (stepId === 'downloadCcSwitch') {
    left.push({
      id: 'download-cc-switch',
      label: t('services.guide.actions.openCcSwitchDownload'),
      variant: 'secondary',
      icon: 'download',
      href: detectedCcSwitchDownload.value.href,
    })
    right.push({
      id: 'manual-config',
      label: t('services.guide.actions.manualInstead'),
      variant: 'secondary',
      onClick: () => setCodexSetupMode('manual'),
    })
    right.push({
      id: 'cc-switch-next',
      label: t('services.guide.actions.downloadedCcSwitchNext'),
      variant: 'primary',
      onClick: () => setCodexSetupMode('cc-switch'),
    })
    return { left, right }
  }

  if (stepId === 'configureService') {
    if (codexSetupMode.value === 'cc-switch') {
      left.push({
        id: 'import-cc-switch',
        label: t('services.guide.actions.importCcSwitch'),
        variant: 'secondary',
        icon: 'link',
        onClick: importCurrentServiceToCcSwitch,
      })
    }
    right.push({
      id: 'configure-next',
      label: t('services.guide.actions.configuredNext'),
      variant: 'primary',
      onClick: goToNextStep,
    })
    return { left, right }
  }

  right.push({
    id: 'finish-guide',
    label: t('services.guide.actions.finish'),
    variant: 'primary',
    onClick: finishGuide,
  })
  return { left, right }
}

function buildGenericFooterActions() {
  const left: FooterAction[] = []
  const right: FooterAction[] = []
  const step = currentStepStep.value
  const steps = selectedClient.value?.steps || []
  const isLastStep = currentStep.value >= steps.length - 1

  if (step?.links?.length) {
    step.links.forEach((link, index) => {
      left.push({
        id: `link-${index}`,
        label: link.label,
        variant: 'secondary',
        icon: link.icon || 'externalLink',
        href: link.href,
      })
    })
  }

  right.push({
    id: isLastStep ? 'finish' : 'next',
    label: isLastStep ? t('services.guide.actions.finish') : t('services.guide.actions.genericNext'),
    variant: 'primary',
    onClick: isLastStep ? finishGuide : goToNextStep,
  })

  return { left, right }
}

function buildCodexSteps(): ClientStep[] {
  const manualConfigFiles = buildCodexManualConfigFiles()

  return [
    {
      id: 'downloadCodex',
      title: t('services.guide.codexSteps.downloadCodex.title'),
      summary: t('services.guide.codexSteps.downloadCodex.summary'),
      description: t('services.guide.clients.codex.steps.download'),
      links: [{ label: t('services.guide.actions.openCodexDownload'), href: 'https://openai.com/codex', icon: 'download' }],
      centerLinks: true,
      heroImage: codexDownloadImage,
      heroImageAlt: t('services.guide.clients.codex.downloadImageAlt'),
    },
    {
      id: 'downloadCcSwitch',
      title: t('services.guide.codexSteps.downloadCcSwitch.title'),
      summary: t('services.guide.codexSteps.downloadCcSwitch.summary'),
      description: t('services.guide.clients.codex.steps.prepareCcSwitch'),
      ccSwitchDownload: true,
    },
    {
      id: 'configureService',
      title: codexSetupMode.value === 'manual'
        ? t('services.guide.codexSteps.manualConfigure.title')
        : t('services.guide.codexSteps.importService.title'),
      summary: codexSetupMode.value === 'manual'
        ? t('services.guide.codexSteps.manualConfigure.summary')
        : t('services.guide.codexSteps.importService.summary'),
      description: codexSetupMode.value === 'manual'
        ? t('services.guide.clients.codex.steps.manualConfigure')
        : t('services.guide.clients.codex.steps.importService'),
      ccSwitchImport: codexSetupMode.value === 'cc-switch',
      manualConfigFiles: codexSetupMode.value === 'manual' ? manualConfigFiles : undefined,
    },
    {
      id: 'start',
      title: t('services.guide.stepTitles.start'),
      summary: t('services.guide.stepSummaries.start'),
      description: t('services.guide.clients.codex.steps.start'),
      callout: {
        tone: 'warning',
        title: t('services.guide.clients.codex.restartNoticeTitle'),
        description: t('services.guide.clients.codex.restartNoticeDescription'),
      },
      checklist: [
        t('services.guide.clients.codex.startChecklist.quitApp'),
        t('services.guide.clients.codex.startChecklist.reopenApp'),
        t('services.guide.clients.codex.startChecklist.sendHi'),
      ],
      gallery: [
        {
          title: t('services.guide.clients.codex.startGallery.restart.title'),
          description: t('services.guide.clients.codex.startGallery.restart.description'),
          src: restartCodexImage,
          alt: t('services.guide.clients.codex.startGallery.restart.alt'),
        },
        {
          title: t('services.guide.clients.codex.startGallery.verify.title'),
          description: t('services.guide.clients.codex.startGallery.verify.description'),
          src: codexHiCheckImage,
          alt: t('services.guide.clients.codex.startGallery.verify.alt'),
        },
      ],
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
      id: 'download',
      title: t('services.guide.stepTitles.download'),
      summary: t('services.guide.stepSummaries.cursor.download'),
      description: t('services.guide.clients.cursor.steps.download'),
      config: 'Open Cursor Settings > Models > Add new API key',
      links: [{ label: t('services.guide.openDownloadPage'), href: 'https://cursor.com' }],
    },
    {
      id: 'configure',
      title: t('services.guide.stepTitles.configure'),
      summary: t('services.guide.stepSummaries.cursor.configure'),
      description: t('services.guide.clients.cursor.steps.configure'),
      config: `Provider: OpenAI compatible
Base URL: ${preferredBase.value}
API Key: ${props.apiKey}`,
    },
    {
      id: 'verify',
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
      id: 'download',
      title: t('services.guide.stepTitles.download'),
      summary: t('services.guide.stepSummaries.opencode.download'),
      description: t('services.guide.clients.opencode.steps.download'),
      config: 'npm install -g opencode',
      links: [{ label: t('services.guide.openDownloadPage'), href: 'https://opencode.ai' }],
    },
    {
      id: 'configure',
      title: t('services.guide.stepTitles.configure'),
      summary: t('services.guide.stepSummaries.opencode.configure'),
      description: t('services.guide.clients.opencode.steps.configure'),
      config: JSON.stringify({
        provider: {
          openai: {
            options: {
              baseURL: preferredBase.value,
              apiKey: props.apiKey,
            },
          },
        },
      }, null, 2),
    },
    {
      id: 'start',
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
      id: 'download',
      title: t('services.guide.stepTitles.download'),
      summary: t('services.guide.stepSummaries.qoder.download'),
      description: t('services.guide.clients.qoder.steps.download'),
      config: 'Install Qoder from the official download page.',
      links: [{ label: t('services.guide.openDownloadPage'), href: 'https://qoder.com' }],
    },
    {
      id: 'configure',
      title: t('services.guide.stepTitles.configure'),
      summary: t('services.guide.stepSummaries.qoder.configure'),
      description: t('services.guide.clients.qoder.steps.configure'),
      config: `Provider: OpenAI compatible
Base URL: ${preferredBase.value}
API Key: ${props.apiKey}`,
    },
    {
      id: 'verify',
      title: t('services.guide.stepTitles.verify'),
      summary: t('services.guide.stepSummaries.verify'),
      description: t('services.guide.clients.qoder.steps.verify'),
      config: 'Open a new project and run a small prompt test.',
    },
  ]
}
</script>
