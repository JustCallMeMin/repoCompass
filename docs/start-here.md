# Start Here: Contributor Guide

Welcome to RepoCompass! This guide will help you navigate the project and make your first contribution.

## 1. What is RepoCompass?

RepoCompass is an engine that analyzes code repositories, evaluates them against rules, and generates onboarding and health reports. It is designed to be highly extensible.

## 2. Paths for Contributors

Depending on what you want to achieve, follow the appropriate path:

### "I want to run the project locally"
Start with Docker:

```bash
make docker-up
```

Then open `http://localhost:3000` and `http://localhost:8080/api/v1/health`.
For detailed setup or host-side commands, see the **[Local Setup Guide](local-setup.md)**.
On Windows, use Git Bash or WSL for Make targets that call shell scripts.

### "I want to understand how it works"
If you want to understand the inner workings of RepoCompass, start with the **[Architecture Walkthrough](architecture-walkthrough.md)**, and then check the **[Codebase Map](structure.md)**.

### "I want to fix a bug or add a core feature"
1. Follow the **[Local Setup Guide](local-setup.md)** to get your environment ready.
2. Check the **[Testing Guide](testing-guide.md)** to learn how to write and run tests.
3. Review the **[Contributor Checklist](contributor-checklist.md)** before opening a PR.
4. Use the root **[Contributing Guide](../CONTRIBUTING.md)** for PR and issue workflow.

### "I want to add support for a new language/framework"
You will need to write an Analyzer. See the **[Analyzer Extension Guide](analyzer-contract.md)**.

## 3. Good First Issues

If you're looking for something to work on, check out issues labeled `good first issue` on GitHub. These are scoped to be approachable for newcomers and usually involve:

- Adding a new fixture or test case
- Fixing a small bug in an existing analyzer
- Improving documentation

## 4. Getting Help

If you're stuck, feel free to open a Discussion on GitHub or ping the maintainers in the relevant issue.
