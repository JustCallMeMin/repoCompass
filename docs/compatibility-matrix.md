# Compatibility Matrix

This document maps which Analyzers run against which Ecosystems, and which Fixtures are used to test them.

## Ecosystem Support

| Analyzer | Go | Node.js | Python | Description |
|----------|----|---------|--------|-------------|
| `readme` | ✅ | ✅ | ✅ | Checks if a root `README.md` exists. |
| `ci` | ✅ | ✅ | ✅ | Checks for GitHub Actions or other CI configs. |
| `contributing` | ✅ | ✅ | ✅ | Checks for `CONTRIBUTING.md`. |
| `scripts` | ✅ | ❌ | ❌ | Checks for `Makefile` or task runner scripts. (Node.js/Python support pending). |

## Fixture Coverage

| Fixture Path | Expected Outcome |
|--------------|------------------|
| `basic-go-repo/` | All checks pass. |
| `basic-nodejs-repo/` | All generic checks pass. |
| `basic-python-repo/` | All generic checks pass. |
| `good-onboarding-repo/` | Baseline repository with all best practices implemented. |
| `missing-ci-repo/` | Triggers `ci` analyzer findings. |
| `missing-contributing-repo/` | Triggers `contributing` analyzer findings. |
| `missing-readme-repo/` | Triggers `readme` analyzer findings. |
| `missing-scripts-repo/` | Triggers `scripts` analyzer findings. |
