# Doctrinoz

Doctrinoz is a Go‑based CLI tool that generates and maintains an agent‑optimized wiki (Markdown) for a codebase.

Full Documentation: [tonyq2k3.github.io](https://tonyq2k3.github.io/doctrinoz/)

## Why it matters

The main problem that Doctrinoz solves is: **Regenerating codebase wiki can be unnecessarily expensive**, particularly when an AI model is involved. A small change to one component should not require regenerating the entire repository wiki.

Doctrinoz uses Git history and diffs to identify what has changed since the last commit, and only update the related pages. Fewer files are sent to the AI model per run means fewer tokens are billed — so cost scales with the size of your changes, not the size of your codebase.


## Installation & Usage

For a guide on installing and using Doctrinoz, visit the [Setup](https://tonyq2k3.github.io/doctrinoz/setup/) and [Usage](https://tonyq2k3.github.io/doctrinoz/usage/) page in the official documentation.

## Contribution

Doctrinoz is currently a single-maintainer project. While direct code contributions are not open at this time, community feedback is highly valued and plays an important role in shaping the project.

If you would like to help improve Doctrinoz, you can:

+ Report bugs

+ Suggest new features or enhancements

+ Ask questions or propose ideas

Please do so by opening an issue on the project’s GitHub repository. Clear descriptions, steps to reproduce, and screenshots (if applicable) are greatly appreciated.