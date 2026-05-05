# Contributor Checklist

Before you open a Pull Request (PR) to RepoCompass, please ensure you have completed the following checklist. This checklist will also be included in the PR template.

## The Checklist

- [ ] **Formatting**: I have run `make fmt` (Backend) or applied standard formatting (Frontend).
- [ ] **Linting**: I have run `make vet` (Backend) and `npm run lint` (Frontend) with no new warnings.
- [ ] **Testing**: I have added tests that prove my fix is effective or that my feature works.
- [ ] **Pass All Tests**: I have run `make test` and `make test-postgres` (if applicable) and all tests pass.
- [ ] **Golden Files**: If I changed report outputs, I have updated the golden tests and reviewed the diffs.
- [ ] **Documentation**: I have updated the relevant documentation (README, guides, API specs) if my changes affect user or contributor workflows.
- [ ] **Migrations**: If I altered the database schema, I have provided a valid `up` and `down` migration and tested it locally.
- [ ] **Clean Commits**: My commits are logically grouped, and I have used one commit per logical change (Lore commit style).

## Opening the PR
When you are ready, push your branch and open a PR against the `main` branch. A maintainer will review your code shortly.
