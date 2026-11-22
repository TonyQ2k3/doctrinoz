---
icon: lucide/astroid
---

## Get an API Key

In order to use Doctrinoz, you need your own AI provider's API key.

+ [OpenRouter](https://openrouter.ai/settings/keys)
+ [OpenAI](https://platform.openai.com/api-keys)

Select a provider for generation:

```shell
doctrinoz config set --provider openrouter
```

For OpenRouter, set `OPENROUTER_API_KEY`. For OpenAI, set `OPENAI_API_KEY`:

```shell
# Windows/Powershell
$Env:OPENAI_API_KEY="my-openai-api-key"
$Env:OPENROUTER_API_KEY="my-openrouter-api-key"

# Linux
export OPENAI_API_KEY="my-openai-api-key"
export OPENROUTER_API_KEY="my-openrouter-api-key"
```

OpenRouter defaults to the `openrouter/free` model. Set a different model with:

```shell
doctrinoz config set --model <model-name>
```

## Initialization

Run the initialize command to prepare for generation.

`doctrinoz init` will create a config file (.doc.yaml) and a state folder, then add those files into your .gitignore so they're properly ignored by VCS.
```shell
doctrinoz init path/to/your_repo
```

## Generate documentation

Generate your documentation (run once, this step could take a few minutes)
```shell
doctrinoz generate [path/to/your_repo]
```

In case you want to force re-generate your doc, use the `--force` option.

## Update your documentation

To incrementally update your doc, run the `update` command:
```shell
doctrinoz update [path/to/your_repo]
```

To check what's stale:
```shell
doctrinoz status [path/to/your_repo]
```

<!-- Integrate with CI/CD pipelines -->