package main

import (
	"bufio"
	"embed"
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

//go:embed templates/**
var templatesFS embed.FS

const (
	projectTypesGlob      = "templates/project_types/*.toml"
	agentsSnippetsDir     = "templates/agents_md_snippets"
	claudeSnippetsDir     = "templates/claude_md_snippets"
	skillTemplatesDir     = "templates/skill_templates"
	agentsFilename        = "AGENTS.md"
	claudeFilename        = "CLAUDE.md"
	projectConfigFilename = ".mavu_llm.toml"
	usageRulesConfigPath  = "lib/_mavubit/essentials/config/essentials_mix.exs"
	usageRulesFilename    = "USAGE_RULES.md"
	version               = "0.1.1"
	defaultFilePermission = 0o644
	defaultDirPermission  = 0o755
)

type ProjectConfig struct {
	Name        string        `toml:"name"`
	Description string        `toml:"description"`
	Agents      SnippetConfig `toml:"agents"`
	Claude      SnippetConfig `toml:"claude"`
	Skills      []string      `toml:"skills"`
}

type SnippetConfig struct {
	Snippets []string `toml:"snippets"`
}

type ProjectType struct {
	ID     string
	Config ProjectConfig
}

type ProjectTypeFile struct {
	Type string `toml:"type"`
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
	fmt.Println("  mavu-llm init --type <project-type> [--path <dir>]")
	fmt.Println("  mavu-llm update [--path <dir>]")
	fmt.Println("  mavu-llm version")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  types    List available project types")
	fmt.Println("  init     Create .codex/AGENTS.md, .claude/CLAUDE.md, and skills directories")
	fmt.Println("  update   Re-run setup using stored project type")
	fmt.Println("  version  Show current version")
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  If [claude] snippets are omitted, [agents] snippets are reused for CLAUDE.md")
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
		cfg := projectTypes[id]
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

func runInit(args []string) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	projectTypeFlag := flags.String("type", "", "Project type ID")
	pathFlag := flags.String("path", "", "Target directory (defaults to cwd)")
	if err := flags.Parse(args); err != nil {
		return err
	}

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

	return runSetup(rootDir, projectTypeID, projectTypes, "Initialized")
}

func runUpdate(args []string) error {
	flags := flag.NewFlagSet("update", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	pathFlag := flags.String("path", "", "Target directory (defaults to cwd)")
	if err := flags.Parse(args); err != nil {
		return err
	}

	rootDir := strings.TrimSpace(*pathFlag)
	if rootDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		rootDir = cwd
	}

	projectTypeID, err := readProjectTypeFile(rootDir)
	if err != nil {
		return err
	}

	projectTypes, err := loadProjectTypes()
	if err != nil {
		return err
	}

	return runSetup(rootDir, projectTypeID, projectTypes, "Updated")
}

func runSetup(rootDir, projectTypeID string, projectTypes map[string]ProjectConfig, action string) error {
	cfg, ok := projectTypes[projectTypeID]
	if !ok {
		return fmt.Errorf("unknown project type: %s", projectTypeID)
	}

	if err := createSkillDirs(rootDir, cfg); err != nil {
		return err
	}

	if err := writeRootDocs(rootDir, cfg); err != nil {
		return err
	}

	if err := writeProjectTypeFile(rootDir, projectTypeID); err != nil {
		return err
	}

	if err := runUsageRulesSync(rootDir); err != nil {
		return err
	}

	fmt.Printf("%s %s in %s\n", action, projectTypeID, rootDir)
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

	cmd := exec.Command("mix", "usage_rules.sync", usageRulesFilename, "--all", "--link-to-folder", "deps")
	cmd.Dir = rootDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message != "" {
			return fmt.Errorf("usage_rules.sync failed: %w\n%s", err, message)
		}
		return fmt.Errorf("usage_rules.sync failed: %w", err)
	}
	if len(output) > 0 {
		fmt.Print(string(output))
	}
	return appendUsageRules(rootDir)
}

func appendUsageRules(rootDir string) error {
	rulesPath := filepath.Join(rootDir, usageRulesFilename)
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

	return os.WriteFile(path, []byte(builder.String()), defaultFilePermission)
}

func loadProjectTypes() (map[string]ProjectConfig, error) {
	matches, err := fs.Glob(templatesFS, projectTypesGlob)
	if err != nil {
		return nil, err
	}

	projectTypes := make(map[string]ProjectConfig, len(matches))
	for _, match := range matches {
		data, err := templatesFS.ReadFile(match)
		if err != nil {
			return nil, err
		}
		var cfg ProjectConfig
		if err := toml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse %s: %w", match, err)
		}
		id := strings.TrimSuffix(filepath.Base(match), filepath.Ext(match))
		projectTypes[id] = cfg
	}
	return projectTypes, nil
}

func writeProjectTypeFile(rootDir, projectTypeID string) error {
	path := filepath.Join(rootDir, projectConfigFilename)
	payload, err := toml.Marshal(ProjectTypeFile{Type: projectTypeID})
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, defaultFilePermission)
}

func readProjectTypeFile(rootDir string) (string, error) {
	path := filepath.Join(rootDir, projectConfigFilename)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	var cfg ProjectTypeFile
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("parse %s: %w", path, err)
	}
	if strings.TrimSpace(cfg.Type) == "" {
		return "", fmt.Errorf("%s missing type", path)
	}
	return strings.TrimSpace(cfg.Type), nil
}

func promptProjectType(projectTypes map[string]ProjectConfig) (string, error) {
	ids := sortedIDs(projectTypes)
	if len(ids) == 0 {
		return "", errors.New("no project types available")
	}

	fmt.Println("Select a project type:")
	for i, id := range ids {
		cfg := projectTypes[id]
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

func createSkillDirs(rootDir string, cfg ProjectConfig) error {
	desiredSkills := make(map[string]struct{})
	for _, skill := range cfg.Skills {
		skill = strings.TrimSpace(skill)
		if skill == "" {
			continue
		}
		desiredSkills[skill] = struct{}{}
	}
	if len(desiredSkills) == 0 {
		return errors.New("project type has no skills configured")
	}

	skillTargets := []string{
		filepath.Join(rootDir, ".claude", "skills"),
		filepath.Join(rootDir, ".codex", "skills"),
	}
	for _, target := range skillTargets {
		if err := os.MkdirAll(target, defaultDirPermission); err != nil {
			return err
		}
		entries, err := os.ReadDir(target)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if _, ok := desiredSkills[entry.Name()]; ok {
				continue
			}
			if err := os.RemoveAll(filepath.Join(target, entry.Name())); err != nil {
				return err
			}
		}
	}

	for _, skill := range cfg.Skills {
		skill = strings.TrimSpace(skill)
		if skill == "" {
			continue
		}
		sourcePath := filepath.ToSlash(filepath.Join(skillTemplatesDir, skill))
		if _, err := fs.Stat(templatesFS, sourcePath); err != nil {
			return fmt.Errorf("skill template not found: %s", skill)
		}
		for _, targetRoot := range skillTargets {
			targetPath := filepath.Join(targetRoot, skill)
			if err := os.MkdirAll(targetPath, defaultDirPermission); err != nil {
				return err
			}
			if err := copyEmbeddedDir(sourcePath, targetPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeRootDocs(rootDir string, cfg ProjectConfig) error {
	agentsContent, err := renderSnippets(agentsSnippetsDir, cfg.Agents.Snippets)
	if err != nil {
		return fmt.Errorf("agents snippets: %w", err)
	}
	claudeSnippets := cfg.Claude.Snippets
	if len(claudeSnippets) == 0 {
		claudeSnippets = cfg.Agents.Snippets
	}
	claudeContent, err := renderSnippets(claudeSnippetsDir, claudeSnippets)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			claudeContent, err = renderSnippets(agentsSnippetsDir, claudeSnippets)
		}
		if err != nil {
			return fmt.Errorf("claude snippets: %w", err)
		}
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
	if err := os.WriteFile(agentsPath, []byte(agentsContent), defaultFilePermission); err != nil {
		return err
	}

	claudePath := filepath.Join(rootDir, ".claude", claudeFilename)
	if err := os.MkdirAll(filepath.Dir(claudePath), defaultDirPermission); err != nil {
		return err
	}
	if err := os.WriteFile(claudePath, []byte(claudeContent), defaultFilePermission); err != nil {
		return err
	}
	return nil
}

func renderSnippets(dir string, snippetNames []string) (string, error) {
	if len(snippetNames) == 0 {
		return "", nil
	}
	var sections []string
	for _, name := range snippetNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		path := filepath.ToSlash(filepath.Join(dir, name+".md"))
		data, err := templatesFS.ReadFile(path)
		if err != nil {
			return "", err
		}
		sections = append(sections, strings.TrimSpace(string(data)))
	}
	return strings.Join(sections, "\n\n") + "\n", nil
}

func copyEmbeddedDir(sourceDir, targetDir string) error {
	return fs.WalkDir(templatesFS, sourceDir, func(path string, entry fs.DirEntry, err error) error {
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
		targetPath := filepath.Join(targetDir, filepath.FromSlash(relative))
		if entry.IsDir() {
			return os.MkdirAll(targetPath, defaultDirPermission)
		}
		data, err := templatesFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, defaultFilePermission)
	})
}

func sortedIDs(projectTypes map[string]ProjectConfig) []string {
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
