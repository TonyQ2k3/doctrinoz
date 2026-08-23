---
icon: lucide/info
---

# About

Doctrinoz is a Go‑based CLI tool that generates and maintains a agent‑optimized wiki (Markdown) for a codebase.

Instead of generating human-friendly documents, Doctrinoz focuses on AI coding agents and Copilots that need reliable and structured knowledge about a codebase.

As the codebase evolves, Doctrinoz uses Git history and diffs to identify what has changed and updates only the pages affected by those changes, rather than repeatedly asking an AI agent to rediscover how a repository works.

## Why it matters

The main problem that Doctrinoz solves is: **Regenerating codebase wiki can be unnecessarily expensive**, particularly when an AI model is involved. A small change to one component should not require regenerating the entire repository wiki.

With Doctrinoz's Git awareness, fewer files are sent to the AI model per run, and fewer tokens are billed — so cost scales with the size of your changes, not the size of your codebase.

## How it works

Under the hood, Doctrinoz uses your own API key to call an AI model provider (e.g: OpenRouter, OpenAI or Anthropic) to actually write and refine the documentation. And because it's Git-aware, it use [Go-git](https://github.com/go-git/go-git) to compare which files have changed since the last commit, read those files, and only update those.

## Features

+ **Automatic wiki generation** — Point Doctrinoz at a repository path and it produces a full set of documentation pages, without you writing a word of it yourself.
+ **Agentic-AI-optimized output** — Markdown content is structured for consumption by LLMs and coding agents.
+ **Git-aware incremental updates** — Doctrinoz diffs against the latest commit, identifies which parts of the codebase changed, and regenerates only the affected pages instead of rebuilding the entire wiki.
+ **Single Go binary** — Simple installation, fast startup and easy integration with CI pipelines.

## Contribution

Doctrinoz is currently a single-maintainer project. While direct code contributions are not open at this time, community feedback is highly valued and plays an important role in shaping the project.

If you would like to help improve Doctrinoz, you can:

+ Report bugs

+ Suggest new features or enhancements

+ Ask questions or propose ideas

Please do so by opening an issue on the project’s GitHub repository. Clear descriptions, steps to reproduce, and screenshots (if applicable) are greatly appreciated.

## References

<a id="1">[1]</a> 
Jimenez, C. E., Yang, J., Wettig, A., Yao, S., Pei, K., Press, O., & Narasimhan, K. (2024).
*Can Language Models Resolve Real-World GitHub Issues*
[iclr.cc](https://proceedings.iclr.cc/paper_files/paper/2024/file/edac78c3e300629acfe6cbe9ca88fb84-Paper-Conference.pdf)

<a id="2">[2]</a> 
Ouyang, S. et al. (2024/2025).
*RepoGraph: Enhancing AI Software Engineering with Repository-Level Code Graph.*
[iclr.cc](https://proceedings.iclr.cc/paper_files/paper/2025/file/4a4a3c197deac042461c677219efd36c-Paper-Conference.pdf)

