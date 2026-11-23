# LaunchPal project guidance

See [AGENTS.md](AGENTS.md) for the domain glossary, architecture decisions, and engineering guide.

## Issue & Release Policy

- Do NOT close GitHub issues until the release PR workflow is complete and merged.
- When implementation is done, update the issue with a comment summarizing the changes, but leave it open.
- Issues are closed as part of the release process, not at implementation time.

## Post-Implementation Checklist

Every time a feature is added, modified, or removed, check whether the following files need updating:

- `README.md` — Features list, screenshots, known limitations
- `.claude/CLAUDE.md` — Directory structure, service capabilities, known limitations
- `CLAUDE.md` — This file (if workflow or tooling changes)
