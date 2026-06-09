package service

import (
	"strings"
	"testing"
)

func TestExecuteBootstrapTemplateUnix(t *testing.T) {
	out, err := executeBootstrapTemplate(unixBootstrapTemplate, bootstrapTemplateData{
		ProviderID:   BootstrapProviderID,
		ProviderName: BootstrapProviderName,
		APIKeyEnv:    BootstrapAPIKeyEnv,
		Model:        BootstrapCodexModel,
		BaseURL:      "https://api.example.com",
		APIKey:       "sk-test",
		NPMRegistry:  BootstrapNPMRegistry,
		NodeMirror:   BootstrapNodeMirror,
	})
	if err != nil {
		t.Fatalf("executeBootstrapTemplate() error = %v", err)
	}

	mustContainAll(t, out,
		`npm install -g "$CODEX_PACKAGE" --registry="$NPM_REGISTRY"`,
		BootstrapNPMRegistry,
		BootstrapNodeMirror,
		"function upsertRootTomlString",
		"model_provider",
		"wire_api",
		"configure_opencode_if_present",
		"未检测到 opencode，已跳过。",
		"configure_ccswitch_if_present",
		"ccswitch://v1/import?",
		`export ERQISHI_API_KEY="$API_KEY"`,
	)
	mustNotContainAny(t, out, "registry.npmjs.org", "github.com", "chatgpt.com/codex/install.sh", "nodejs.org")
}

func TestExecuteBootstrapTemplateWindows(t *testing.T) {
	out, err := executeBootstrapTemplate(windowsBootstrapTemplate, bootstrapTemplateData{
		ProviderID:   BootstrapProviderID,
		ProviderName: BootstrapProviderName,
		APIKeyEnv:    BootstrapAPIKeyEnv,
		Model:        BootstrapCodexModel,
		BaseURL:      "https://api.example.com",
		APIKey:       "sk-test",
		NPMRegistry:  BootstrapNPMRegistry,
		NodeMirror:   BootstrapNodeMirror,
	})
	if err != nil {
		t.Fatalf("executeBootstrapTemplate() error = %v", err)
	}

	mustContainAll(t, out,
		`$ErrorActionPreference = "Stop"`,
		"Invoke-WebRequest -Uri $url",
		"node-$NodeVersion-win-$arch.zip",
		"npm install -g $CodeXPath --registry=$NpmRegistry",
		`[Environment]::SetEnvironmentVariable($ApiKeyEnv, $ApiKey, "User")`,
		"function Set-RootTomlString",
		"ConvertTo-HashtableCompat",
		"wire_api",
		"Configure-OpenCodeIfPresent",
		"未检测到 opencode，已跳过。",
		"Configure-CCSwitchIfPresent",
		"ccswitch://v1/import?",
		"Start-Process $deeplink",
		BootstrapNPMRegistry,
		BootstrapNodeMirror,
	)
	mustNotContainAny(t, out, "registry.npmjs.org", "github.com", "chatgpt.com/codex/install.sh", "nodejs.org")
}

func TestSanitizeBootstrapFilenamePart(t *testing.T) {
	if got := sanitizeBootstrapFilenamePart("Codex 套餐"); got != "codex" {
		t.Fatalf("sanitizeBootstrapFilenamePart() = %q, want codex", got)
	}
	if got := sanitizeBootstrapFilenamePart("中文"); got != "service" {
		t.Fatalf("sanitizeBootstrapFilenamePart() = %q, want service", got)
	}
}

func TestNormalizeBootstrapBaseURL(t *testing.T) {
	got, err := normalizeBootstrapBaseURL(" https://api.example.com/v1/ ")
	if err != nil {
		t.Fatalf("normalizeBootstrapBaseURL() error = %v", err)
	}
	if got != "https://api.example.com/v1" {
		t.Fatalf("normalizeBootstrapBaseURL() = %q, want https://api.example.com/v1", got)
	}

	if _, err := normalizeBootstrapBaseURL("javascript:alert(1)"); err == nil {
		t.Fatal("normalizeBootstrapBaseURL() expected error for non-http URL")
	}
}

func TestBootstrapLiteralEscaping(t *testing.T) {
	out, err := executeBootstrapTemplate(windowsBootstrapTemplate, bootstrapTemplateData{
		ProviderID:   BootstrapProviderID,
		ProviderName: "Erqi'shi",
		APIKeyEnv:    BootstrapAPIKeyEnv,
		Model:        BootstrapCodexModel,
		BaseURL:      "https://api.example.com/a'b",
		APIKey:       "sk-o'clock",
		NPMRegistry:  BootstrapNPMRegistry,
		NodeMirror:   BootstrapNodeMirror,
	})
	if err != nil {
		t.Fatalf("executeBootstrapTemplate() error = %v", err)
	}
	mustContainAll(t, out, `$BaseUrl = 'https://api.example.com/a''b'`, `$ApiKey = 'sk-o''clock'`)

	out, err = executeBootstrapTemplate(unixBootstrapTemplate, bootstrapTemplateData{
		ProviderID:   BootstrapProviderID,
		ProviderName: "Erqi'shi",
		APIKeyEnv:    BootstrapAPIKeyEnv,
		Model:        BootstrapCodexModel,
		BaseURL:      "https://api.example.com/a'b",
		APIKey:       "sk-o'clock",
		NPMRegistry:  BootstrapNPMRegistry,
		NodeMirror:   BootstrapNodeMirror,
	})
	if err != nil {
		t.Fatalf("executeBootstrapTemplate() error = %v", err)
	}
	mustContainAll(t, out, `BASE_URL='https://api.example.com/a'\''b'`, `API_KEY='sk-o'\''clock'`)
}

func mustContainAll(t *testing.T, value string, needles ...string) {
	t.Helper()
	for _, needle := range needles {
		if !strings.Contains(value, needle) {
			t.Fatalf("expected output to contain %q", needle)
		}
	}
}

func mustNotContainAny(t *testing.T, value string, needles ...string) {
	t.Helper()
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			t.Fatalf("expected output not to contain %q", needle)
		}
	}
}
