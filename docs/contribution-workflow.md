# Contribution Workflow

This document outlines the standard workflow for contributing to RepoCompass.

## 1. Find or Create an Issue
Before writing code, ensure there is an open issue. This helps avoid duplicated effort and ensures your work aligns with the project goals.

## 2. Fork and Clone
1. Fork the repository on GitHub.
2. Clone your fork locally.
3. Add the upstream remote: `git remote add upstream https://github.com/JustCallMeMin/repoCompass.git`.

## 3. Branching Strategy
Create a new branch for your work:
```bash
git checkout -b feat/your-feature-name
```
or
```bash
git checkout -b fix/your-bug-fix
```

## 4. Development
- Write your code.
- Add tests.
- Ensure `make fmt`, `make vet`, and `make test` all pass.

## 5. Committing
We prefer clean, descriptive commits. If you have many small work-in-progress commits, please squash them before opening the PR.

## 6. Pull Request
Open a PR against the `main` branch. Fill out the PR template completely.
