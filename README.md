# Doctrinoz

Doctrinoz is a Go‑based CLI tool that generates and maintains a agent‑optimized wiki (Markdown) for a codebase.

# Building the CLI

```powershell
# Build the executable
go build

# For Linux
go build -o doctrinoz
```

# Install locally

```powershell
go install github.com/TonyQ2k3/doctrinoz@latest
```

# Usage

```powershell
# Set your API Key
$Env:DOCTRINOZ_API_KEY="myapikey"

# Initialize (Create YAML and state directory)
doctrinoz init path/to/your_repo

# Generate documentations
doctrinoz generate path/to/your_repo
```