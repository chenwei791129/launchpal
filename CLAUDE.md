<!-- SPECTRA:START v1.0.1 -->

# Spectra Instructions

This project uses Spectra for Spec-Driven Development(SDD). Specs live in `openspec/specs/`, change proposals in `openspec/changes/`.

## Use `/spectra:*` skills when:

- A discussion needs structure before coding → `/spectra:discuss`
- User wants to plan, propose, or design a change → `/spectra:propose`
- Tasks are ready to implement → `/spectra:apply`
- There's an in-progress change to continue → `/spectra:ingest`
- User asks about specs or how something works → `/spectra:ask`
- Implementation is done → `/spectra:archive`

## Workflow

discuss? → propose → apply ⇄ ingest → archive

- `discuss` is optional — skip if requirements are clear
- Requirements change mid-work? Plan mode → `ingest` → resume `apply`

## Parked Changes

Changes can be parked（暫存）— temporarily moved out of `openspec/changes/`. Parked changes won't appear in `spectra list` but can be found with `spectra list --parked`. To restore: `spectra unpark <name>`. The `/spectra:apply` and `/spectra:ingest` skills handle parked changes automatically.

<!-- SPECTRA:END -->

## Issue & Release Policy

- Do NOT close GitHub issues until the release PR workflow is complete and merged.
- When implementation is done, update the issue with a comment summarizing the changes, but leave it open.
- Issues are closed as part of the release process, not at implementation time.

## Post-Implementation Checklist

Every time a feature is added, modified, or removed, check whether the following files need updating:

- `README.md` — Features list, screenshots, known limitations
- `.claude/CLAUDE.md` — Directory structure, service capabilities, known limitations
- `CLAUDE.md` — This file (if workflow or tooling changes)
