# Project Hub — Documentation

This `docs/` directory contains the design documents and mock assets for Project Hub TUI.

## Contents

- `design_doc.md` — Japanese design document and notes
- `design_doc_english.md` — English design document
- `moc/` — mock UI assets and notes
  - `github-projects-tui-mock.tsx` — React mock demonstrating intended UI
  - `README.md` — explanation and guidance for matching TUI to the mock

## Purpose

This folder collects authoritative docs for contributors: design decisions, UI mock, and guidance for implementation and visual parity with the mock.

## How to use

- Read `design_doc.md` (or `design_doc_english.md`) for architecture and UX rationale.
- Use `moc/github-projects-tui-mock.tsx` as the visual reference when implementing UI in the TUI code.
- Use `moc/README.md` for practical steps to align Bubbletea/Lipgloss styles to the mock.

## Contribution

When updating docs, prefer small, focused edits and add examples where helpful. Keep Japanese and English docs in sync.
