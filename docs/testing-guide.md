# Testing Guide

RepoCompass relies on comprehensive testing to ensure stability. This guide explains how to run tests and write new ones.

## Running Tests

### Backend
The backend uses standard Go testing tools. We provide `make` targets to simplify execution.

1. **Unit Tests**:
   ```bash
   make test
   ```
   This runs all standard Go unit tests. It does not require a database.

2. **Integration Tests (PostgreSQL)**:
   ```bash
   make test-postgres
   ```
   This runs tests that require a running PostgreSQL instance. Ensure you have run `make db-up` and `make migrate-up` before running this.

3. **Linting and Formatting**:
   Before committing, ensure your code matches project standards:
   ```bash
   make fmt
   make vet
   ```

### Frontend
The frontend uses standard npm scripts.

```bash
cd frontend
npm run lint
npm run typecheck
npm test
npm audit --audit-level=moderate
npm run build
```

## Fixture Tests

Analyzers rely heavily on fixtures (sample repositories with specific conditions). Fixtures are located in `backend/testdata/fixtures/`.

To run fixture tests:
```bash
make test
```
The standard test runner automatically picks up fixture-based tests for analyzers.

### Adding a new fixture
When adding a new analyzer, create a minimal, deterministic fixture directory inside `backend/testdata/fixtures/`.
- Do not include real credentials.
- Keep the number of files as small as possible.
- Avoid committing large binaries or `node_modules`.

## Golden Tests

For report generation (e.g., Markdown or JSON output), we use Golden Tests. These tests compare the actual output against an expected "golden" file stored in the repository.

### Updating Golden Files
If you make a change to a Renderer that intentionally alters the output format, the golden tests will fail. You must update the golden files:

```bash
UPDATE_GOLDEN=true go test ./internal/report/...
```
*(Check the specific test implementation for the exact flag used to update golden files).*

Always review the `git diff` of the golden files carefully to ensure the new output is correct.

## Adding Tests for a New Analyzer

When you create a new Analyzer, you must provide:
1. A valid fixture that the Analyzer should pass.
2. An invalid fixture that triggers the Analyzer's findings.
3. Unit tests that invoke the Analyzer against both fixtures and assert the generated `Findings`.
