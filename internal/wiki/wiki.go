package wiki

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/TonyQ2k3/doctrinoz/internal/analyzer"
	"github.com/TonyQ2k3/doctrinoz/internal/config"
	"github.com/TonyQ2k3/doctrinoz/internal/git"
	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

const (
	stateDir  = ".doctrinoz"
	stateFile = ".doctrinoz/state.json"
)

// ── Init ─────────────────────────────────────────────────────────────────────

func Init(repoPath string, cfg *config.Config) error {
	if !git.IsGitRepo(repoPath) {
		return fmt.Errorf("%s is not a git repository", repoPath)
	}

	stateFullPath := filepath.Join(repoPath, stateDir)
	if err := os.MkdirAll(stateFullPath, 0o755); err != nil {
		return fmt.Errorf("creating state dir: %w", err)
	}

	// Write .doctrinoz.yaml if absent.
	configPath := filepath.Join(repoPath, ".doctrinoz.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := fmt.Sprintf(`# doctrinoz configuration
provider: %s
model: %s
max_tokens: 4096
concurrent_pages: 3
ignore_patterns:
  - vendor/
  - node_modules/
  - .git/
  - dist/
  - build/
  - "*.generated.go"
  - "*.pb.go"
`, cfg.Provider, cfg.Model)
		if err := os.WriteFile(configPath, []byte(defaultConfig), 0o644); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
		color.Green("✓ Created .doctrinoz.yaml")
	}

	// Add .doctrinoz to .gitignore.
	gitignorePath := filepath.Join(repoPath, ".gitignore")
	appendToGitignore(gitignorePath, stateDir)

	color.Green("✓ doctrinoz initialised in %s", repoPath)
	color.Cyan("  Run 'doctrinoz generate' to create the initial wiki.")
	return nil
}

// ── Generate ──────────────────────────────────────────────────────────────────

func Generate(repoPath string, cfg *config.Config, force bool) error {
	modules, meta, gen, err := prepareGeneration(repoPath, cfg)
	if err != nil {
		return err
	}

	state := loadState(repoPath)
	headCommit, _ := git.HeadCommit(repoPath)

	bar := progressbar.NewOptions(len(modules)+1,
		progressbar.OptionSetDescription("Generating wiki..."),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer: "█", SaucerPadding: "░", BarStart: "[", BarEnd: "]",
		}),
	)

	ctx := context.Background()

	// Root AGENT.md
	rootContent, err := gen.GenerateRoot(ctx, meta, modules)
	if err != nil {
		return fmt.Errorf("generating root AGENT.md: %w", err)
	}
	if err := writeAgentMD(repoPath, ".", rootContent); err != nil {
		return err
	}
	state.PageCommits["."] = headCommit
	bar.Add(1)

	// Per-module AGENT.md files — respect concurrency limit.
	sem := make(chan struct{}, cfg.ConcurrentPages)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var genErr error

	for _, mod := range modules {
		if mod.Path == "." {
			bar.Add(1)
			continue
		}
		wg.Add(1)
		go func(m analyzer.Module) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			content, err := gen.GenerateModule(ctx, m)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				genErr = fmt.Errorf("module %s: %w", m.Path, err)
				bar.Add(1)
				return
			}
			if writeErr := writeAgentMD(repoPath, m.Path, content); writeErr != nil {
				genErr = writeErr
			} else {
				state.PageCommits[m.Path] = headCommit
			}
			bar.Add(1)
		}(mod)
	}

	wg.Wait()
	if genErr != nil {
		return genErr
	}

	state.LastCommit = headCommit
	state.GeneratedAt = time.Now()
	saveState(repoPath, state)

	fmt.Println()
	color.Green("✓ Wiki generated: %d AGENT.md files", len(modules)+1)
	return nil
}

// ── Update ────────────────────────────────────────────────────────────────────

func Update(repoPath string, cfg *config.Config) error {
	state := loadState(repoPath)
	if state.LastCommit == "" {
		color.Yellow("⚠ No previous state found — running full generation.")
		return Generate(repoPath, cfg, false)
	}

	changedFiles, err := git.ChangedFiles(repoPath, state.LastCommit)
	if err != nil {
		return fmt.Errorf("computing diff: %w", err)
	}
	if len(changedFiles) == 0 {
		color.Green("✓ No changes since last generation.")
		return nil
	}

	modules, _, gen, err := prepareGeneration(repoPath, cfg)
	if err != nil {
		return err
	}

	affected := analyzer.AffectedModules(modules, changedFiles)
	if len(affected) == 0 {
		color.Green("✓ Changed files don't affect any documented modules.")
		return nil
	}

	headCommit, _ := git.HeadCommit(repoPath)
	ctx := context.Background()
	bar := progressbar.NewOptions(len(affected),
		progressbar.OptionSetDescription(fmt.Sprintf("Updating %d modules...", len(affected))),
		progressbar.OptionShowCount(),
	)

	sem := make(chan struct{}, cfg.ConcurrentPages)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var updateErr error

	for _, mod := range affected {
		wg.Add(1)
		go func(m analyzer.Module) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			oldContent := readAgentMD(repoPath, m.Path)
			modChanged := filterChangedForModule(changedFiles, m)

			var content string
			var err error
			if oldContent == "" {
				content, err = gen.GenerateModule(ctx, m)
			} else {
				content, err = gen.UpdateModule(ctx, m, modChanged, oldContent)
			}

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				updateErr = fmt.Errorf("updating %s: %w", m.Path, err)
				bar.Add(1)
				return
			}
			if writeErr := writeAgentMD(repoPath, m.Path, content); writeErr != nil {
				updateErr = writeErr
			} else {
				state.PageCommits[m.Path] = headCommit
			}
			bar.Add(1)
		}(mod)
	}

	wg.Wait()
	if updateErr != nil {
		return updateErr
	}

	state.LastCommit = headCommit
	state.GeneratedAt = time.Now()
	saveState(repoPath, state)

	fmt.Println()
	color.Green("✓ Updated %d AGENT.md files", len(affected))
	return nil
}

// ── Status ────────────────────────────────────────────────────────────────────

func Status(repoPath string, cfg *config.Config) error {
	state := loadState(repoPath)

	headCommit, err := git.HeadCommit(repoPath)
	if err != nil {
		return err
	}

	fmt.Printf("HEAD commit  : %s\n", headCommit[:12])
	if state.LastCommit != "" {
		fmt.Printf("Last wiki run: %s (%s)\n",
			state.LastCommit[:12],
			state.GeneratedAt.Format("2006-01-02 15:04"))
	} else {
		color.Yellow("Wiki has not been generated yet — run 'doctrinoz generate'")
		return nil
	}

	if state.LastCommit == headCommit {
		color.Green("✓ Wiki is up to date")
		return nil
	}

	changedFiles, err := git.ChangedFiles(repoPath, state.LastCommit)
	if err != nil {
		return err
	}

	modules, _, _, err := prepareGeneration(repoPath, cfg)
	if err != nil {
		return err
	}

	affected := analyzer.AffectedModules(modules, changedFiles)
	if len(affected) == 0 {
		color.Green("✓ No documented modules affected by changes")
		return nil
	}

	color.Yellow("⚠ %d modules need regeneration:", len(affected))
	for _, m := range affected {
		fmt.Printf("  - %s/AGENT.md\n", m.Path)
	}
	color.Cyan("  Run 'doctrinoz update' to refresh.")
	return nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func prepareGeneration(repoPath string, cfg *config.Config) ([]analyzer.Module, *git.RepoMetadata, *Generator, error) {
	modules, err := analyzer.AnalyzeRepo(repoPath, cfg.IgnorePatterns)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("analyzing repo: %w", err)
	}

	meta, err := git.GatherMetadata(repoPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("reading git metadata: %w", err)
	}

	gen, err := NewGenerator(repoPath, cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	return modules, meta, gen, nil
}

func writeAgentMD(repoPath, modulePath, content string) error {
	dir := filepath.Join(repoPath, modulePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, "AGENT.md")
	return os.WriteFile(path, []byte(content), 0o644)
}

func readAgentMD(repoPath, modulePath string) string {
	path := filepath.Join(repoPath, modulePath, "AGENT.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

type state struct {
	LastCommit  string            `json:"last_commit"`
	GeneratedAt time.Time         `json:"generated_at"`
	PageCommits map[string]string `json:"page_commits"`
}

func loadState(repoPath string) *state {
	path := filepath.Join(repoPath, stateFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return &state{PageCommits: map[string]string{}}
	}
	var s state
	if err := json.Unmarshal(data, &s); err != nil {
		return &state{PageCommits: map[string]string{}}
	}
	if s.PageCommits == nil {
		s.PageCommits = map[string]string{}
	}
	return &s
}

func saveState(repoPath string, s *state) error {
	if err := os.MkdirAll(filepath.Join(repoPath, stateDir), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(repoPath, stateFile), data, 0o644)
}

func appendToGitignore(gitignorePath, entry string) {
	content, _ := os.ReadFile(gitignorePath)
	if len(content) > 0 && !contains(string(content), entry) {
		f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			defer f.Close()
			f.WriteString("\n# doctrinoz state\n" + entry + "\n")
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && (s[:len(substr)] == substr ||
			len(s) > len(substr) && containsAt(s, substr)))
}

func containsAt(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func filterChangedForModule(changedFiles []string, m analyzer.Module) []string {
	moduleSet := map[string]struct{}{}
	for _, f := range m.Files {
		moduleSet[f] = struct{}{}
	}
	var result []string
	for _, f := range changedFiles {
		if _, ok := moduleSet[f]; ok {
			result = append(result, f)
		}
	}
	return result
}
