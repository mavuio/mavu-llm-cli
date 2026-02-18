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
	"time"

	"github.com/pelletier/go-toml/v2"
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

	assertFileExists(t, filepath.Join(rootDir, mavuDirName, localConfigFilename))
	assertFileExists(t, filepath.Join(rootDir, opencodeConfigFilename))
	assertFileExists(t, filepath.Join(rootDir, mcpConfigFilename))
	assertFileExists(t, filepath.Join(rootDir, ".codex", agentsFilename))
	assertFileExists(t, filepath.Join(rootDir, ".codex", "config.toml"))
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

	var codexCfg map[string]any
	loadTOML(t, filepath.Join(rootDir, ".codex", "config.toml"), &codexCfg)
	mcpServers, ok := codexCfg["mcp_servers"].(map[string]any)
	if !ok {
		t.Fatalf("expected mcp_servers table in codex config")
	}
	if len(mcpServers) != 2 {
		t.Fatalf("expected 2 mcp servers in codex config, got %d", len(mcpServers))
	}
	if _, ok := mcpServers["demo"]; !ok {
		t.Fatalf("expected demo in codex config")
	}
	if _, ok := mcpServers["tidewave"]; !ok {
		t.Fatalf("expected tidewave in codex config")
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

func TestSessionProjectNameMatchesDirectChild(t *testing.T) {
	name, ok := sessionProjectName("/www/chatty", []string{"/www"})
	if !ok {
		t.Fatal("expected project name to be detected")
	}
	if name != "chatty" {
		t.Fatalf("expected chatty, got %s", name)
	}
}

func TestSessionProjectNameRejectsNestedDirectories(t *testing.T) {
	if _, ok := sessionProjectName("/www/chatty/subdir", []string{"/www"}); ok {
		t.Fatal("expected nested directory to be ignored")
	}
}

func TestSessionProjectNameSupportsResolvedRootCandidate(t *testing.T) {
	rootCandidates := []string{"/www", "/Users/manfred/Documents/www"}
	name, ok := sessionProjectName("/Users/manfred/Documents/www/wah-ex", rootCandidates)
	if !ok {
		t.Fatal("expected project name to be detected using resolved root")
	}
	if name != "wah-ex" {
		t.Fatalf("expected wah-ex, got %s", name)
	}
}

func TestFormatSessionUpdatedAtShowsTimeForToday(t *testing.T) {
	now := time.Date(2026, time.February, 13, 10, 30, 0, 0, time.UTC)
	updated := time.Date(2026, time.February, 13, 8, 45, 0, 0, time.UTC).UnixMilli()

	formatted := formatSessionUpdatedAt(updated, now)
	if formatted != "08:45" {
		t.Fatalf("expected time-only format for today, got %q", formatted)
	}
}

func TestFormatSessionUpdatedAtShowsDateForOlderSessions(t *testing.T) {
	now := time.Date(2026, time.February, 13, 10, 30, 0, 0, time.UTC)
	updated := time.Date(2026, time.February, 12, 22, 15, 0, 0, time.UTC).UnixMilli()

	formatted := formatSessionUpdatedAt(updated, now)
	if formatted != "2026-02-12" {
		t.Fatalf("expected date-only format for older sessions, got %q", formatted)
	}
}

func TestFindOpenCodeSessionsReadsStorageAndExcludesPrefix(t *testing.T) {
	rootDir := t.TempDir()
	storageDir := t.TempDir()

	chattyDir := filepath.Join(rootDir, "chatty")
	archiveDir := filepath.Join(rootDir, "archive-old")
	outsideDir := filepath.Join(t.TempDir(), "outside")

	mustMkdirAll(t, chattyDir)
	mustMkdirAll(t, archiveDir)
	mustMkdirAll(t, outsideDir)
	mustMkdirAll(t, filepath.Join(storageDir, "project"))
	mustMkdirAll(t, filepath.Join(storageDir, "session", "proj-chatty"))
	mustMkdirAll(t, filepath.Join(storageDir, "session", "proj-archive"))

	writeJSONFile(t, filepath.Join(storageDir, "project", "proj-chatty.json"), map[string]any{
		"id":       "proj-chatty",
		"worktree": chattyDir,
	})
	writeJSONFile(t, filepath.Join(storageDir, "project", "proj-archive.json"), map[string]any{
		"id":       "proj-archive",
		"worktree": archiveDir,
	})
	writeJSONFile(t, filepath.Join(storageDir, "project", "proj-outside.json"), map[string]any{
		"id":       "proj-outside",
		"worktree": outsideDir,
	})
	writeJSONFile(t, filepath.Join(storageDir, "project", "proj-missing.json"), map[string]any{
		"id":       "proj-missing",
		"worktree": filepath.Join(rootDir, "missing-project"),
	})

	writeJSONFile(t, filepath.Join(storageDir, "session", "proj-chatty", "ses-new.json"), map[string]any{
		"id":    "ses-new",
		"title": "Newest Chatty Session",
		"time": map[string]any{
			"updated": int64(200),
		},
	})
	writeJSONFile(t, filepath.Join(storageDir, "session", "proj-chatty", "ses-old.json"), map[string]any{
		"id":    "ses-old",
		"title": "",
		"time": map[string]any{
			"updated": int64(100),
		},
	})
	writeJSONFile(t, filepath.Join(storageDir, "session", "proj-chatty", "ses-archive-title.json"), map[string]any{
		"id":    "ses-archive-title",
		"title": "archive hidden session",
		"time": map[string]any{
			"updated": int64(250),
		},
	})
	writeJSONFile(t, filepath.Join(storageDir, "session", "proj-chatty", "ses-explore.json"), map[string]any{
		"id":    "ses-explore",
		"title": "Inspect map API (@explore subagent)",
		"time": map[string]any{
			"updated": int64(240),
		},
	})
	writeJSONFile(t, filepath.Join(storageDir, "session", "proj-archive", "ses-archive.json"), map[string]any{
		"id":    "ses-archive",
		"title": "Archive Session",
		"time": map[string]any{
			"updated": int64(300),
		},
	})
	mustMkdirAll(t, filepath.Join(storageDir, "session", "proj-missing"))
	writeJSONFile(t, filepath.Join(storageDir, "session", "proj-missing", "ses-missing.json"), map[string]any{
		"id":    "ses-missing",
		"title": "Missing Project Session",
		"time": map[string]any{
			"updated": int64(400),
		},
	})

	sessions, err := findOpenCodeSessions(rootDir, "archive", storageDir)
	if err != nil {
		t.Fatalf("find sessions: %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	if sessions[0].ID != "ses-new" {
		t.Fatalf("expected first session ses-new, got %s", sessions[0].ID)
	}
	if sessions[1].ID != "ses-old" {
		t.Fatalf("expected second session ses-old, got %s", sessions[1].ID)
	}
	if sessions[0].Project != "chatty" {
		t.Fatalf("expected project chatty, got %s", sessions[0].Project)
	}
}

func TestFindOpenCodeSessionsUsesSessionDirectoryWhenPresent(t *testing.T) {
	rootDir := t.TempDir()
	storageDir := t.TempDir()

	chattyDir := filepath.Join(rootDir, "chatty")
	archiveDir := filepath.Join(rootDir, "archive-old")
	mustMkdirAll(t, chattyDir)
	mustMkdirAll(t, archiveDir)
	mustMkdirAll(t, filepath.Join(storageDir, "project"))
	mustMkdirAll(t, filepath.Join(storageDir, "session", "proj-shared"))

	writeJSONFile(t, filepath.Join(storageDir, "project", "proj-shared.json"), map[string]any{
		"id":       "proj-shared",
		"worktree": archiveDir,
	})
	writeJSONFile(t, filepath.Join(storageDir, "session", "proj-shared", "ses-1.json"), map[string]any{
		"id":        "ses-1",
		"title":     "Session from sandbox directory",
		"directory": chattyDir,
		"time": map[string]any{
			"updated": int64(50),
		},
	})

	sessions, err := findOpenCodeSessions(rootDir, "archive", storageDir)
	if err != nil {
		t.Fatalf("find sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].Project != "chatty" {
		t.Fatalf("expected project chatty, got %s", sessions[0].Project)
	}
}

func TestWorktreeExistsFalseForMissingPath(t *testing.T) {
	if worktreeExists(filepath.Join(t.TempDir(), "does-not-exist")) {
		t.Fatal("expected missing path to be false")
	}
}

func TestListOpenCodeSessionsPrintsSessionTitles(t *testing.T) {
	rootDir := t.TempDir()
	storageDir := t.TempDir()

	chattyDir := filepath.Join(rootDir, "chatty")
	mustMkdirAll(t, chattyDir)
	mustMkdirAll(t, filepath.Join(storageDir, "project"))
	mustMkdirAll(t, filepath.Join(storageDir, "session", "proj-chatty"))

	writeJSONFile(t, filepath.Join(storageDir, "project", "proj-chatty.json"), map[string]any{
		"id":       "proj-chatty",
		"worktree": chattyDir,
	})
	writeJSONFile(t, filepath.Join(storageDir, "session", "proj-chatty", "ses-1.json"), map[string]any{
		"id":    "ses-1",
		"title": "Chatty Session One",
		"time": map[string]any{
			"updated": int64(1739443560000),
		},
	})
	writeJSONFile(t, filepath.Join(storageDir, "session", "proj-chatty", "ses-explore.json"), map[string]any{
		"id":    "ses-explore",
		"title": "Debug route (@explore subagent)",
		"time": map[string]any{
			"updated": int64(1739443561000),
		},
	})

	var runErr error
	output := captureOutput(t, func() {
		runErr = listOpenCodeSessions([]string{"--path", rootDir, "--storage-path", storageDir})
	})
	if runErr != nil {
		t.Fatalf("list opencode sessions: %v", runErr)
	}

	if !strings.Contains(output, "chatty") {
		t.Fatalf("expected output to include project name, got %q", output)
	}
	if !strings.Contains(output, "Chatty Session One") {
		t.Fatalf("expected output to include session title, got %q", output)
	}
	expectedTime := formatSessionUpdated(1739443560000)
	if !strings.Contains(output, expectedTime) {
		t.Fatalf("expected output to include session timestamp %q, got %q", expectedTime, output)
	}
	if strings.Contains(output, "UTC") {
		t.Fatalf("expected output to omit UTC marker, got %q", output)
	}
	if strings.Contains(output, "(@explore subagent)") {
		t.Fatalf("expected output to omit explore-subagent titles, got %q", output)
	}
	if strings.Contains(output, chattyDir) {
		t.Fatalf("expected output not to include directory path, got %q", output)
	}
}

func TestListOpenCodeSessionsAppliesLineFilter(t *testing.T) {
	rootDir := t.TempDir()
	storageDir := t.TempDir()

	filmDir := filepath.Join(rootDir, "filmarchiv-ex")
	chattyDir := filepath.Join(rootDir, "chatty")
	orfDir := filepath.Join(rootDir, "orfaudio")
	mustMkdirAll(t, filmDir)
	mustMkdirAll(t, chattyDir)
	mustMkdirAll(t, orfDir)
	mustMkdirAll(t, filepath.Join(storageDir, "project"))
	mustMkdirAll(t, filepath.Join(storageDir, "session", "proj-film"))
	mustMkdirAll(t, filepath.Join(storageDir, "session", "proj-chatty"))
	mustMkdirAll(t, filepath.Join(storageDir, "session", "proj-orf"))

	writeJSONFile(t, filepath.Join(storageDir, "project", "proj-film.json"), map[string]any{
		"id":       "proj-film",
		"worktree": filmDir,
	})
	writeJSONFile(t, filepath.Join(storageDir, "project", "proj-chatty.json"), map[string]any{
		"id":       "proj-chatty",
		"worktree": chattyDir,
	})
	writeJSONFile(t, filepath.Join(storageDir, "project", "proj-orf.json"), map[string]any{
		"id":       "proj-orf",
		"worktree": orfDir,
	})
	writeJSONFile(t, filepath.Join(storageDir, "session", "proj-film", "ses-film.json"), map[string]any{
		"id":    "ses-film",
		"title": "Map issue",
		"time": map[string]any{
			"updated": int64(100),
		},
	})
	writeJSONFile(t, filepath.Join(storageDir, "session", "proj-chatty", "ses-chatty.json"), map[string]any{
		"id":    "ses-chatty",
		"title": "General chat",
		"time": map[string]any{
			"updated": int64(90),
		},
	})
	writeJSONFile(t, filepath.Join(storageDir, "session", "proj-orf", "ses-orf.json"), map[string]any{
		"id":    "ses-orf",
		"title": "Managing orfaudio/filmarchiv-ex container instances",
		"time": map[string]any{
			"updated": int64(80),
		},
	})

	var runErr error
	output := captureOutput(t, func() {
		runErr = listOpenCodeSessions([]string{"--path", rootDir, "--storage-path", storageDir, "filmarchiv-ex:"})
	})
	if runErr != nil {
		t.Fatalf("list opencode sessions: %v", runErr)
	}

	if !strings.Contains(output, "filmarchiv-ex") {
		t.Fatalf("expected output to include filtered project line, got %q", output)
	}
	if strings.Contains(output, "chatty") {
		t.Fatalf("expected output to exclude non-matching line, got %q", output)
	}
	if strings.Contains(output, "orfaudio") {
		t.Fatalf("expected output to exclude title-only substring matches, got %q", output)
	}
}

func TestShouldSkipSessionTitle(t *testing.T) {
	if !shouldSkipSessionTitle("archive step 1", "archive") {
		t.Fatal("expected archive-prefixed title to be skipped")
	}
	if !shouldSkipSessionTitle("future: add caching", "archive") {
		t.Fatal("expected future-prefixed title to be skipped")
	}
	if !shouldSkipSessionTitle("Inspect map (@explore subagent)", "archive") {
		t.Fatal("expected explore-subagent title to be skipped")
	}
	if shouldSkipSessionTitle("Regular session title", "archive") {
		t.Fatal("expected normal title not to be skipped")
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
	dirs := []string{projectTypesDir, skillTemplatesDir, commandTemplatesDir, mcpTemplatesDir, sessionTemplatesDir, snippetsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	return root
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func writeJSONFile(t *testing.T, path string, payload any) {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal %s: %v", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
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

func loadTOML(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := toml.Unmarshal(data, target); err != nil {
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

func TestSessionsGenerateVSCodeTasks(t *testing.T) {
	templatesDir := createTemplateRoot(t)
	writeTemplateFile(t, templatesDir, filepath.Join(projectTypesDir, "sessions_test.toml"), `name = "Sessions Test"
skills = ["core-skill"]
autostart_sessions = ["dev_server", "watch_css"]
ondemand_sessions = ["deploy", "logs"]
`)
	writeTemplateFile(t, templatesDir, filepath.Join(skillTemplatesDir, "core-skill", "SKILL.md"), "core")
	writeTemplateFile(t, templatesDir, filepath.Join(sessionTemplatesDir, "dev_server.toml"), `name = "Dev Server"
command = "mix phx.server"
`)
	writeTemplateFile(t, templatesDir, filepath.Join(sessionTemplatesDir, "watch_css.toml"), `name = "Watch CSS"
command = "pnpm watch-css"
`)
	writeTemplateFile(t, templatesDir, filepath.Join(sessionTemplatesDir, "deploy.toml"), `name = "Deploy"
command = "git push dokku main"
`)
	writeTemplateFile(t, templatesDir, filepath.Join(sessionTemplatesDir, "logs.toml"), `name = "Logs"
command = "dokku logs -t"
`)
	t.Setenv(templatesEnvVar, templatesDir)

	rootDir := t.TempDir()
	if err := runInit([]string{"--type", "sessions_test", "--path", rootDir}); err != nil {
		t.Fatalf("run init: %v", err)
	}

	tasksPath := filepath.Join(rootDir, ".vscode", "tasks.json")
	assertFileExists(t, tasksPath)

	var tasksJSON map[string]any
	loadJSON(t, tasksPath, &tasksJSON)

	if tasksJSON["version"] != "2.0.0" {
		t.Fatalf("expected version 2.0.0, got %v", tasksJSON["version"])
	}

	tasks, ok := tasksJSON["tasks"].([]any)
	if !ok {
		t.Fatal("expected tasks array")
	}
	// 4 individual tasks + 1 compound task
	if len(tasks) != 5 {
		t.Fatalf("expected 5 tasks, got %d", len(tasks))
	}

	// Check autostart tasks have runOn: folderOpen
	tasksByLabel := make(map[string]map[string]any)
	for _, task := range tasks {
		taskMap := task.(map[string]any)
		label := taskMap["label"].(string)
		tasksByLabel[label] = taskMap
	}

	// Individual tasks should NOT have runOptions
	devServer := tasksByLabel["Dev Server"]
	if devServer == nil {
		t.Fatal("expected Dev Server task")
	}
	if _, hasRunOptions := devServer["runOptions"]; hasRunOptions {
		t.Fatal("individual task should not have runOptions")
	}

	// Compound task should depend on autostart sessions, no runOptions
	compound := tasksByLabel["__ Start Default Terminal Sessions"]
	if compound == nil {
		t.Fatal("expected compound task")
	}
	dependsOn, ok := compound["dependsOn"].([]any)
	if !ok {
		t.Fatal("expected dependsOn array in compound task")
	}
	if len(dependsOn) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(dependsOn))
	}
	if _, hasRunOptions := compound["runOptions"]; hasRunOptions {
		t.Fatal("compound task should not have runOptions")
	}
}

func TestSessionsMissingTemplate(t *testing.T) {
	templatesDir := createTemplateRoot(t)
	writeTemplateFile(t, templatesDir, filepath.Join(projectTypesDir, "missing_session.toml"), `name = "Missing Session"
skills = ["core-skill"]
autostart_sessions = ["nonexistent"]
`)
	writeTemplateFile(t, templatesDir, filepath.Join(skillTemplatesDir, "core-skill", "SKILL.md"), "core")
	t.Setenv(templatesEnvVar, templatesDir)

	rootDir := t.TempDir()
	err := runInit([]string{"--type", "missing_session", "--path", rootDir})
	if err == nil {
		t.Fatal("expected missing session template error")
	}
	if !strings.Contains(err.Error(), "session template not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSessionsMissingCommand(t *testing.T) {
	templatesDir := createTemplateRoot(t)
	writeTemplateFile(t, templatesDir, filepath.Join(projectTypesDir, "empty_session.toml"), `name = "Empty Session"
skills = ["core-skill"]
ondemand_sessions = ["empty"]
`)
	writeTemplateFile(t, templatesDir, filepath.Join(skillTemplatesDir, "core-skill", "SKILL.md"), "core")
	writeTemplateFile(t, templatesDir, filepath.Join(sessionTemplatesDir, "empty.toml"), `name = "Empty"
`)
	t.Setenv(templatesEnvVar, templatesDir)

	rootDir := t.TempDir()
	err := runInit([]string{"--type", "empty_session", "--path", rootDir})
	if err == nil {
		t.Fatal("expected session missing command error")
	}
	if !strings.Contains(err.Error(), "missing command") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSessionsNoSessionsConfigured(t *testing.T) {
	templatesDir := createTemplateRoot(t)
	writeTemplateFile(t, templatesDir, filepath.Join(projectTypesDir, "no_sessions.toml"), `name = "No Sessions"
skills = ["core-skill"]
`)
	writeTemplateFile(t, templatesDir, filepath.Join(skillTemplatesDir, "core-skill", "SKILL.md"), "core")
	t.Setenv(templatesEnvVar, templatesDir)

	rootDir := t.TempDir()
	if err := runInit([]string{"--type", "no_sessions", "--path", rootDir}); err != nil {
		t.Fatalf("run init: %v", err)
	}

	// No tasks.json should be created
	tasksPath := filepath.Join(rootDir, ".vscode", "tasks.json")
	assertFileNotExists(t, tasksPath)
}

func TestSessionsLocalConfigAppend(t *testing.T) {
	templatesDir := createTemplateRoot(t)
	writeTemplateFile(t, templatesDir, filepath.Join(projectTypesDir, "local_sessions.toml"), `name = "Local Sessions"
skills = ["core-skill"]
autostart_sessions = ["dev_server"]
ondemand_sessions = ["deploy"]
`)
	writeTemplateFile(t, templatesDir, filepath.Join(skillTemplatesDir, "core-skill", "SKILL.md"), "core")
	writeTemplateFile(t, templatesDir, filepath.Join(sessionTemplatesDir, "dev_server.toml"), `name = "Dev Server"
command = "mix phx.server"
`)
	writeTemplateFile(t, templatesDir, filepath.Join(sessionTemplatesDir, "deploy.toml"), `name = "Deploy"
command = "git push"
`)
	writeTemplateFile(t, templatesDir, filepath.Join(sessionTemplatesDir, "watch_css.toml"), `name = "Watch CSS"
command = "pnpm watch"
`)
	writeTemplateFile(t, templatesDir, filepath.Join(sessionTemplatesDir, "logs.toml"), `name = "Logs"
command = "tail -f log"
`)
	t.Setenv(templatesEnvVar, templatesDir)

	rootDir := t.TempDir()
	// First init
	if err := runInit([]string{"--type", "local_sessions", "--path", rootDir}); err != nil {
		t.Fatalf("run init: %v", err)
	}

	// Add local config with append
	localConfig := filepath.Join(rootDir, mavuDirName, localConfigFilename)
	localContent := `type = "local_sessions"
autostart_sessions_append = ["watch_css"]
ondemand_sessions_append = ["logs"]
`
	if err := os.WriteFile(localConfig, []byte(localContent), 0o644); err != nil {
		t.Fatalf("write local config: %v", err)
	}

	// Run update
	if err := runUpdate([]string{"--path", rootDir}); err != nil {
		t.Fatalf("run update: %v", err)
	}

	tasksPath := filepath.Join(rootDir, ".vscode", "tasks.json")
	var tasksJSON map[string]any
	loadJSON(t, tasksPath, &tasksJSON)

	tasks := tasksJSON["tasks"].([]any)
	tasksByLabel := make(map[string]bool)
	for _, task := range tasks {
		taskMap := task.(map[string]any)
		tasksByLabel[taskMap["label"].(string)] = true
	}

	// Check all sessions are present
	expectedTasks := []string{"Dev Server", "Watch CSS", "Deploy", "Logs", "__ Start Default Terminal Sessions"}
	for _, label := range expectedTasks {
		if !tasksByLabel[label] {
			t.Fatalf("expected task %q not found", label)
		}
	}
}

func TestSessionsLocalConfigOverride(t *testing.T) {
	templatesDir := createTemplateRoot(t)
	writeTemplateFile(t, templatesDir, filepath.Join(projectTypesDir, "override_sessions.toml"), `name = "Override Sessions"
skills = ["core-skill"]
autostart_sessions = ["dev_server", "watch_css"]
`)
	writeTemplateFile(t, templatesDir, filepath.Join(skillTemplatesDir, "core-skill", "SKILL.md"), "core")
	writeTemplateFile(t, templatesDir, filepath.Join(sessionTemplatesDir, "dev_server.toml"), `name = "Dev Server"
command = "mix phx.server"
`)
	writeTemplateFile(t, templatesDir, filepath.Join(sessionTemplatesDir, "watch_css.toml"), `name = "Watch CSS"
command = "pnpm watch"
`)
	writeTemplateFile(t, templatesDir, filepath.Join(sessionTemplatesDir, "custom.toml"), `name = "Custom"
command = "custom cmd"
`)
	t.Setenv(templatesEnvVar, templatesDir)

	rootDir := t.TempDir()
	if err := runInit([]string{"--type", "override_sessions", "--path", rootDir}); err != nil {
		t.Fatalf("run init: %v", err)
	}

	// Override autostart sessions completely
	localConfig := filepath.Join(rootDir, mavuDirName, localConfigFilename)
	localContent := `type = "override_sessions"
autostart_sessions = ["custom"]
`
	if err := os.WriteFile(localConfig, []byte(localContent), 0o644); err != nil {
		t.Fatalf("write local config: %v", err)
	}

	if err := runUpdate([]string{"--path", rootDir}); err != nil {
		t.Fatalf("run update: %v", err)
	}

	tasksPath := filepath.Join(rootDir, ".vscode", "tasks.json")
	var tasksJSON map[string]any
	loadJSON(t, tasksPath, &tasksJSON)

	tasks := tasksJSON["tasks"].([]any)
	tasksByLabel := make(map[string]bool)
	for _, task := range tasks {
		taskMap := task.(map[string]any)
		tasksByLabel[taskMap["label"].(string)] = true
	}

	// Only custom should be present (override replaces entirely)
	if !tasksByLabel["Custom"] {
		t.Fatal("expected Custom task")
	}
	if tasksByLabel["Dev Server"] {
		t.Fatal("Dev Server should be overridden")
	}
	if tasksByLabel["Watch CSS"] {
		t.Fatal("Watch CSS should be overridden")
	}
}
