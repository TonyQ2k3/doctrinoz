package wiki

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/TonyQ2k3/doctrinoz/internal/analyzer"
	"github.com/TonyQ2k3/doctrinoz/internal/config"
	"github.com/TonyQ2k3/doctrinoz/internal/git"
	"github.com/TonyQ2k3/doctrinoz/internal/providers"
)

const maxBytesPerFile = 32_000 // ~8k tokens per file

// Generator orchestrates page creation.
type Generator struct {
	cfg      *config.Config
	provider providers.Provider
	repoPath string
}

func NewGenerator(repoPath string, cfg *config.Config) (*Generator, error) {
	p, err := providers.New(cfg)
	if err != nil {
		return nil, err
	}
	return &Generator{cfg: cfg, provider: p, repoPath: repoPath}, nil
}

// GenerateRoot creates the repo-level AGENT.md.
func (g *Generator) GenerateRoot(ctx context.Context, meta *git.RepoMetadata, modules []analyzer.Module) (string, error) {
	// Build file tree summary.
	var tree strings.Builder
	for _, m := range modules {
		tree.WriteString(fmt.Sprintf("- %s/ (%s, %d files)\n", m.Path, m.Language, m.FileCount))
		for _, f := range m.Files {
			if len(m.Files) <= 10 {
				tree.WriteString(fmt.Sprintf("  - %s\n", f))
			}
		}
	}

	// Summarise languages.
	langCounts := map[string]int{}
	for _, m := range modules {
		langCounts[m.Language] += m.FileCount
	}
	var langs []string
	for lang, count := range langCounts {
		if lang != "" {
			langs = append(langs, fmt.Sprintf("%s (%d files)", lang, count))
		}
	}

	data := map[string]any{
		"RepoName":    meta.Name,
		"Description": meta.Description,
		"Languages":   strings.Join(langs, ", "),
		"TotalFiles":  meta.TotalFiles,
		"Authors":     strings.Join(meta.RecentAuthors, ", "),
		"FileTree":    tree.String(),
	}

	prompt, err := renderTemplate(RootAgentPrompt, data)
	if err != nil {
		return "", err
	}

	return g.complete(ctx, prompt)
}

// GenerateModule creates an AGENT.md for a single module.
func (g *Generator) GenerateModule(ctx context.Context, module analyzer.Module) (string, error) {
	source := analyzer.ReadModuleFiles(g.repoPath, module, maxBytesPerFile)

	data := map[string]any{
		"ModuleName": module.Name,
		"ModulePath": module.Path,
		"Language":   module.Language,
		"FileCount":  module.FileCount,
		"SourceCode": source,
	}

	prompt, err := renderTemplate(ModuleAgentPrompt, data)
	if err != nil {
		return "", err
	}

	return g.complete(ctx, prompt)
}

// UpdateModule regenerates an AGENT.md incorporating only the changed files.
func (g *Generator) UpdateModule(ctx context.Context, module analyzer.Module, changedFiles []string, oldContent string) (string, error) {
	source := analyzer.ReadModuleFiles(g.repoPath, module, maxBytesPerFile)

	data := map[string]any{
		"ChangedFiles": strings.Join(changedFiles, "\n"),
		"OldContent":   oldContent,
		"SourceCode":   source,
	}

	prompt, err := renderTemplate(UpdatePrompt, data)
	if err != nil {
		return "", err
	}

	return g.complete(ctx, prompt)
}

// ── internals ─────────────────────────────────────────────────────────────────

func (g *Generator) complete(ctx context.Context, userPrompt string) (string, error) {
	req := providers.CompletionRequest{
		Messages: []providers.Message{
			{Role: "system", Content: SystemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:   g.cfg.MaxTokens,
		Temperature: 0.2, // low temperature = more deterministic docs
	}
	return g.provider.Complete(ctx, req)
}

func renderTemplate(tmpl string, data map[string]any) (string, error) {
	t, err := template.New("prompt").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
