package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadProjectTypesFromEnvRoot(t *testing.T) {
	templatesDir := templatesPath(t)
	t.Setenv(templatesEnvVar, templatesDir)

	projectTypes, err := loadProjectTypes()
	if err != nil {
		t.Fatalf("load project types: %v", err)
	}

	expected := []string{"elixir_phoenix_project", "php_silverstripe_project"}
	for _, id := range expected {
		if _, ok := projectTypes[id]; !ok {
			t.Fatalf("expected project type %s", id)
		}
	}
}

func TestRunInitCreatesConfigs(t *testing.T) {
	templatesDir := templatesPath(t)
	t.Setenv(templatesEnvVar, templatesDir)

	rootDir := t.TempDir()
	if err := runInit([]string{"--type", "php_silverstripe_project", "--path", rootDir}); err != nil {
		t.Fatalf("run init: %v", err)
	}

	assertFileExists(t, filepath.Join(rootDir, projectConfigFilename))
	assertFileExists(t, filepath.Join(rootDir, opencodeConfigFilename))
	assertFileExists(t, filepath.Join(rootDir, mcpConfigFilename))
	assertFileExists(t, filepath.Join(rootDir, ".codex", agentsFilename))
	assertFileExists(t, filepath.Join(rootDir, ".claude", claudeFilename))
	assertDirExists(t, filepath.Join(rootDir, ".codex", "skills", "beans"))
	assertDirExists(t, filepath.Join(rootDir, ".claude", "skills", "beans"))
	assertFileExists(t, filepath.Join(rootDir, ".opencode", "command", "teach.md"))
	assertFileExists(t, filepath.Join(rootDir, ".claude", "commands", "teach.md"))
}

func TestRunInitVerboseLogsCreatedFiles(t *testing.T) {
	defer func() {
		verboseOutput = false
	}()

	templatesDir := templatesPath(t)
	t.Setenv(templatesEnvVar, templatesDir)

	rootDir := t.TempDir()
	var runErr error
	output := captureOutput(t, func() {
		runErr = runInit([]string{"--type", "php_silverstripe_project", "--path", rootDir, "--verbose"})
	})
	if runErr != nil {
		t.Fatalf("run init: %v", runErr)
	}

	commandPath := filepath.Join(rootDir, ".opencode", "command", "teach.md")
	expected := fmt.Sprintf("Created %s", commandPath)
	if !strings.Contains(output, expected) {
		t.Fatalf("expected verbose output to include %q", expected)
	}
}

func TestRunUpdateUsesStoredProjectType(t *testing.T) {
	templatesDir := templatesPath(t)
	t.Setenv(templatesEnvVar, templatesDir)

	rootDir := t.TempDir()
	if err := runInit([]string{"--type", "php_silverstripe_project", "--path", rootDir}); err != nil {
		t.Fatalf("run init: %v", err)
	}

	claudeSkill := filepath.Join(rootDir, ".claude", "skills", "beans", "SKILL.md")
	if err := os.Remove(claudeSkill); err != nil {
		t.Fatalf("remove skill: %v", err)
	}

	if err := runUpdate([]string{"--path", rootDir}); err != nil {
		t.Fatalf("run update: %v", err)
	}

	assertFileExists(t, claudeSkill)
	projectType, err := readProjectTypeFile(rootDir)
	if err != nil {
		t.Fatalf("read project type: %v", err)
	}
	if projectType != "php_silverstripe_project" {
		t.Fatalf("expected project type php_silverstripe_project, got %s", projectType)
	}
}

func TestCommandOverrides(t *testing.T) {
	templatesDir := createTemplateRoot(t)
	writeTemplateFile(t, templatesDir, filepath.Join(projectTypesDir, "command_overrides.toml"), `name = "Command Overrides"
skills = ["core-skill"]
commands = ["teach"]

[claude]
commands = ["coach"]

[codex]
commands = ["lead"]
`)
	writeTemplateFile(t, templatesDir, filepath.Join(skillTemplatesDir, "core-skill", "SKILL.md"), "core")
	writeTemplateFile(t, templatesDir, filepath.Join(commandTemplatesDir, "teach.md"), "teach")
	writeTemplateFile(t, templatesDir, filepath.Join(commandTemplatesDir, "coach.md"), "coach")
	writeTemplateFile(t, templatesDir, filepath.Join(commandTemplatesDir, "lead.md"), "lead")
	t.Setenv(templatesEnvVar, templatesDir)

	rootDir := t.TempDir()
	if err := runInit([]string{"--type", "command_overrides", "--path", rootDir}); err != nil {
		t.Fatalf("run init: %v", err)
	}

	assertFileExists(t, filepath.Join(rootDir, ".claude", "commands", "coach.md"))
	assertFileNotExists(t, filepath.Join(rootDir, ".claude", "commands", "teach.md"))
	assertFileExists(t, filepath.Join(rootDir, ".opencode", "command", "lead.md"))
	assertFileNotExists(t, filepath.Join(rootDir, ".opencode", "command", "teach.md"))
}

func TestMcpMergeDedupe(t *testing.T) {
	templatesDir := createTemplateRoot(t)
	writeTemplateFile(t, templatesDir, filepath.Join(projectTypesDir, "mcp_merge.toml"), `name = "MCP Merge"
skills = ["core-skill"]
commands = ["teach"]
mcps = ["demo"]

[claude]
mcps = ["demo", "tidewave"]

[codex]
mcps = ["demo"]
`)
	writeTemplateFile(t, templatesDir, filepath.Join(skillTemplatesDir, "core-skill", "SKILL.md"), "core")
	writeTemplateFile(t, templatesDir, filepath.Join(commandTemplatesDir, "teach.md"), "teach")
	writeTemplateFile(t, templatesDir, filepath.Join(mcpTemplatesDir, "demo.mcp.json"), `{"demo": {"type": "remote", "url": "http://localhost/demo", "enabled": true}}`)
	writeTemplateFile(t, templatesDir, filepath.Join(mcpTemplatesDir, "tidewave.mcp.json"), `{"tidewave": {"type": "remote", "url": "http://localhost/tidewave", "enabled": true}}`)
	t.Setenv(templatesEnvVar, templatesDir)

	rootDir := t.TempDir()
	if err := runInit([]string{"--type", "mcp_merge", "--path", rootDir}); err != nil {
		t.Fatalf("run init: %v", err)
	}

	var openCfg openCodeConfig
	loadJSON(t, filepath.Join(rootDir, opencodeConfigFilename), &openCfg)
	if len(openCfg.Mcp) != 2 {
		t.Fatalf("expected 2 mcp entries, got %d", len(openCfg.Mcp))
	}
	if _, ok := openCfg.Mcp["demo"]; !ok {
		t.Fatalf("expected demo in opencode config")
	}
	if _, ok := openCfg.Mcp["tidewave"]; !ok {
		t.Fatalf("expected tidewave in opencode config")
	}

	var mcpCfg mcpConfig
	loadJSON(t, filepath.Join(rootDir, mcpConfigFilename), &mcpCfg)
	if len(mcpCfg.McpServers) != 2 {
		t.Fatalf("expected 2 mcp servers, got %d", len(mcpCfg.McpServers))
	}
	if _, ok := mcpCfg.McpServers["demo"]; !ok {
		t.Fatalf("expected demo in mcp config")
	}
	if _, ok := mcpCfg.McpServers["tidewave"]; !ok {
		t.Fatalf("expected tidewave in mcp config")
	}
}

func TestCommandCopyLayout(t *testing.T) {
	templatesDir := createTemplateRoot(t)
	writeTemplateFile(t, templatesDir, filepath.Join(projectTypesDir, "command_layout.toml"), `name = "Command Layout"
skills = ["core-skill"]
commands = ["teach"]
`)
	writeTemplateFile(t, templatesDir, filepath.Join(skillTemplatesDir, "core-skill", "SKILL.md"), "core")
	writeTemplateFile(t, templatesDir, filepath.Join(commandTemplatesDir, "teach.md"), "teach")
	t.Setenv(templatesEnvVar, templatesDir)

	rootDir := t.TempDir()
	if err := runInit([]string{"--type", "command_layout", "--path", rootDir}); err != nil {
		t.Fatalf("run init: %v", err)
	}

	assertDirExists(t, filepath.Join(rootDir, ".opencode", "command"))
	assertDirExists(t, filepath.Join(rootDir, ".claude", "commands"))
	assertFileExists(t, filepath.Join(rootDir, ".opencode", "command", "teach.md"))
	assertFileExists(t, filepath.Join(rootDir, ".claude", "commands", "teach.md"))
}

func TestRunInitMissingCommandTemplate(t *testing.T) {
	templatesDir := createTemplateRoot(t)
	writeTemplateFile(t, templatesDir, filepath.Join(projectTypesDir, "missing_command.toml"), `name = "Missing Command"
skills = ["core-skill"]
commands = ["missing"]
`)
	writeTemplateFile(t, templatesDir, filepath.Join(skillTemplatesDir, "core-skill", "SKILL.md"), "core")
	t.Setenv(templatesEnvVar, templatesDir)

	rootDir := t.TempDir()
	if err := runInit([]string{"--type", "missing_command", "--path", rootDir}); err == nil {
		t.Fatal("expected missing command template error")
	} else if !strings.Contains(err.Error(), "command template not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunInitEmptyMcpTemplate(t *testing.T) {
	templatesDir := createTemplateRoot(t)
	writeTemplateFile(t, templatesDir, filepath.Join(projectTypesDir, "empty_mcp.toml"), `name = "Empty MCP"
skills = ["core-skill"]
commands = ["teach"]
mcps = ["empty"]
`)
	writeTemplateFile(t, templatesDir, filepath.Join(skillTemplatesDir, "core-skill", "SKILL.md"), "core")
	writeTemplateFile(t, templatesDir, filepath.Join(commandTemplatesDir, "teach.md"), "teach")
	writeTemplateFile(t, templatesDir, filepath.Join(mcpTemplatesDir, "empty.mcp.json"), `{}`)
	t.Setenv(templatesEnvVar, templatesDir)

	rootDir := t.TempDir()
	if err := runInit([]string{"--type", "empty_mcp", "--path", rootDir}); err == nil {
		t.Fatal("expected empty mcp template error")
	} else if !strings.Contains(err.Error(), "mcp template empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExpandEnvVarsReportsMissing(t *testing.T) {
	t.Setenv("KNOWN_ENV", "value")

	expanded, missing := expandEnvVars("start ${KNOWN_ENV} ${MISSING_ENV}")
	if expanded != "start value ${MISSING_ENV}" {
		t.Fatalf("unexpected expanded value: %s", expanded)
	}
	if len(missing) != 1 || missing[0] != "MISSING_ENV" {
		t.Fatalf("unexpected missing vars: %v", missing)
	}
}

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	_ = w.Close()
	os.Stdout = original
	output := <-done
	_ = r.Close()
	return output
}

func createTemplateRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dirs := []string{projectTypesDir, skillTemplatesDir, commandTemplatesDir, mcpTemplatesDir, snippetsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	return root
}

func writeTemplateFile(t *testing.T, root, relPath, content string) {
	t.Helper()
	path := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", relPath, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", relPath, err)
	}
}

func loadJSON(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
}

func templatesPath(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	return filepath.Join(cwd, "templates")
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected file %s: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("expected file but found directory: %s", path)
	}
}

func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	if err == nil {
		t.Fatalf("expected no file at %s", path)
	}
	if !os.IsNotExist(err) {
		t.Fatalf("expected not exist for %s: %v", path, err)
	}
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected directory %s: %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("expected directory but found file: %s", path)
	}
}
