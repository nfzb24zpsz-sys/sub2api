package service

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"text/template"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	BootstrapProviderID    = "erqishi"
	BootstrapProviderName  = "Erqishi"
	BootstrapAPIKeyEnv     = "ERQISHI_API_KEY"
	BootstrapCodexModel    = "gpt-5.4"
	BootstrapNPMRegistry   = "https://registry.npmmirror.com"
	BootstrapNodeMirror    = "https://npmmirror.com/mirrors/node"
	BootstrapTargetUnix    = "unix"
	BootstrapTargetWindows = "windows"
)

type BootstrapScriptRequest struct {
	GroupID *int64
	OS      string
	BaseURL string
}

type BootstrapScriptResult struct {
	Filename    string
	Content     string
	ContentType string
}

type bootstrapTemplateData struct {
	ProviderID   string
	ProviderName string
	APIKeyEnv    string
	Model        string
	BaseURL      string
	APIKey       string
	NPMRegistry  string
	NodeMirror   string
}

func (s *APIKeyService) GenerateBootstrapScript(ctx context.Context, userID int64, req BootstrapScriptRequest) (*BootstrapScriptResult, error) {
	key, err := s.getOrCreateBootstrapAPIKey(ctx, userID, req.GroupID)
	if err != nil {
		return nil, err
	}

	baseURL, err := normalizeBootstrapBaseURL(req.BaseURL)
	if err != nil {
		return nil, err
	}
	if key.Group != nil && key.Group.Platform == PlatformAntigravity {
		baseURL += "/antigravity"
	}

	targetOS := normalizeBootstrapTargetOS(req.OS)
	data := bootstrapTemplateData{
		ProviderID:   BootstrapProviderID,
		ProviderName: BootstrapProviderName,
		APIKeyEnv:    BootstrapAPIKeyEnv,
		Model:        BootstrapCodexModel,
		BaseURL:      baseURL,
		APIKey:       key.Key,
		NPMRegistry:  BootstrapNPMRegistry,
		NodeMirror:   BootstrapNodeMirror,
	}

	var tmpl string
	contentType := "text/x-shellscript; charset=utf-8"
	ext := ".sh"
	if targetOS == BootstrapTargetWindows {
		tmpl = windowsBootstrapTemplate
		contentType = "text/plain; charset=utf-8"
		ext = ".ps1"
	} else {
		tmpl = unixBootstrapTemplate
	}

	content, err := executeBootstrapTemplate(tmpl, data)
	if err != nil {
		return nil, fmt.Errorf("render bootstrap script: %w", err)
	}

	serviceName := "service"
	if key.Group != nil && strings.TrimSpace(key.Group.Name) != "" {
		serviceName = key.Group.Name
	}

	return &BootstrapScriptResult{
		Filename:    "erqishi-" + sanitizeBootstrapFilenamePart(serviceName) + "-setup" + ext,
		Content:     content,
		ContentType: contentType,
	}, nil
}

func (s *APIKeyService) getOrCreateBootstrapAPIKey(ctx context.Context, userID int64, groupID *int64) (*APIKey, error) {
	if groupID == nil || *groupID <= 0 {
		return nil, infraBadBootstrapRequest("group_id is required")
	}

	keys, _, err := s.List(ctx, userID, pagination.PaginationParams{
		Page:      1,
		PageSize:  100,
		SortBy:    "created_at",
		SortOrder: "asc",
	}, APIKeyListFilters{
		Status:  StatusActive,
		GroupID: groupID,
	})
	if err != nil {
		return nil, err
	}
	for i := range keys {
		if keys[i].GroupID != nil && *keys[i].GroupID == *groupID && keys[i].Status == StatusActive {
			if keys[i].Group == nil {
				if group, groupErr := s.groupRepo.GetByID(ctx, *groupID); groupErr == nil {
					keys[i].Group = group
				}
			}
			return &keys[i], nil
		}
	}

	group, err := s.groupRepo.GetByID(ctx, *groupID)
	if err != nil {
		return nil, fmt.Errorf("get group: %w", err)
	}
	name := strings.TrimSpace(group.Name)
	if name == "" {
		name = "Service"
	}
	key, err := s.Create(ctx, userID, CreateAPIKeyRequest{
		Name:    name + " 连接",
		GroupID: groupID,
	})
	if err != nil {
		return nil, err
	}
	key.Group = group
	return key, nil
}

func normalizeBootstrapTargetOS(os string) string {
	switch strings.ToLower(strings.TrimSpace(os)) {
	case BootstrapTargetWindows, "win", "win32", "win64":
		return BootstrapTargetWindows
	default:
		return BootstrapTargetUnix
	}
}

func normalizeBootstrapBaseURL(raw string) (string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(raw), "/")
	if baseURL == "" {
		return "", infraBadBootstrapRequest("base_url is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", infraBadBootstrapRequest("base_url must be a valid absolute URL")
	}
	switch parsed.Scheme {
	case "http", "https":
	default:
		return "", infraBadBootstrapRequest("base_url must use http or https")
	}
	return baseURL, nil
}

func sanitizeBootstrapFilenamePart(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "service"
	}
	return filepath.Base(out)
}

func executeBootstrapTemplate(raw string, data bootstrapTemplateData) (string, error) {
	tmpl, err := template.New("bootstrap").Funcs(template.FuncMap{
		"psq": powershellSingleQuote,
		"shq": shellSingleQuote,
	}).Parse(raw)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

func powershellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func infraBadBootstrapRequest(message string) error {
	return infraerrors.BadRequest("INVALID_BOOTSTRAP_REQUEST", message)
}

const unixBootstrapTemplate = `#!/usr/bin/env bash
set -euo pipefail

PROVIDER_ID={{ .ProviderID | shq }}
PROVIDER_NAME={{ .ProviderName | shq }}
MODEL={{ .Model | shq }}
BASE_URL={{ .BaseURL | shq }}
API_KEY={{ .APIKey | shq }}
NPM_REGISTRY={{ .NPMRegistry | shq }}
NODE_MIRROR={{ .NodeMirror | shq }}
CODEX_PACKAGE="@openai/codex"
USAGE_SCRIPT='({
  request: {
    url: "__BASE_URL__/v1/usage",
    method: "GET",
    headers: { "Authorization": "Bearer __API_KEY__" }
  },
  extractor: function(response) {
    const remaining = response && (response.remaining || (response.quota && response.quota.remaining) || response.balance);
    const unit = response && (response.unit || (response.quota && response.quota.unit)) || "USD";
    return {
      isValid: response && (response.is_active !== undefined ? response.is_active : (response.isValid !== undefined ? response.isValid : true)),
      remaining: remaining,
      unit: unit
    };
  }
})'

log() { printf '%s\n' "$1"; }
has_cmd() { command -v "$1" >/dev/null 2>&1; }

detect_shell_profile() {
  if [ -n "${ZDOTDIR:-}" ] && [ -f "$ZDOTDIR/.zshrc" ]; then
    printf '%s\n' "$ZDOTDIR/.zshrc"
    return
  fi
  if [ -n "${SHELL:-}" ]; then
    case "$SHELL" in
      */zsh) printf '%s\n' "$HOME/.zshrc"; return ;;
      */bash) printf '%s\n' "$HOME/.bashrc"; return ;;
    esac
  fi
  if [ -f "$HOME/.zshrc" ]; then printf '%s\n' "$HOME/.zshrc"; else printf '%s\n' "$HOME/.bashrc"; fi
}

ensure_node() {
  if has_cmd npm; then return; fi
  log "未检测到 npm，正在从国内镜像安装便携版 Node.js..."
  local os arch node_arch ext version archive url runtime_dir
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"
  version="v22.11.0"
  runtime_dir="$HOME/.erqishi/runtime"
  mkdir -p "$runtime_dir"
  case "$arch" in
    arm64|aarch64) node_arch="arm64" ;;
    x86_64|amd64) node_arch="x64" ;;
    *) log "暂不支持的 CPU 架构：$arch，请先手动安装 Node.js 后重试。"; exit 1 ;;
  esac
  case "$os" in
    darwin) ext="tar.gz" ;;
    linux) ext="tar.xz" ;;
    *) log "暂不支持的系统：$os，请先手动安装 Node.js 后重试。"; exit 1 ;;
  esac
  archive="$runtime_dir/node-$version-$os-$node_arch.$ext"
  url="$NODE_MIRROR/$version/node-$version-$os-$node_arch.$ext"
  if has_cmd curl; then
    curl -fL --retry 3 --connect-timeout 20 "$url" -o "$archive"
  elif has_cmd wget; then
    wget -O "$archive" "$url"
  else
    log "未检测到 curl/wget，无法下载 Node.js。"; exit 1
  fi
  rm -rf "$runtime_dir/node"
  mkdir -p "$runtime_dir/node"
  tar -xf "$archive" -C "$runtime_dir/node" --strip-components 1
  export PATH="$runtime_dir/node/bin:$PATH"
  if ! has_cmd npm; then log "Node.js 安装失败，请检查网络或手动安装 Node.js。"; exit 1; fi
}

ensure_codex() {
  if has_cmd codex && codex --version >/dev/null 2>&1; then
    log "已检测到 Codex：$(codex --version)"
    return
  fi
  ensure_node
  log "未检测到 Codex，正在通过国内 npm 镜像安装..."
  npm install -g "$CODEX_PACKAGE" --registry="$NPM_REGISTRY"
  if ! has_cmd codex || ! codex --version >/dev/null 2>&1; then
    log "Codex 安装失败，请检查 npm 全局 bin 目录是否在 PATH 中。"; exit 1
  fi
  log "Codex 安装完成：$(codex --version)"
}

write_env_file() {
  mkdir -p "$HOME/.codex"
  cat > "$HOME/.codex/$PROVIDER_ID.env" <<EOF_ENV
export {{ .APIKeyEnv }}="$API_KEY"
EOF_ENV
  local profile begin end source_line
  profile="$(detect_shell_profile)"
  begin="# >>> $PROVIDER_ID managed"
  end="# <<< $PROVIDER_ID managed"
  source_line="export PATH=\"\$HOME/.erqishi/runtime/node/bin:\$PATH\"
[ -f \"\$HOME/.codex/$PROVIDER_ID.env\" ] && . \"\$HOME/.codex/$PROVIDER_ID.env\""
  touch "$profile"
  BEGIN_MARKER="$begin" END_MARKER="$end" SOURCE_LINE="$source_line" PROFILE_PATH="$profile" node <<'NODE'
const fs = require('fs')
const path = process.env.PROFILE_PATH
const begin = process.env.BEGIN_MARKER
const end = process.env.END_MARKER
const block = begin + "\n" + process.env.SOURCE_LINE + "\n" + end
let text = fs.existsSync(path) ? fs.readFileSync(path, 'utf8') : ''
function escapeRegExp(value) { return value.replace(/[\\^$.*+?()[\]{}|]/g, '\\$&') }
const re = new RegExp(escapeRegExp(begin) + '[\\s\\S]*?' + escapeRegExp(end))
text = re.test(text) ? text.replace(re, block) : text.replace(/\s*$/, '') + (text.trim() ? "\n\n" : "") + block + "\n"
fs.writeFileSync(path, text)
NODE
}

configure_codex() {
  mkdir -p "$HOME/.codex"
  local config="$HOME/.codex/config.toml"
  [ -f "$config" ] || touch "$config"
  cp "$config" "$config.$(date +%Y%m%d%H%M%S).erqishi.bak"
  CONFIG_PATH="$config" PROVIDER_ID="$PROVIDER_ID" PROVIDER_NAME="$PROVIDER_NAME" MODEL="$MODEL" BASE_URL="$BASE_URL" node <<'NODE'
const fs = require('fs')
const path = process.env.CONFIG_PATH
const providerId = process.env.PROVIDER_ID
const providerName = process.env.PROVIDER_NAME
const model = process.env.MODEL
const baseUrl = process.env.BASE_URL
const begin = '# >>> erqishi managed'
const end = '# <<< erqishi managed'
const block = begin + "\n"
  + "[model_providers." + providerId + "]\n"
  + "name = " + JSON.stringify(providerName) + "\n"
  + "base_url = " + JSON.stringify(baseUrl) + "\n"
  + "env_key = " + JSON.stringify('{{ .APIKeyEnv }}') + "\n"
  + "wire_api = \"responses\"\n"
  + end
let text = fs.existsSync(path) ? fs.readFileSync(path, 'utf8') : ''
function upsertRootTomlString(text, key, value) {
  const line = key + " = " + JSON.stringify(value)
  const lines = text ? text.split(/\r?\n/) : []
  let rootEnd = lines.length
  for (let i = 0; i < lines.length; i++) {
    if (/^\s*\[/.test(lines[i])) { rootEnd = i; break }
  }
  const keyRe = new RegExp("^\\s*" + key.replace(/[\\^$.*+?()[\]{}|]/g, '\\$&') + "\\s*=.*$")
  for (let i = 0; i < rootEnd; i++) {
    if (keyRe.test(lines[i])) {
      lines[i] = line
      return lines.join("\n")
    }
  }
  lines.unshift(line)
  return lines.join("\n")
}
text = upsertRootTomlString(text, 'model_provider', providerId)
text = upsertRootTomlString(text, 'model', model)
function escapeRegExp(value) { return value.replace(/[\\^$.*+?()[\]{}|]/g, '\\$&') }
const re = new RegExp(escapeRegExp(begin) + '[\\s\\S]*?' + escapeRegExp(end))
text = re.test(text) ? text.replace(re, block) : text.replace(/\s*$/, '') + (text.trim() ? "\n\n" : "") + block + "\n"
fs.writeFileSync(path, text)
NODE
  log "Codex 已切换到 $PROVIDER_NAME 中转。"
}

configure_opencode_if_present() {
  if ! has_cmd opencode || ! opencode --version >/dev/null 2>&1; then log "未检测到 opencode，已跳过。"; return; fi
  local dir config
  dir="${XDG_CONFIG_HOME:-$HOME/.config}/opencode"
  config="$dir/opencode.json"
  mkdir -p "$dir"
  [ -f "$config" ] && cp "$config" "$config.$(date +%Y%m%d%H%M%S).erqishi.bak"
  OPENCODE_CONFIG="$config" PROVIDER_ID="$PROVIDER_ID" PROVIDER_NAME="$PROVIDER_NAME" MODEL="$MODEL" BASE_URL="$BASE_URL" node <<'NODE'
const fs = require('fs')
const path = process.env.OPENCODE_CONFIG
const providerId = process.env.PROVIDER_ID
const model = process.env.MODEL
let cfg = {}
if (fs.existsSync(path)) {
  try { const raw = fs.readFileSync(path, 'utf8').trim(); cfg = raw ? JSON.parse(raw) : {} }
  catch { fs.renameSync(path, path + '.invalid-erqishi-bak'); cfg = {} }
}
cfg.provider = cfg.provider && typeof cfg.provider === 'object' ? cfg.provider : {}
cfg.provider[providerId] = {
  npm: '@ai-sdk/openai',
  name: process.env.PROVIDER_NAME,
  options: { baseURL: withV1(process.env.BASE_URL), apiKey: '{env:{{ .APIKeyEnv }}}' },
  models: { [model]: { name: model } }
}
cfg.model = providerId + '/' + model
cfg.enabled_providers = Array.from(new Set([...(Array.isArray(cfg.enabled_providers) ? cfg.enabled_providers : []), providerId]))
fs.writeFileSync(path, JSON.stringify(cfg, null, 2) + "\n")
function withV1(value) {
  const trimmed = String(value || '').replace(/\/+$/, '')
  return trimmed.endsWith('/v1') ? trimmed : trimmed + '/v1'
}
NODE
  log "opencode 已配置为 $PROVIDER_NAME 中转。"
}

configure_ccswitch_if_present() {
  local detected="false"
  if has_cmd ccswitch; then detected="true"; elif [ "$(uname -s)" = "Darwin" ] && [ -d "/Applications/CCSwitch.app" ]; then detected="true"; fi
  if [ "$detected" != "true" ]; then log "未检测到 CCSwitch，已跳过。"; return; fi
  local deeplink
  deeplink="$(BASE_URL="$BASE_URL" API_KEY="$API_KEY" PROVIDER_NAME="$PROVIDER_NAME" MODEL="$MODEL" USAGE_SCRIPT="$USAGE_SCRIPT" node <<'NODE'
const params = new URLSearchParams([
  ['resource', 'provider'], ['app', 'codex'], ['model', process.env.MODEL],
  ['name', process.env.PROVIDER_NAME], ['homepage', process.env.BASE_URL],
  ['endpoint', process.env.BASE_URL], ['apiKey', process.env.API_KEY],
  ['configFormat', 'json'], ['usageEnabled', 'true'],
  ['usageScript', Buffer.from(process.env.USAGE_SCRIPT || '', 'utf8').toString('base64')],
  ['usageAutoInterval', '30'],
])
process.stdout.write('ccswitch://v1/import?' + params.toString())
NODE
)"
  if has_cmd ccswitch && ccswitch --help 2>/dev/null | grep -qi "import"; then
    if ccswitch import "$deeplink" >/dev/null 2>&1; then log "CCSwitch 已配置为 $PROVIDER_NAME 中转。"; return; fi
  fi
  if [ "$(uname -s)" = "Darwin" ] && has_cmd open; then open "$deeplink" >/dev/null 2>&1 || true; log "已唤起 CCSwitch 导入配置。"
  elif has_cmd xdg-open; then xdg-open "$deeplink" >/dev/null 2>&1 || true; log "已唤起 CCSwitch 导入配置。"
  else log "检测到 CCSwitch，但未找到可用的协议打开命令，已跳过。"; fi
}

log "开始配置 $PROVIDER_NAME 一键开用..."
ensure_codex
ensure_node
write_env_file
configure_codex
configure_opencode_if_present
configure_ccswitch_if_present
log ""
log "完成。重开终端后直接运行 codex 即可使用 $PROVIDER_NAME 中转。"
log "当前终端可先运行：source \"$HOME/.codex/$PROVIDER_ID.env\" && codex"
`

const windowsBootstrapTemplate = `$ErrorActionPreference = "Stop"

$ProviderId = {{ .ProviderID | psq }}
$ProviderName = {{ .ProviderName | psq }}
$Model = {{ .Model | psq }}
$BaseUrl = {{ .BaseURL | psq }}
$ApiKey = {{ .APIKey | psq }}
$NpmRegistry = {{ .NPMRegistry | psq }}
$NodeMirror = {{ .NodeMirror | psq }}
$NodeVersion = "v22.11.0"
$CodeXPath = "@openai/codex"
$ApiKeyEnv = {{ .APIKeyEnv | psq }}

function Write-Log([string]$Message) { Write-Host $Message }
function Test-Command([string]$Name) { return $null -ne (Get-Command $Name -ErrorAction SilentlyContinue) }
function ConvertTo-TomlString([string]$Value) { return ($Value | ConvertTo-Json -Compress) }
function Join-OpenCodeV1BaseUrl([string]$Value) {
  $trimmed = $Value.TrimEnd("/")
  if ($trimmed.EndsWith("/v1")) { return $trimmed }
  return "$trimmed/v1"
}

function ConvertTo-HashtableCompat($InputObject) {
  if ($null -eq $InputObject) { return $null }
  if ($InputObject -is [System.Collections.IDictionary]) {
    $result = [ordered]@{}
    foreach ($key in $InputObject.Keys) { $result[$key] = ConvertTo-HashtableCompat $InputObject[$key] }
    return $result
  }
  if ($InputObject -is [System.Management.Automation.PSCustomObject]) {
    $result = [ordered]@{}
    foreach ($property in $InputObject.PSObject.Properties) { $result[$property.Name] = ConvertTo-HashtableCompat $property.Value }
    return $result
  }
  if ($InputObject -is [System.Collections.IEnumerable] -and -not ($InputObject -is [string])) {
    $items = @()
    foreach ($item in $InputObject) { $items += ,(ConvertTo-HashtableCompat $item) }
    return $items
  }
  return $InputObject
}

function Set-RootTomlString([string]$Text, [string]$Key, [string]$Value) {
  $line = $Key + " = " + (ConvertTo-TomlString $Value)
  $lines = @()
  if ($Text.Length -gt 0) { $lines = [regex]::Split($Text, "\r?\n") }
  $rootEnd = $lines.Count
  for ($i = 0; $i -lt $lines.Count; $i++) {
    if ($lines[$i] -match '^\s*\[') { $rootEnd = $i; break }
  }
  $keyPattern = '^\s*' + [regex]::Escape($Key) + '\s*=.*$'
  for ($i = 0; $i -lt $rootEnd; $i++) {
    if ($lines[$i] -match $keyPattern) {
      $lines[$i] = $line
      return ($lines -join [Environment]::NewLine)
    }
  }
  return $line + [Environment]::NewLine + ($lines -join [Environment]::NewLine)
}

function Ensure-Node {
  if (Test-Command "npm") { return }
  Write-Log "未检测到 npm，正在从国内镜像安装便携版 Node.js..."
  $arch = if ([Environment]::Is64BitOperatingSystem) { "x64" } else { throw "暂不支持 32 位 Windows，请先手动安装 Node.js。" }
  $runtimeDir = Join-Path $env:USERPROFILE ".erqishi\runtime"
  $nodeDir = Join-Path $runtimeDir "node"
  $archive = Join-Path $runtimeDir "node-$NodeVersion-win-$arch.zip"
  $url = "$NodeMirror/$NodeVersion/node-$NodeVersion-win-$arch.zip"
  New-Item -ItemType Directory -Force -Path $runtimeDir | Out-Null
  Invoke-WebRequest -Uri $url -OutFile $archive -UseBasicParsing
  if (Test-Path $nodeDir) { Remove-Item -Recurse -Force $nodeDir }
  Expand-Archive -Path $archive -DestinationPath $runtimeDir -Force
  $expanded = Join-Path $runtimeDir "node-$NodeVersion-win-$arch"
  Rename-Item -Path $expanded -NewName "node" -Force
  $env:PATH = $nodeDir + ";" + (Join-Path $nodeDir "bin") + ";" + $env:PATH
  $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
  if (-not $userPath) { $userPath = "" }
  if ($userPath -notlike "*$nodeDir*") { [Environment]::SetEnvironmentVariable("Path", ($nodeDir + ";" + $userPath).TrimEnd(";"), "User") }
  if (-not (Test-Command "npm")) { throw "Node.js 安装失败，请检查网络或手动安装 Node.js。" }
}

function Ensure-Codex {
  if (Test-Command "codex") {
    try {
      $version = (& codex --version) -join " "
      if ($LASTEXITCODE -eq 0) { Write-Log "已检测到 Codex：$version"; return }
    } catch {}
  }
  Ensure-Node
  Write-Log "未检测到 Codex，正在通过国内 npm 镜像安装..."
  npm install -g $CodeXPath --registry=$NpmRegistry
  if (-not (Test-Command "codex")) { throw "Codex 安装失败，请检查 npm 全局 bin 目录是否在 PATH 中。" }
  Write-Log "Codex 安装完成：$((& codex --version) -join ' ')"
}

function Write-EnvConfig {
  $codexDir = Join-Path $env:USERPROFILE ".codex"
  New-Item -ItemType Directory -Force -Path $codexDir | Out-Null
  $envFile = Join-Path $codexDir "$ProviderId.env.ps1"
  "Set-Item -Path Env:$ApiKeyEnv -Value '$($ApiKey.Replace("'", "''"))'" | Set-Content -Path $envFile -Encoding UTF8
  [Environment]::SetEnvironmentVariable($ApiKeyEnv, $ApiKey, "User")
  Set-Item -Path "Env:$ApiKeyEnv" -Value $ApiKey
}

function Configure-Codex {
  $codexDir = Join-Path $env:USERPROFILE ".codex"
  New-Item -ItemType Directory -Force -Path $codexDir | Out-Null
  $config = Join-Path $codexDir "config.toml"
  if (-not (Test-Path $config)) { New-Item -ItemType File -Path $config | Out-Null }
  Copy-Item $config "$config.$(Get-Date -Format yyyyMMddHHmmss).erqishi.bak"
  $begin = "# >>> erqishi managed"
  $end = "# <<< erqishi managed"
  $providerNameToml = ConvertTo-TomlString $ProviderName
  $baseUrlToml = ConvertTo-TomlString $BaseUrl
  $apiKeyEnvToml = ConvertTo-TomlString $ApiKeyEnv
  $block = @"
$begin
[model_providers.$ProviderId]
name = $providerNameToml
base_url = $baseUrlToml
env_key = $apiKeyEnvToml
wire_api = "responses"
$end
"@.TrimEnd()
  $text = Get-Content -Path $config -Raw
  if ($null -eq $text) { $text = "" }
  $text = Set-RootTomlString $text "model_provider" $ProviderId
  $text = Set-RootTomlString $text "model" $Model
  $pattern = [regex]::Escape($begin) + "[\s\S]*?" + [regex]::Escape($end)
  if ([regex]::IsMatch($text, $pattern)) {
    $text = [regex]::Replace($text, $pattern, [System.Text.RegularExpressions.MatchEvaluator]{ param($m) $block }, 1)
  } else {
    $nl = [Environment]::NewLine
    $text = $text.TrimEnd() + $nl + $nl + $block + $nl
  }
  Set-Content -Path $config -Value $text -Encoding UTF8
  Write-Log "Codex 已切换到 $ProviderName 中转。"
}

function Configure-OpenCodeIfPresent {
  if (-not (Test-Command "opencode")) { Write-Log "未检测到 opencode，已跳过。"; return }
  try { & opencode --version *> $null; if ($LASTEXITCODE -ne 0) { Write-Log "未检测到可用 opencode，已跳过。"; return } } catch { Write-Log "未检测到可用 opencode，已跳过。"; return }
  $configDir = Join-Path $env:APPDATA "opencode"
  $config = Join-Path $configDir "opencode.json"
  New-Item -ItemType Directory -Force -Path $configDir | Out-Null
  if (Test-Path $config) { Copy-Item $config "$config.$(Get-Date -Format yyyyMMddHHmmss).erqishi.bak" }
  $cfg = [ordered]@{}
  if (Test-Path $config) {
    try { $raw = Get-Content -Path $config -Raw; if ($raw.Trim()) { $cfg = ConvertTo-HashtableCompat ($raw | ConvertFrom-Json) } }
    catch { Rename-Item -Path $config -NewName ((Split-Path $config -Leaf) + ".invalid-erqishi-bak") -Force; $cfg = [ordered]@{} }
  }
  if (-not $cfg.Contains("provider") -or -not ($cfg["provider"] -is [System.Collections.IDictionary])) { $cfg["provider"] = [ordered]@{} }
  $models = [ordered]@{}
  $models[$Model] = [ordered]@{ name = $Model }
  $cfg["provider"][$ProviderId] = [ordered]@{
    npm = "@ai-sdk/openai"
    name = $ProviderName
    options = [ordered]@{ baseURL = (Join-OpenCodeV1BaseUrl $BaseUrl); apiKey = "{env:$ApiKeyEnv}" }
    models = $models
  }
  $cfg["model"] = "$ProviderId/$Model"
  $providers = @()
  if ($cfg.Contains("enabled_providers")) {
    if ($cfg["enabled_providers"] -is [System.Array]) { $providers += $cfg["enabled_providers"] }
    elseif ($cfg["enabled_providers"]) { $providers += $cfg["enabled_providers"] }
  }
  if ($providers -notcontains $ProviderId) { $providers += $ProviderId }
  $cfg["enabled_providers"] = $providers
  $cfg | ConvertTo-Json -Depth 10 | Set-Content -Path $config -Encoding UTF8
  Write-Log "opencode 已配置为 $ProviderName 中转。"
}

function Configure-CCSwitchIfPresent {
  $ccswitch = Get-Command "ccswitch" -ErrorAction SilentlyContinue
  $appPath = Join-Path $env:LOCALAPPDATA "Programs\CCSwitch\CCSwitch.exe"
  if (-not $ccswitch -and -not (Test-Path $appPath)) { Write-Log "未检测到 CCSwitch，已跳过。"; return }
  Add-Type -AssemblyName System.Web
  $usageScript = '({ request: { url: "__BASE_URL__/v1/usage", method: "GET", headers: { "Authorization": "Bearer __API_KEY__" } }, extractor: function(response) { return { isValid: true, remaining: response && (response.remaining || response.balance), unit: "USD" }; } })'
  $usageEncoded = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($usageScript))
  $pairs = [ordered]@{
    resource = "provider"; app = "codex"; model = $Model; name = $ProviderName; homepage = $BaseUrl; endpoint = $BaseUrl; apiKey = $ApiKey; configFormat = "json"; usageEnabled = "true"; usageScript = $usageEncoded; usageAutoInterval = "30"
  }
  $query = ($pairs.GetEnumerator() | ForEach-Object { [System.Web.HttpUtility]::UrlEncode($_.Key) + "=" + [System.Web.HttpUtility]::UrlEncode([string]$_.Value) }) -join "&"
  $deeplink = "ccswitch://v1/import?$query"
  if ($ccswitch) {
    try { & ccswitch import $deeplink *> $null; if ($LASTEXITCODE -eq 0) { Write-Log "CCSwitch 已配置为 $ProviderName 中转。"; return } } catch {}
  }
  Start-Process $deeplink
  Write-Log "已唤起 CCSwitch 导入配置。"
}

Write-Log "开始配置 $ProviderName 一键开用..."
Ensure-Codex
Ensure-Node
Write-EnvConfig
Configure-Codex
Configure-OpenCodeIfPresent
Configure-CCSwitchIfPresent
Write-Log ""
Write-Log "完成。新开的 PowerShell/CMD 窗口中直接运行 codex 即可使用 $ProviderName 中转。"
Write-Log ('当前 PowerShell 可先运行：. "' + $env:USERPROFILE + '\.codex\' + $ProviderId + '.env.ps1"; codex')
`
