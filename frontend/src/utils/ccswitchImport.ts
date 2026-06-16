import type { GroupPlatform } from '@/types'

export const OPENAI_CC_SWITCH_CODEX_MODEL = 'gpt-5.5'

export type CcSwitchClientType = 'claude' | 'gemini' | 'opencode'

export interface CcSwitchImportConfig {
  app: string
  endpoint: string
  model?: string
}

export interface CcSwitchImportDeeplinkInput {
  baseUrl: string
  platform?: GroupPlatform | null
  clientType: CcSwitchClientType
  providerName: string
  apiKey: string
  usageScript: string
}

const ccSwitchProviderTranslations: Array<[RegExp, string]> = [
  [/按量付费/g, '-pay-as-you-go-'],
  [/套餐订阅/g, '-subscription-plan-'],
  [/x刀\/?天/gi, '-x-dollar-per-day-'],
  [/(\d+(?:\.\d+)?)刀\/?天/g, '-$1-dollar-per-day-'],
  [/(\d+(?:\.\d+)?)刀/g, '-$1-dollar-'],
  [/按量/g, '-usage-based-'],
  [/订阅/g, '-subscription-'],
  [/套餐/g, '-plan-'],
  [/免费/g, '-free-'],
  [/专业/g, '-pro-'],
  [/高级/g, '-pro-'],
  [/基础/g, '-basic-'],
  [/标准/g, '-standard-'],
  [/极速/g, '-fast-'],
  [/月/g, '-month-'],
  [/日/g, '-day-'],
]

export function normalizeCcSwitchProviderId(value: string, fallback = 'sub2api'): string {
  const translated = ccSwitchProviderTranslations.reduce(
    (current, [pattern, replacement]) => current.replace(pattern, replacement),
    value.trim()
  )

  const normalized = translated
    .trim()
    .toLowerCase()
    .replace(/[\s_]+/g, '-')
    .replace(/[^a-z0-9-]/g, '')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')

  return normalized || fallback
}

export function resolveCcSwitchImportConfig(
  platform: GroupPlatform | undefined | null,
  clientType: CcSwitchClientType,
  baseUrl: string
): CcSwitchImportConfig {
  switch (platform || 'anthropic') {
    case 'antigravity':
      return {
        app: clientType === 'gemini' ? 'gemini' : 'claude',
        endpoint: `${baseUrl}/antigravity`
      }
    case 'openai':
      if (clientType === 'opencode') {
        return {
          app: 'opencode',
          endpoint: baseUrl
        }
      }
      return {
        app: 'codex',
        endpoint: baseUrl,
        model: OPENAI_CC_SWITCH_CODEX_MODEL
      }
    case 'gemini':
      return {
        app: 'gemini',
        endpoint: baseUrl
      }
    default:
      return {
        app: 'claude',
        endpoint: baseUrl
      }
  }
}

export function buildCcSwitchImportDeeplink(input: CcSwitchImportDeeplinkInput): string {
  const config = resolveCcSwitchImportConfig(input.platform, input.clientType, input.baseUrl)
  const entries: [string, string][] = [
    ['resource', 'provider'],
    ['app', config.app],
    ['name', input.providerName],
    ['homepage', input.baseUrl],
    ['endpoint', config.endpoint],
    ['apiKey', input.apiKey],
    ['configFormat', 'json'],
    ['usageEnabled', 'true'],
    ['usageScript', btoa(input.usageScript)],
    ['usageAutoInterval', '30']
  ]

  if (config.model) {
    entries.splice(2, 0, ['model', config.model])
  }

  return `ccswitch://v1/import?${new URLSearchParams(entries).toString()}`
}
