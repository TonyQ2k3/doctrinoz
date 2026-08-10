# Doctrinoz

Doctrinoz is a Go‑based CLI tool that generates and maintains a agent‑optimized wiki (Markdown) for a codebase.

# Building the CLI

```powershell
# Build the executable
go build

# For Linux
go build -o doctrinoz
```

# Usage

```powershell
# Set your API Key
$Env:DOCTRINOZ_API_KEY="myapikey"

# Initialize (Create YAML and state directory)
.\doctrinoz.exe init path/to/your_repo

# Generate documentations
.\doctrinoz.exe generate path/to/your_repo

```