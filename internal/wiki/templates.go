package wiki

// SystemPrompt is the master instruction for every generation request.
const SystemPrompt = `You are an expert technical writer creating documentation for AI coding agents.
Your output will be read by Claude Code, Cursor, GitHub Copilot, and similar tools — not humans.

Rules:
- Be precise and dense. Agents prefer signal over narrative.
- Use Markdown with clear headings (##, ###) and code blocks.
- Every public function/type worth knowing should be mentioned.
- Always include: purpose, key abstractions, data flows, gotchas/non-obvious behaviour.
- Include concrete code snippets for key patterns.
- Never include fluff, disclaimers, or padding.`

// RootAgentPrompt generates the repo-level AGENT.md
const RootAgentPrompt = `Generate a comprehensive AGENT.md for the root of this repository.

Repository: {{.RepoName}}
Description: {{.Description}}
Languages detected: {{.Languages}}
Total source files: {{.TotalFiles}}
Recent contributors: {{.Authors}}

## Required sections (in this order):

### 1. Project Overview
- What this project does and why it exists
- Core domain concepts an agent must know

### 2. Repository Structure  
- Directory layout with one-line explanations
- Where to find: entry points, business logic, tests, config

### 3. Tech Stack & Dependencies
- Key frameworks/libraries and their roles
- Any version constraints that matter

### 4. Architecture & Data Flow
- How the major components fit together
- Request/event lifecycle (if applicable)
- Mermaid diagram if the flow is complex

### 5. Development Workflow
- How to build, test, and run
- Environment variables required
- Common dev commands

### 6. Key Patterns & Conventions
- Naming conventions, package layout rules
- Error handling patterns
- Testing approach

### 7. Agent-Specific Notes
- Files to NEVER edit (generated, vendor, etc.)
- Files that must always be updated together
- Gotchas and footguns

Source files:
{{.FileTree}}`

// ModuleAgentPrompt generates an AGENT.md for a subdirectory/package.
const ModuleAgentPrompt = `Generate a focused AGENT.md for this module/package.

Module: {{.ModuleName}} ({{.ModulePath}})
Language: {{.Language}}
Files in this module: {{.FileCount}}

## Required sections:

### 1. Module Purpose
One paragraph: what this module does and what problem it solves.

### 2. Public API
Every exported function, type, and constant that callers use:
- Signature (code block)
- What it does
- Parameters and return values
- Errors it can return
- Any side effects

### 3. Internal Design
- Key types and their relationships
- Important unexported helpers worth knowing

### 4. Dependencies
- Which other internal modules does this depend on?
- Which external packages?

### 5. Usage Examples
2-3 concrete code snippets showing idiomatic usage.

### 6. Gotchas
- Thread-safety, ordering requirements, performance cliffs, etc.

Module source:
{{.SourceCode}}`

// UpdatePrompt is used for incremental regeneration — it includes the old content.
const UpdatePrompt = `Update this existing AGENT.md based on the changed source files.

## Changed files in this module:
{{.ChangedFiles}}

## Previous AGENT.md content:
{{.OldContent}}

## Updated source code:
{{.SourceCode}}

Instructions:
- Preserve all sections that are still accurate.
- Update only what has changed.
- Add new public API entries for new functions/types.
- Remove entries for deleted functions/types.
- Mark breaking changes with ⚠️.
- Keep the same section structure unless the old structure is clearly wrong.`