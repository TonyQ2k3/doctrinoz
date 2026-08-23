---
icon: lucide/astroid
---

## Get an API Key

In order to use Doctrinoz, you need your own AI provider's API key.

+ [OpenRouter](https://openrouter.ai/settings/keys)

Once you've done setting up your key, you can set it with an environment variable
```shell
# Windows/Powershell
$Env:DOCTRINOZ_API_KEY="myapikey"

# Linux
export DOCTRINOZ_API_KEY="myapikey"
```

!!! note "Note"

    At the moment, Doctrinoz only supports OpenRouter.
    More model providers/aggregators will be added in future updates.

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