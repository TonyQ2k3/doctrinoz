package analyzer

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Module represents a logical unit of the codebase (directory or package).
type Module struct {
	Path         string   // relative path from repo root
	Name         string   // human name (last segment)
	Files        []string // files belonging to this module
	FileCount    int
	Language     string // dominant language
	IsEntryPoint bool   // contains main/index/app file
}

// AnalyzeRepo groups all source files into modules and returns them ordered
// by importance (entry points first, then by file count).
func AnalyzeRepo(repoPath string, ignorePatterns []string) ([]Module, error) {
	// Walk the directory tree.
	dirFiles := map[string][]string{}

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(repoPath, path)
		if shouldIgnore(rel, ignorePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() || !isSourceFile(path) {
			return nil
		}
		dir := filepath.Dir(rel)
		dirFiles[dir] = append(dirFiles[dir], rel)
		return nil
	})
	if err != nil {
		return nil, err
	}

	modules := make([]Module, 0, len(dirFiles))
	for dir, files := range dirFiles {
		name := filepath.Base(dir)
		if dir == "." {
			name = filepath.Base(repoPath)
		}
		lang := dominantLanguage(files)
		modules = append(modules, Module{
			Path:         dir,
			Name:         name,
			Files:        files,
			FileCount:    len(files),
			Language:     lang,
			IsEntryPoint: hasEntryPoint(files),
		})
	}

	sort.Slice(modules, func(i, j int) bool {
		if modules[i].IsEntryPoint != modules[j].IsEntryPoint {
			return modules[i].IsEntryPoint
		}
		return modules[i].FileCount > modules[j].FileCount
	})

	return modules, nil
}

// AffectedModules returns the subset of modules that contain any changed file.
func AffectedModules(modules []Module, changedFiles []string) []Module {
	changed := map[string]struct{}{}
	for _, f := range changedFiles {
		changed[f] = struct{}{}
		changed[filepath.Dir(f)] = struct{}{}
	}

	var affected []Module
	for _, m := range modules {
		for _, f := range m.Files {
			if _, ok := changed[f]; ok {
				affected = append(affected, m)
				break
			}
		}
	}
	return affected
}

// ReadModuleFiles returns the concatenated (truncated) content of a module's files.
// maxBytesPerFile prevents enormous files from blowing up context windows.
func ReadModuleFiles(repoPath string, module Module, maxBytesPerFile int) string {
	var sb strings.Builder
	for _, f := range module.Files {
		content, err := os.ReadFile(filepath.Join(repoPath, f))
		if err != nil {
			continue
		}
		sb.WriteString("\n\n### File: ")
		sb.WriteString(f)
		sb.WriteString("\n```")
		sb.WriteString(langFromExt(filepath.Ext(f)))
		sb.WriteString("\n")
		if len(content) > maxBytesPerFile {
			sb.Write(content[:maxBytesPerFile])
			sb.WriteString("\n// ... (truncated)\n")
		} else {
			sb.Write(content)
		}
		sb.WriteString("```\n")
	}
	return sb.String()
}

// ── helpers ───────────────────────────────────────────────────────────────────

var sourceExts = map[string]string{
	".go": "go", ".ts": "typescript", ".tsx": "typescript",
	".js": "javascript", ".jsx": "javascript",
	".py": "python", ".rs": "rust", ".java": "java",
	".kt": "kotlin", ".swift": "swift", ".cs": "csharp",
	".cpp": "cpp", ".c": "c", ".h": "c",
	".rb": "ruby", ".php": "php", ".scala": "scala",
	".ex": "elixir", ".exs": "elixir",
	".zig": "zig", ".lua": "lua",
	".sh": "bash", ".bash": "bash",
	".yaml": "yaml", ".yml": "yaml",
	".toml": "toml", ".json": "json",
	".md": "markdown", ".sql": "sql",
	".proto": "protobuf", ".graphql": "graphql",
	".tf": "terraform", ".hcl": "hcl",
	".dockerfile": "dockerfile",
}

func isSourceFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if base == "dockerfile" || base == "makefile" || base == "jenkinsfile" {
		return true
	}
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := sourceExts[ext]
	return ok
}

func langFromExt(ext string) string {
	if lang, ok := sourceExts[strings.ToLower(ext)]; ok {
		return lang
	}
	return ""
}

func dominantLanguage(files []string) string {
	counts := map[string]int{}
	for _, f := range files {
		if lang, ok := sourceExts[strings.ToLower(filepath.Ext(f))]; ok {
			counts[lang]++
		}
	}
	best, bestCount := "", 0
	for lang, count := range counts {
		if count > bestCount {
			best, bestCount = lang, count
		}
	}
	return best
}

var entryPointNames = map[string]bool{
	"main.go": true, "main.py": true, "main.rs": true,
	"main.ts": true, "main.js": true,
	"index.ts": true, "index.js": true, "index.py": true,
	"app.py": true, "app.ts": true, "app.js": true,
	"server.go": true, "server.ts": true, "server.js": true,
}

func hasEntryPoint(files []string) bool {
	for _, f := range files {
		if entryPointNames[strings.ToLower(filepath.Base(f))] {
			return true
		}
	}
	return false
}

func shouldIgnore(path string, patterns []string) bool {
	path = filepath.ToSlash(path)
	for _, pat := range patterns {
		pat = filepath.ToSlash(pat)
		if strings.HasSuffix(pat, "/") {
			if strings.HasPrefix(path, pat) || strings.Contains(path, "/"+pat) {
				return true
			}
		} else {
			matched, _ := filepath.Match(pat, filepath.Base(path))
			if matched {
				return true
			}
		}
	}
	return false
}
