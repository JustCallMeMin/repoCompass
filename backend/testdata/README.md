# Testdata Guide

This directory contains test-only assets for the RepoCompass backend.

## Purpose

Use `backend/testdata/` for static files that support automated tests without belonging to production code or runtime configuration.

Use `backend/testdata/fixtures/` for reusable fixture files such as:

- sample repository metadata
- sample configuration files
- expected output snapshots
- empty or minimal input cases

## Naming Conventions

- Prefer descriptive names such as `small-repo.json`, `sample-config.yaml`, or `empty-input.txt`.
- Keep names short, specific, and tied to the behavior under test.
- Use file extensions that match the real fixture format.

## Grouping Conventions

- Keep fixtures flat while the set is small.
- Introduce subdirectories only when grouping by behavior or domain makes the fixture set easier to navigate.
- Do not create deep fixture hierarchies without a clear need.

## Usage in Tests

- Reference fixtures through paths under `backend/testdata/fixtures/`.
- Keep fixtures reusable across tests when possible.
- Prefer minimal fixtures over large or noisy real-world dumps.

## What Not to Commit

- secrets or credentials
- machine-specific paths or local-only environment data
- generated noise that is not required for a test assertion
- oversized fixtures when a smaller focused example would work

## Current Status

- `backend/testdata/fixtures/` contains the first local repository fixture
- empty fixture subdirectories should keep `.gitkeep` only until they contain real fixture files
