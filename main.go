package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const (
	projectTypesDir        = "project_types"
	projectTypesGlob       = "project_types/*.toml"
	snippetsDir            = "snippets"
	skillTemplatesDir      = "skill_templates"
	commandTemplatesDir    = "command_templates"
	commandTemplateExt     = ".md"
	mcpTemplatesDir        = "mcp_templates"
	mcpTemplateExt         = ".mcp.json"
	sessionTemplatesDir    = "session_templates"
	sessionTemplateExt     = ".toml"
	templatesEnvVar        = "MAVU_LLM_TEMPLATES_DIR"
	agentsFilename         = "AGENTS.md"
	claudeFilename         = "CLAUDE.md"
	mavuDirName            = ".mavu"
	localConfigFilename    = "config.toml"
	legacyConfigFilename   = ".mavu_llm.toml"
	opencodeConfigFilename = "opencode.json"
	mcpConfigFilename      = ".mcp.json"
	usageRulesConfigPath   = "lib/_mavubit/essentials/config/essentials_mix.exs"
	usageRulesFilename     = "USAGE_RULES.md"
	usageRulesOutputPath   = "USAGE_RULES.md"
	version                = "0.2.0"
	defaultFilePermission  = 0o644
	defaultDirPermission   = 0o755
)

var verboseOutput bool

type ProjectConfig struct {
	Name              string     `toml:"name"`
	Description       string     `toml:"description"`
	Skills            []string   `toml:"skills"`
	Commands          []string   `toml:"commands"`
	Mcps              []string   `toml:"mcps"`
	Snippets          []string   `toml:"snippets"`
	AutostartSessions []string   `toml:"autostart_sessions"`
	OndemandSessions  []string   `toml:"ondemand_sessions"`
	Claude            ToolConfig `toml:"claude"`
	Codex             ToolConfig `toml:"codex"`
	Agents            ToolConfig `toml:"agents"`
}

type SessionTemplate struct {
	Name    string `toml:"name"`
	Command string `toml:"command"`
}

type ToolConfig struct {
	Skills          []string `toml:"skills"`
	Commands        []string `toml:"commands"`
	Mcps            []string `toml:"mcps"`
	Snippets        []string `toml:"snippets"`
	SnippetsAppend  []string `toml:"snippets_append"`
	SnippetsPrepend []string `toml:"snippets_prepend"`
}

type ResolvedToolConfig struct {
	Skills   []string
	Commands []string
	Mcps     []string
	Snippets []string
}

type ProjectType struct {
	ID     string
	Config ProjectConfig
	Root   string
}

type ProjectTypeFile struct {
	Type                    string   `toml:"type"`
	AutostartSessions       []string `toml:"autostart_sessions"`
	AutostartSessionsAppend []string `toml:"autostart_sessions_append"`
	OndemandSessions        []string `toml:"ondemand_sessions"`
	OndemandSessionsAppend  []string `toml:"ondemand_sessions_append"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	switch command {
	case "types":
		if err := listProjectTypes(); err != nil {
			exitWithError(err)
		}
	case "init":
		if err := runInit(os.Args[2:]); err != nil {
			exitWithError(err)
		}
	case "update":
		if err := runUpdate(os.Args[2:]); err != nil {
			exitWithError(err)
		}
	case "template-paths":
		if err := listTemplatePaths(); err != nil {
			exitWithError(err)
		}
	case "version", "--version", "-v":
		printVersion()
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("mavu-llm - manage LLM setup templates")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  mavu-llm types")
	fmt.Println("  mavu-llm init --type <project-type> [--path <dir>] [--verbose]")
	fmt.Println("  mavu-llm update [--path <dir>] [--verbose]")
	fmt.Println("  mavu-llm template-paths")
	fmt.Println("  mavu-llm version")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  types           List available project types")
	fmt.Println("  init            Create .codex/AGENTS.md, .claude/CLAUDE.md, and skills directories")
	fmt.Println("  update          Re-run setup using stored project type")
	fmt.Println("  template-paths  Show template search paths")
	fmt.Println("  version         Show current version")
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  Top-level skills/commands/snippets apply to codex and claude unless overridden.")
	fmt.Println("  Use snippets_prepend/snippets_append for tool-specific additions.")
	fmt.Printf("  Set %s to the template root directory.\n", templatesEnvVar)
}

func printVersion() {
	fmt.Printf("mavu-llm %s\n", version)
}

func listProjectTypes() error {
	projectTypes, err := loadProjectTypes()
	if err != nil {
		return err
	}

	ids := sortedIDs(projectTypes)
	if len(ids) == 0 {
		return errors.New("no project types found")
	}

	fmt.Println("Available project types:")
	for _, id := range ids {
		projectType := projectTypes[id]
		cfg := projectType.Config
		label := cfg.Name
		if label == "" {
			label = id
		}
		if cfg.Description != "" {
			fmt.Printf("  %s - %s (%s)\n", id, label, cfg.Description)
			continue
		}
		fmt.Printf("  %s - %s\n", id, label)
	}
	return nil
}

func listTemplatePaths() error {
	roots, err := templateRoots()
	if err != nil {
		return err
	}
	if len(roots) == 0 {
		return errors.New("no template roots found")
	}
	if env := strings.TrimSpace(os.Getenv(templatesEnvVar)); env != "" {
		fmt.Printf("Using %s=%s\n", templatesEnvVar, env)
	}

	// Show local project skills info
	cwd, _ := os.Getwd()
	localSkillsPath := filepath.Join(cwd, mavuDirName, skillTemplatesDir)
	if _, err := os.Stat(localSkillsPath); err == nil {
		fmt.Printf("Local project skills: %s (takes precedence)\n", localSkillsPath)
	}
	localSessionsPath := filepath.Join(cwd, mavuDirName, sessionTemplatesDir)
	if _, err := os.Stat(localSessionsPath); err == nil {
		fmt.Printf("Local project sessions: %s (takes precedence)\n", localSessionsPath)
	}

	fmt.Println("Template roots:")
	for _, root := range roots {
		fmt.Printf("  %s\n", root)
		fmt.Printf("    project_types: %s\n", filepath.Join(root, projectTypesDir))
		fmt.Printf("    snippets: %s\n", filepath.Join(root, snippetsDir))
		fmt.Printf("    skill_templates: %s\n", filepath.Join(root, skillTemplatesDir))
		fmt.Printf("    command_templates: %s\n", filepath.Join(root, commandTemplatesDir))
		fmt.Printf("    mcp_templates: %s\n", filepath.Join(root, mcpTemplatesDir))
		fmt.Printf("    session_templates: %s\n", filepath.Join(root, sessionTemplatesDir))
	}
	return nil
}

func runInit(args []string) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	projectTypeFlag := flags.String("type", "", "Project type ID")
	pathFlag := flags.String("path", "", "Target directory (defaults to cwd)")
	verboseFlag := flags.Bool("verbose", false, "Enable verbose logging")
	if err := flags.Parse(args); err != nil {
		return err
	}
	verboseOutput = *verboseFlag

	projectTypes, err := loadProjectTypes()
	if err != nil {
		return err
	}

	projectTypeID := strings.TrimSpace(*projectTypeFlag)
	if projectTypeID == "" {
		selection, err := promptProjectType(projectTypes)
		if err != nil {
			return err
		}
		projectTypeID = selection
	}

	rootDir := strings.TrimSpace(*pathFlag)
	if rootDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		rootDir = cwd
	}

	projectType, ok := projectTypes[projectTypeID]
	if !ok {
		return fmt.Errorf("unknown project type: %s", projectTypeID)
	}
	// For init, try to read existing local config (may not exist yet)
	localConfig, _ := readLocalConfig(rootDir)
	return runSetup(rootDir, projectType, localConfig, "Initialized")
}

func runUpdate(args []string) error {
	flags := flag.NewFlagSet("update", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	pathFlag := flags.String("path", "", "Target directory (defaults to cwd)")
	verboseFlag := flags.Bool("verbose", false, "Enable verbose logging")
	if err := flags.Parse(args); err != nil {
		return err
	}
	verboseOutput = *verboseFlag

	rootDir := strings.TrimSpace(*pathFlag)
	if rootDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		rootDir = cwd
	}

	localConfig, err := readLocalConfig(rootDir)
	if err != nil {
		return err
	}

	projectTypes, err := loadProjectTypes()
	if err != nil {
		return err
	}

	projectType, ok := projectTypes[localConfig.Type]
	if !ok {
		return fmt.Errorf("unknown project type: %s", localConfig.Type)
	}
	return runSetup(rootDir, projectType, localConfig, "Updated")
}

func runSetup(rootDir string, projectType ProjectType, localConfig ProjectTypeFile, action string) error {
	cfg := projectType.Config
	templateRoot := projectType.Root
	codexConfig := resolveCodexConfig(cfg)
	claudeConfig := resolveToolConfig(cfg, cfg.Claude)

	if err := createSkillDirs(rootDir, templateRoot, codexConfig, claudeConfig); err != nil {
		return err
	}

	if err := createCommandDirs(rootDir, templateRoot, codexConfig, claudeConfig); err != nil {
		return err
	}

	// Merge sessions: local config can override or append to project type sessions
	autostartSessions := mergeSessionsConfig(cfg.AutostartSessions, localConfig.AutostartSessions, localConfig.AutostartSessionsAppend)
	ondemandSessions := mergeSessionsConfig(cfg.OndemandSessions, localConfig.OndemandSessions, localConfig.OndemandSessionsAppend)

	if err := writeSessionTasks(rootDir, templateRoot, autostartSessions, ondemandSessions); err != nil {
		return err
	}

	mcpNames := uniqueOrdered(codexConfig.Mcps, claudeConfig.Mcps)
	if len(mcpNames) > 0 {
		mcpEntries, err := loadMcpEntries(rootDir, templateRoot, mcpNames)
		if err != nil {
			return err
		}
		if err := writeOpenCodeConfig(rootDir, mcpEntries); err != nil {
			return err
		}
		if err := writeMcpConfig(rootDir, mcpEntries); err != nil {
			return err
		}
	}

	if err := writeRootDocs(rootDir, templateRoot, codexConfig, claudeConfig); err != nil {
		return err
	}

	if err := writeProjectTypeFile(rootDir, projectType.ID); err != nil {
		return err
	}

	if err := runUsageRulesSync(rootDir); err != nil {
		return err
	}

	fmt.Printf("%s %s in %s\n", action, projectType.ID, rootDir)
	return nil
}

func runUsageRulesSync(rootDir string) error {
	configPath := filepath.Join(rootDir, usageRulesConfigPath)
	if _, err := os.Stat(configPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	args := []string{"usage_rules.sync", usageRulesOutputPath, "--all", "--link-to-folder", "deps", "--yes"}
	fmt.Printf("Running: mix %s\n", strings.Join(args, " "))
	cmd := exec.Command("mix", args...)
	cmd.Dir = rootDir
	output, err := cmd.CombinedOutput()
	tail := tailLines(string(output), 10)
	if err != nil {
		message := strings.TrimSpace(tail)
		if message != "" {
			return fmt.Errorf("usage_rules.sync failed: %w\n%s", err, message)
		}
		return fmt.Errorf("usage_rules.sync failed: %w", err)
	}
	if strings.TrimSpace(tail) != "" {
		fmt.Printf("%s\n", tail)
	}
	return appendUsageRules(rootDir)
}

func appendUsageRules(rootDir string) error {
	rulesPath := filepath.Join(rootDir, usageRulesOutputPath)
	data, err := os.ReadFile(rulesPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	rules := strings.TrimSpace(string(data))
	if rules == "" {
		return os.Remove(rulesPath)
	}

	rules = strings.ReplaceAll(rules, "(deps/", "(../deps/")

	targets := []string{
		filepath.Join(rootDir, ".claude", claudeFilename),
		filepath.Join(rootDir, ".codex", agentsFilename),
	}
	for _, target := range targets {
		if err := appendWithSeparator(target, rules); err != nil {
			return err
		}
	}

	return os.Remove(rulesPath)
}

func appendWithSeparator(path, content string) error {
	existing, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	existingText := string(existing)
	builder := strings.Builder{}
	builder.WriteString(existingText)
	if len(existingText) > 0 && !strings.HasSuffix(existingText, "\n") {
		builder.WriteString("\n")
	}
	if len(existingText) > 0 {
		builder.WriteString("\n")
	}
	builder.WriteString(content)
	builder.WriteString("\n")

	return writeFile(path, []byte(builder.String()))
}

func tailLines(text string, limit int) string {
	cleaned := strings.TrimRight(text, "\n")
	if cleaned == "" || limit <= 0 {
		return ""
	}
	lines := strings.Split(cleaned, "\n")
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	return strings.Join(lines, "\n")
}

func resolveCodexConfig(cfg ProjectConfig) ResolvedToolConfig {
	if toolConfigEmpty(cfg.Codex) && !toolConfigEmpty(cfg.Agents) {
		return resolveToolConfig(cfg, cfg.Agents)
	}
	return resolveToolConfig(cfg, cfg.Codex)
}

func resolveToolConfig(cfg ProjectConfig, tool ToolConfig) ResolvedToolConfig {
	return ResolvedToolConfig{
		Skills:   resolveSkills(cfg.Skills, tool),
		Commands: resolveCommands(cfg.Commands, tool),
		Mcps:     resolveMcps(cfg.Mcps, tool),
		Snippets: resolveSnippets(cfg.Snippets, tool),
	}
}

func resolveSkills(defaults []string, tool ToolConfig) []string {
	toolSkills := normalizedList(tool.Skills)
	if len(toolSkills) > 0 {
		return toolSkills
	}
	return normalizedList(defaults)
}

func resolveCommands(defaults []string, tool ToolConfig) []string {
	toolCommands := normalizedList(tool.Commands)
	if len(toolCommands) > 0 {
		return toolCommands
	}
	return normalizedList(defaults)
}

func resolveMcps(defaults []string, tool ToolConfig) []string {
	toolMcps := normalizedList(tool.Mcps)
	if len(toolMcps) > 0 {
		return toolMcps
	}
	return normalizedList(defaults)
}

func resolveSnippets(defaults []string, tool ToolConfig) []string {
	toolSnippets := normalizedList(tool.Snippets)
	if len(toolSnippets) > 0 {
		return toolSnippets
	}
	resolved := normalizedList(defaults)
	prepend := normalizedList(tool.SnippetsPrepend)
	appendList := normalizedList(tool.SnippetsAppend)
	if len(prepend) > 0 {
		resolved = append(prepend, resolved...)
	}
	if len(appendList) > 0 {
		resolved = append(resolved, appendList...)
	}
	return resolved
}

func toolConfigEmpty(cfg ToolConfig) bool {
	return len(cfg.Skills) == 0 && len(cfg.Commands) == 0 && len(cfg.Mcps) == 0 && len(cfg.Snippets) == 0 && len(cfg.SnippetsAppend) == 0 && len(cfg.SnippetsPrepend) == 0
}

func normalizedList(items []string) []string {
	cleaned := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		cleaned = append(cleaned, trimmed)
	}
	return cleaned
}

func uniqueOrdered(lists ...[]string) []string {
	seen := make(map[string]struct{})
	var combined []string
	for _, list := range lists {
		for _, item := range list {
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			combined = append(combined, item)
		}
	}
	return combined
}

// mergeSessionsConfig merges session configuration.
// If override is non-empty, it replaces defaults entirely.
// Otherwise, appendList is appended to defaults.
func mergeSessionsConfig(defaults, override, appendList []string) []string {
	override = normalizedList(override)
	if len(override) > 0 {
		return override
	}
	defaults = normalizedList(defaults)
	appendList = normalizedList(appendList)
	return uniqueOrdered(defaults, appendList)
}

func commandTemplateFilename(name string) string {
	trimmed := strings.TrimSpace(name)
	if strings.HasSuffix(trimmed, commandTemplateExt) {
		return trimmed
	}
	return trimmed + commandTemplateExt
}

func sessionTemplateFilename(name string) string {
	trimmed := strings.TrimSpace(name)
	if strings.HasSuffix(trimmed, sessionTemplateExt) {
		return trimmed
	}
	return trimmed + sessionTemplateExt
}

func mcpTemplateFilename(name string) string {
	trimmed := strings.TrimSpace(name)
	if strings.HasSuffix(trimmed, mcpTemplateExt) {
		return trimmed
	}
	return trimmed + mcpTemplateExt
}

func normalizeMcpServer(entry map[string]any) map[string]any {
	normalized := make(map[string]any, len(entry))
	for key, value := range entry {
		if key == "enabled" {
			continue
		}
		normalized[key] = value
	}
	if value, ok := normalized["type"].(string); ok && value == "remote" {
		normalized["type"] = "http"
	}
	return normalized
}

func loadMcpTemplate(path string) (map[string]any, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	expanded, missing := expandEnvVars(string(data))
	var entry map[string]any
	if err := json.Unmarshal([]byte(expanded), &entry); err != nil {
		return nil, missing, fmt.Errorf("parse %s: %w", path, err)
	}
	return entry, missing, nil
}

// loadLocalMcpEntries loads MCP entries from .mavu/mcp.json if present.
// Returns map of MCP server configs, missing env vars, and error.
func loadLocalMcpEntries(rootDir string) (map[string]any, []string, error) {
	localPath := filepath.Join(rootDir, mavuDirName, "mcp.json")

	// Check if file exists
	if _, err := os.Stat(localPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, nil // No local MCPs - not an error
		}
		return nil, nil, err
	}

	// Read and parse similar to loadMcpTemplate
	data, err := os.ReadFile(localPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", localPath, err)
	}

	// Expand env vars
	expanded, missing := expandEnvVars(string(data))

	// Parse JSON
	var entries map[string]any
	if err := json.Unmarshal([]byte(expanded), &entries); err != nil {
		return nil, missing, fmt.Errorf("parse %s: %w", localPath, err)
	}

	// Support both flat format and wrapped format (with "mcpServers" key)
	if mcpServers, ok := entries["mcpServers"].(map[string]any); ok {
		return mcpServers, missing, nil
	}

	return entries, missing, nil
}

func expandEnvVars(input string) (string, []string) {
	missing := make(map[string]struct{})
	expanded := os.Expand(input, func(key string) string {
		if value, ok := os.LookupEnv(key); ok {
			return value
		}
		missing[key] = struct{}{}
		return fmt.Sprintf("${%s}", key)
	})
	if len(missing) == 0 {
		return expanded, nil
	}
	missingKeys := make([]string, 0, len(missing))
	for key := range missing {
		missingKeys = append(missingKeys, key)
	}
	sort.Strings(missingKeys)
	return expanded, missingKeys
}

func warnMissingEnv(path string, missing []string) {
	if len(missing) == 0 {
		return
	}
	fmt.Printf("Warning: %s missing env vars: %s\n", path, strings.Join(missing, ", "))
}

func loadProjectTypes() (map[string]ProjectType, error) {
	roots, err := templateRoots()
	if err != nil {
		return nil, err
	}
	if len(roots) == 0 {
		return nil, fmt.Errorf("no template roots found (set %s or ensure %s exists)", templatesEnvVar, projectTypesDir)
	}

	projectTypes := make(map[string]ProjectType)
	for _, root := range roots {
		if err := loadProjectTypesFromRoot(root, projectTypes); err != nil {
			return nil, err
		}
	}
	if len(projectTypes) == 0 {
		return nil, errors.New("no project types found")
	}
	return projectTypes, nil
}

func loadProjectTypesFromRoot(root string, projectTypes map[string]ProjectType) error {
	glob := filepath.Join(root, projectTypesGlob)
	matches, err := filepath.Glob(glob)
	if err != nil {
		return err
	}
	for _, match := range matches {
		data, err := os.ReadFile(match)
		if err != nil {
			return err
		}
		if err := parseProjectTypeFile(match, data, root, projectTypes); err != nil {
			return err
		}
	}
	return nil
}

func parseProjectTypeFile(path string, data []byte, root string, projectTypes map[string]ProjectType) error {
	var cfg ProjectConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	id := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	projectTypes[id] = ProjectType{ID: id, Config: cfg, Root: root}
	return nil
}

func templateRoots() ([]string, error) {
	if env := strings.TrimSpace(os.Getenv(templatesEnvVar)); env != "" {
		if _, err := os.Stat(filepath.Join(env, projectTypesDir)); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("%s is missing %s", env, projectTypesDir)
			}
			return nil, err
		}
		return []string{env}, nil
	}

	var candidates []string
	exePath, err := os.Executable()
	if err == nil {
		resolved, err := filepath.EvalSymlinks(exePath)
		if err == nil {
			resolvedDir := filepath.Dir(resolved)
			if resolvedDir != "" {
				candidates = append(candidates, resolvedDir, filepath.Join(resolvedDir, "templates"))
			}
		}
		exeDir := filepath.Dir(exePath)
		if exeDir != "" {
			candidates = append(candidates, exeDir, filepath.Join(exeDir, "templates"))
		}
	}
	cwd, err := os.Getwd()
	if err == nil {
		candidates = append(candidates, cwd, filepath.Join(cwd, "templates"))
	}

	seen := make(map[string]struct{}, len(candidates))
	var roots []string
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		if _, err := os.Stat(filepath.Join(candidate, projectTypesDir)); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}
		seen[candidate] = struct{}{}
		roots = append(roots, candidate)
	}
	return roots, nil
}

// projectConfigPath returns the path to the project config file.
// Checks .mavu/config.toml first (new), falls back to .mavu_llm.toml (legacy).
func projectConfigPath(rootDir string) (string, bool, error) {
	// Check new location first
	newPath := filepath.Join(rootDir, mavuDirName, localConfigFilename)
	if _, err := os.Stat(newPath); err == nil {
		return newPath, false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", false, err
	}

	// Fall back to legacy location
	legacyPath := filepath.Join(rootDir, legacyConfigFilename)
	if _, err := os.Stat(legacyPath); err == nil {
		return legacyPath, true, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", false, err
	}

	return "", false, fmt.Errorf("no project config found (checked %s and %s)", newPath, legacyPath)
}

func writeProjectTypeFile(rootDir, projectTypeID string) error {
	// Create .mavu directory if it doesn't exist
	mavuDir := filepath.Join(rootDir, mavuDirName)
	if err := os.MkdirAll(mavuDir, defaultDirPermission); err != nil {
		return err
	}

	// Write to new location
	path := filepath.Join(mavuDir, localConfigFilename)
	payload, err := toml.Marshal(ProjectTypeFile{Type: projectTypeID})
	if err != nil {
		return err
	}
	return writeFile(path, payload)
}

func readProjectTypeFile(rootDir string) (string, error) {
	cfg, err := readLocalConfig(rootDir)
	if err != nil {
		return "", err
	}
	return cfg.Type, nil
}

func readLocalConfig(rootDir string) (ProjectTypeFile, error) {
	path, isLegacy, err := projectConfigPath(rootDir)
	if err != nil {
		return ProjectTypeFile{}, err
	}
	if isLegacy {
		fmt.Printf("Note: Found legacy config at %s (will migrate to %s on next update)\n",
			legacyConfigFilename, filepath.Join(mavuDirName, localConfigFilename))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ProjectTypeFile{}, fmt.Errorf("read %s: %w", path, err)
	}
	var cfg ProjectTypeFile
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return ProjectTypeFile{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if strings.TrimSpace(cfg.Type) == "" {
		return ProjectTypeFile{}, fmt.Errorf("%s missing type", path)
	}
	cfg.Type = strings.TrimSpace(cfg.Type)
	return cfg, nil
}

func promptProjectType(projectTypes map[string]ProjectType) (string, error) {
	ids := sortedIDs(projectTypes)
	if len(ids) == 0 {
		return "", errors.New("no project types available")
	}

	fmt.Println("Select a project type:")
	for i, id := range ids {
		cfg := projectTypes[id].Config
		label := cfg.Name
		if label == "" {
			label = id
		}
		fmt.Printf("  %d) %s (%s)\n", i+1, id, label)
	}

	fmt.Print("Enter number: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.TrimSpace(input)
	index, err := strconv.Atoi(input)
	if err != nil || index < 1 || index > len(ids) {
		return "", errors.New("invalid selection")
	}
	return ids[index-1], nil
}

// discoverLocalSkills scans .mavu/skill_templates/ and returns all local skill names.
func discoverLocalSkills(rootDir string) ([]string, error) {
	localSkillsDir := filepath.Join(rootDir, mavuDirName, skillTemplatesDir)
	entries, err := os.ReadDir(localSkillsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var skills []string
	for _, entry := range entries {
		if entry.IsDir() {
			skills = append(skills, entry.Name())
		}
	}
	return skills, nil
}

// mergeSkills combines configured skills with local skills.
// Local skills are added to the list, maintaining order (configured first, then local-only).
func mergeSkills(configured, local []string) []string {
	if len(local) == 0 {
		return configured
	}

	seen := make(map[string]struct{}, len(configured))
	for _, skill := range configured {
		seen[skill] = struct{}{}
	}

	merged := make([]string, len(configured))
	copy(merged, configured)

	for _, skill := range local {
		if _, ok := seen[skill]; !ok {
			merged = append(merged, skill)
		}
	}
	return merged
}

// findSkillTemplatePath finds the skill template directory.
// Returns local project skill if found, otherwise returns global template skill.
func findSkillTemplatePath(rootDir, templateRoot, skillName string) (string, error) {
	// Check local project skill first (takes precedence)
	localPath := filepath.Join(rootDir, mavuDirName, skillTemplatesDir, skillName)
	if info, err := os.Stat(localPath); err == nil && info.IsDir() {
		return localPath, nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	// Fall back to global template skill
	globalPath := filepath.Join(templateRoot, skillTemplatesDir, skillName)
	if info, err := os.Stat(globalPath); err == nil && info.IsDir() {
		return globalPath, nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	return "", fmt.Errorf("skill template not found: %s (checked %s and %s)",
		skillName, localPath, globalPath)
}

func findSessionTemplatePath(rootDir, templateRoot, sessionName string) (string, error) {
	filename := sessionTemplateFilename(sessionName)
	if filename == "" {
		return "", errors.New("session template name is empty")
	}
	localPath := filepath.Join(rootDir, mavuDirName, sessionTemplatesDir, filename)
	if info, err := os.Stat(localPath); err == nil && !info.IsDir() {
		return localPath, nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	globalPath := filepath.Join(templateRoot, sessionTemplatesDir, filename)
	if info, err := os.Stat(globalPath); err == nil && !info.IsDir() {
		return globalPath, nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	return "", fmt.Errorf("session template not found: %s (checked %s and %s)",
		filename, localPath, globalPath)
}

func createSkillDirs(rootDir, templateRoot string, codexConfig, claudeConfig ResolvedToolConfig) error {
	// Discover local skills from .mavu/skill_templates/
	localSkills, err := discoverLocalSkills(rootDir)
	if err != nil {
		return fmt.Errorf("discover local skills: %w", err)
	}

	targets := []struct {
		label  string
		path   string
		skills []string
	}{
		{
			label:  "codex",
			path:   filepath.Join(rootDir, ".codex", "skills"),
			skills: mergeSkills(codexConfig.Skills, localSkills),
		},
		{
			label:  "claude",
			path:   filepath.Join(rootDir, ".claude", "skills"),
			skills: mergeSkills(claudeConfig.Skills, localSkills),
		},
	}

	for _, target := range targets {
		if len(target.skills) == 0 {
			return fmt.Errorf("project type has no skills configured for %s", target.label)
		}
		desiredSkills := make(map[string]struct{}, len(target.skills))
		for _, skill := range target.skills {
			desiredSkills[skill] = struct{}{}
		}
		if err := os.MkdirAll(target.path, defaultDirPermission); err != nil {
			return err
		}
		entries, err := os.ReadDir(target.path)
		if err != nil {
			return err
		}
		var toRemove []string
		for _, entry := range entries {
			if _, ok := desiredSkills[entry.Name()]; ok {
				continue
			}
			toRemove = append(toRemove, entry.Name())
		}
		if len(toRemove) > 0 {
			fmt.Printf("Warning: %d skill(s) not in config will be removed from %s: %s\n", len(toRemove), target.label, strings.Join(toRemove, ", "))
			fmt.Print("Move to .mavu/skill_templates/ to keep them? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))

			if input == "y" || input == "yes" {
				localSkillsDir := filepath.Join(rootDir, mavuDirName, skillTemplatesDir)
				if err := os.MkdirAll(localSkillsDir, defaultDirPermission); err != nil {
					return err
				}
				for _, name := range toRemove {
					src := filepath.Join(target.path, name)
					dst := filepath.Join(localSkillsDir, name)
					if _, err := os.Stat(dst); err == nil {
						fmt.Printf("  Skipping %s (already exists in .mavu/skill_templates/)\n", name)
						continue
					}
					if err := os.Rename(src, dst); err != nil {
						return fmt.Errorf("move %s: %w", name, err)
					}
					fmt.Printf("  Moved %s to .mavu/skill_templates/\n", name)
				}
				// Re-discover local skills and update desired skills
				localSkills, err = discoverLocalSkills(rootDir)
				if err != nil {
					return fmt.Errorf("re-discover local skills: %w", err)
				}
				for _, skill := range localSkills {
					desiredSkills[skill] = struct{}{}
				}
				toRemove = nil // Clear since we moved them
			}
		}
		for _, name := range toRemove {
			if err := os.RemoveAll(filepath.Join(target.path, name)); err != nil {
				return err
			}
		}

		for _, skill := range target.skills {
			sourcePath, err := findSkillTemplatePath(rootDir, templateRoot, skill)
			if err != nil {
				return err
			}
			targetPath := filepath.Join(target.path, skill)
			if err := os.MkdirAll(targetPath, defaultDirPermission); err != nil {
				return err
			}
			if err := copyDir(sourcePath, targetPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func createCommandDirs(rootDir, templateRoot string, codexConfig, claudeConfig ResolvedToolConfig) error {
	targets := []struct {
		path     string
		commands []string
	}{
		{
			path:     filepath.Join(rootDir, ".opencode", "command"),
			commands: codexConfig.Commands,
		},
		{
			path:     filepath.Join(rootDir, ".claude", "commands"),
			commands: claudeConfig.Commands,
		},
	}

	for _, target := range targets {
		if len(target.commands) == 0 {
			continue
		}
		normalizedCommands := make([]string, 0, len(target.commands))
		desiredCommands := make(map[string]struct{}, len(target.commands))
		for _, command := range target.commands {
			filename := commandTemplateFilename(command)
			if filename == "" {
				continue
			}
			if _, ok := desiredCommands[filename]; ok {
				continue
			}
			desiredCommands[filename] = struct{}{}
			normalizedCommands = append(normalizedCommands, filename)
		}
		if len(normalizedCommands) == 0 {
			continue
		}
		if err := os.MkdirAll(target.path, defaultDirPermission); err != nil {
			return err
		}
		entries, err := os.ReadDir(target.path)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if _, ok := desiredCommands[entry.Name()]; ok {
				continue
			}
			if err := os.RemoveAll(filepath.Join(target.path, entry.Name())); err != nil {
				return err
			}
		}

		for _, command := range normalizedCommands {
			sourcePath := filepath.Join(templateRoot, commandTemplatesDir, command)
			if _, err := os.Stat(sourcePath); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("command template not found: %s", command)
				}
				return err
			}
			targetPath := filepath.Join(target.path, command)
			if err := copyFile(sourcePath, targetPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeSessionTasks(rootDir, templateRoot string, autostartNames, ondemandNames []string) error {
	autostart := uniqueOrdered(normalizedList(autostartNames))
	ondemand := uniqueOrdered(normalizedList(ondemandNames))
	allSessions := uniqueOrdered(autostart, ondemand)

	if len(allSessions) == 0 {
		return nil
	}

	autostartSet := make(map[string]struct{}, len(autostart))
	for _, name := range autostart {
		autostartSet[name] = struct{}{}
	}

	var tasks []map[string]any
	for _, sessionName := range allSessions {
		templatePath, err := findSessionTemplatePath(rootDir, templateRoot, sessionName)
		if err != nil {
			return err
		}

		session, err := loadSessionTemplate(templatePath)
		if err != nil {
			return err
		}

		_, isAutostart := autostartSet[sessionName]
		task := buildVSCodeTask(sessionName, session, isAutostart)
		tasks = append(tasks, task)
	}

	// Add compound task for autostart sessions if there are any
	if len(autostart) > 0 {
		var autostartLabels []string
		for _, sessionName := range autostart {
			templatePath, err := findSessionTemplatePath(rootDir, templateRoot, sessionName)
			if err == nil {
				session, err := loadSessionTemplate(templatePath)
				if err == nil {
					label := session.Name
					if label == "" {
						label = sessionName
					}
					autostartLabels = append(autostartLabels, label)
				}
			}
		}
		if len(autostartLabels) > 0 {
			compoundTask := map[string]any{
				"label":          "__ Start Default Terminal Sessions",
				"dependsOn":      autostartLabels,
				"problemMatcher": []any{},
			}
			tasks = append(tasks, compoundTask)
		}
	}

	tasksJSON := map[string]any{
		"version": "2.0.0",
		"tasks":   tasks,
	}

	targetPath := filepath.Join(rootDir, ".vscode", "tasks.json")
	if err := os.MkdirAll(filepath.Dir(targetPath), defaultDirPermission); err != nil {
		return err
	}

	payload, err := json.MarshalIndent(tasksJSON, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return writeFile(targetPath, payload)
}

func loadSessionTemplate(path string) (SessionTemplate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SessionTemplate{}, err
	}
	var session SessionTemplate
	if err := toml.Unmarshal(data, &session); err != nil {
		return SessionTemplate{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if session.Command == "" {
		return SessionTemplate{}, fmt.Errorf("session template missing command: %s", path)
	}
	return session, nil
}

func buildVSCodeTask(id string, session SessionTemplate, isAutostart bool) map[string]any {
	label := session.Name
	if label == "" {
		label = id
	}

	command := session.Command + "; exec $SHELL"

	task := map[string]any{
		"label":        label,
		"type":         "shell",
		"command":      command,
		"isBackground": true,
		"problemMatcher": []any{},
		"options": map[string]any{
			"shell": map[string]any{
				"args": []string{"-i", "-l"},
			},
		},
		"presentation": map[string]any{
			"panel": "dedicated",
			"close": false,
			"focus": true,
		},
	}

	if isAutostart {
		task["runOptions"] = map[string]any{
			"runOn": "folderOpen",
		}
	}

	return task
}

type openCodeConfig struct {
	Mcp    map[string]any `json:"mcp"`
	Schema string         `json:"$schema"`
}

type mcpConfig struct {
	McpServers map[string]any `json:"mcpServers"`
}

func loadMcpEntries(rootDir, templateRoot string, mcpNames []string) (map[string]any, error) {
	// Load local MCPs first (these take precedence)
	localEntries, localMissing, err := loadLocalMcpEntries(rootDir)
	if err != nil {
		return nil, fmt.Errorf("load local mcps: %w", err)
	}
	if len(localMissing) > 0 {
		warnMissingEnv(filepath.Join(rootDir, mavuDirName, "mcp.json"), localMissing)
	}

	// Start with local entries (these have precedence)
	mcpEntries := make(map[string]any, len(localEntries))
	for key, value := range localEntries {
		mcpEntries[key] = value
	}

	// Load global templates - only add if not in local
	for _, mcpName := range mcpNames {
		templatePath := filepath.Join(templateRoot, mcpTemplatesDir, mcpTemplateFilename(mcpName))
		entry, missingEnv, err := loadMcpTemplate(templatePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("mcp template not found: %s", mcpName)
			}
			return nil, err
		}
		if len(missingEnv) > 0 {
			warnMissingEnv(templatePath, missingEnv)
		}
		if len(entry) == 0 {
			return nil, fmt.Errorf("mcp template empty: %s", templatePath)
		}

		// Merge entries - local takes precedence
		for key, value := range entry {
			if _, existsInLocal := mcpEntries[key]; existsInLocal {
				// Skip - local overrides global
				continue
			}
			mcpEntries[key] = value
		}
	}
	return mcpEntries, nil
}

func writeOpenCodeConfig(rootDir string, mcpEntries map[string]any) error {
	path := filepath.Join(rootDir, opencodeConfigFilename)
	payload, err := json.MarshalIndent(openCodeConfig{
		Mcp:    mcpEntries,
		Schema: "https://opencode.ai/config.json",
	}, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return writeFile(path, payload)
}

func writeMcpConfig(rootDir string, mcpEntries map[string]any) error {
	path := filepath.Join(rootDir, mcpConfigFilename)
	mcpServers := make(map[string]any, len(mcpEntries))
	for name, value := range mcpEntries {
		server, ok := value.(map[string]any)
		if !ok {
			mcpServers[name] = value
			continue
		}
		mcpServers[name] = normalizeMcpServer(server)
	}
	payload, err := json.MarshalIndent(mcpConfig{McpServers: mcpServers}, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return writeFile(path, payload)
}

func writeRootDocs(rootDir, templateRoot string, codexConfig, claudeConfig ResolvedToolConfig) error {
	codexContent, err := renderSnippets(templateRoot, snippetsDir, codexConfig.Snippets)
	if err != nil {
		return fmt.Errorf("codex snippets: %w", err)
	}
	claudeSnippets := claudeConfig.Snippets
	claudeContent, err := renderSnippets(templateRoot, snippetsDir, claudeSnippets)
	if err != nil {
		return fmt.Errorf("claude snippets: %w", err)
	}

	rootClaudePath := filepath.Join(rootDir, claudeFilename)
	if _, err := os.Stat(rootClaudePath); err == nil {
		fmt.Printf("Warning: %s already exists and will not be modified.\n", rootClaudePath)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	rootAgentsPath := filepath.Join(rootDir, agentsFilename)
	if _, err := os.Stat(rootAgentsPath); err == nil {
		fmt.Printf("Warning: %s already exists and will not be modified.\n", rootAgentsPath)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	agentsPath := filepath.Join(rootDir, ".codex", agentsFilename)
	if err := os.MkdirAll(filepath.Dir(agentsPath), defaultDirPermission); err != nil {
		return err
	}
	if err := writeFile(agentsPath, []byte(codexContent)); err != nil {
		return err
	}

	claudePath := filepath.Join(rootDir, ".claude", claudeFilename)
	if err := os.MkdirAll(filepath.Dir(claudePath), defaultDirPermission); err != nil {
		return err
	}
	if err := writeFile(claudePath, []byte(claudeContent)); err != nil {
		return err
	}
	return nil
}

func renderSnippets(templateRoot, dir string, snippetNames []string) (string, error) {
	if len(snippetNames) == 0 {
		return "", nil
	}
	var sections []string
	for _, name := range snippetNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		path := filepath.Join(templateRoot, dir, name+".md")
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		sections = append(sections, strings.TrimSpace(string(data)))
	}
	return strings.Join(sections, "\n\n") + "\n", nil
}

func copyDir(sourceDir, targetDir string) error {
	return filepath.WalkDir(sourceDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		targetPath := filepath.Join(targetDir, relative)
		if entry.IsDir() {
			return os.MkdirAll(targetPath, defaultDirPermission)
		}
		return copyFile(path, targetPath)
	})
}

func copyFile(sourcePath, targetPath string) error {
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	return writeFile(targetPath, data)
}

func writeFile(path string, data []byte) error {
	created := false
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			created = true
		} else {
			return err
		}
	}
	if err := os.WriteFile(path, data, defaultFilePermission); err != nil {
		return err
	}
	if created {
		logCreated(path)
	}
	return nil
}

func logCreated(path string) {
	if !verboseOutput {
		return
	}
	fmt.Printf("Created %s\n", path)
}

func sortedIDs(projectTypes map[string]ProjectType) []string {
	ids := make([]string, 0, len(projectTypes))
	for id := range projectTypes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
