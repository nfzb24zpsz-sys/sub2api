package service

import (
	"archive/zip"
	"bytes"
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
		"write_codex_auth_file",
		"OPENAI_API_KEY",
		"requires_openai_auth",
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
		"Write-CodexAuthFile",
		"OPENAI_API_KEY",
		"requires_openai_auth",
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

func TestBuildBootstrapPackageWindows(t *testing.T) {
	data := bootstrapTemplateData{
		ProviderID:   BootstrapProviderID,
		ProviderName: BootstrapProviderName,
		APIKeyEnv:    BootstrapAPIKeyEnv,
		Model:        BootstrapCodexModel,
		BaseURL:      "https://api.example.com",
		APIKey:       "sk-test",
		NPMRegistry:  BootstrapNPMRegistry,
		NodeMirror:   BootstrapNodeMirror,
	}
	pkg, err := buildBootstrapPackage(data, BootstrapTargetWindows)
	if err != nil {
		t.Fatalf("buildBootstrapPackage() error = %v", err)
	}
	files := readZipFiles(t, pkg)
	mustHaveZipFile(t, files, "一键开用.cmd")
	mustHaveZipFile(t, files, "README.txt")
	mustNotHaveZipFile(t, files, "一键开用.command")
	cmd := readZipFile(t, files["一键开用.cmd"])
	mustContainAll(t, cmd, "powershell -NoProfile -ExecutionPolicy Bypass", "::ERQISHI_POWERSHELL::")
	readme := readZipFile(t, files["README.txt"])
	mustContainAll(t, readme, "Windows:", "macOS:")
}

func TestBuildBootstrapPackageUnix(t *testing.T) {
	data := bootstrapTemplateData{
		ProviderID:   BootstrapProviderID,
		ProviderName: BootstrapProviderName,
		APIKeyEnv:    BootstrapAPIKeyEnv,
		Model:        BootstrapCodexModel,
		BaseURL:      "https://api.example.com",
		APIKey:       "sk-test",
		NPMRegistry:  BootstrapNPMRegistry,
		NodeMirror:   BootstrapNodeMirror,
	}
	pkg, err := buildBootstrapPackage(data, BootstrapTargetUnix)
	if err != nil {
		t.Fatalf("buildBootstrapPackage() error = %v", err)
	}
	files := readZipFiles(t, pkg)
	mustHaveZipFile(t, files, "一键开用.command")
	mustHaveZipFile(t, files, "README.txt")
	mustNotHaveZipFile(t, files, "一键开用.cmd")
	if got := files["一键开用.command"].Mode().Perm(); got != 0o755 {
		t.Fatalf("macOS command mode = %o, want 755", got)
	}
	mac := readZipFile(t, files["一键开用.command"])
	mustContainAll(t, mac, "#!/usr/bin/env bash", "ensure_codex")
	readme := readZipFile(t, files["README.txt"])
	mustContainAll(t, readme, "Windows:", "macOS:")
}

func readZipFiles(t *testing.T, content []byte) map[string]*zip.File {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("zip.NewReader() error = %v", err)
	}
	files := map[string]*zip.File{}
	for _, file := range zr.File {
		files[file.Name] = file
	}
	return files
}

func mustHaveZipFile(t *testing.T, files map[string]*zip.File, name string) {
	t.Helper()
	if files[name] == nil {
		t.Fatalf("expected zip to contain %s", name)
	}
}

func mustNotHaveZipFile(t *testing.T, files map[string]*zip.File, name string) {
	t.Helper()
	if files[name] != nil {
		t.Fatalf("expected zip not to contain %s", name)
	}
}

func readZipFile(t *testing.T, file *zip.File) string {
	t.Helper()
	rc, err := file.Open()
	if err != nil {
		t.Fatalf("open zip file %s: %v", file.Name, err)
	}
	defer rc.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(rc); err != nil {
		t.Fatalf("read zip file %s: %v", file.Name, err)
	}
	return buf.String()
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
