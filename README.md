# RepoCompass

RepoCompass is an open-source repository analysis and onboarding engine. It helps engineering teams instantly understand a new codebase, identify missing best practices, and assess the health of their repositories.

## Quickstart

Start the full stack through Docker first. This path starts PostgreSQL, the API,
and the dashboard with the least host setup:

```bash
git clone https://github.com/JustCallMeMin/repoCompass.git
cd repoCompass
make docker-up
```

- **Dashboard**: [http://localhost:3000](http://localhost:3000)
- **API Health**: [http://localhost:8080/api/v1/health](http://localhost:8080/api/v1/health)

To run an offline deterministic CLI demo:

```bash
make demo
```

To prepare optional public demo repositories:

```bash
make demo-prepare
```

To scan a public repository directly:

```bash
make build
./backend/bin/repocompass scan https://github.com/octocat/Hello-World
```

On Windows, run Make targets from Git Bash or WSL. Without POSIX shell tools,
use Docker Compose commands directly.

## Documentation

Whether you want to learn how RepoCompass works or contribute to its development, our documentation covers everything:

- **[Start Here: Contributor Guide](docs/start-here.md)** - Entry point for new contributors.
- **[Contributing](CONTRIBUTING.md)** - Root contributor workflow and PR checklist links.
- **[Local Setup](docs/local-setup.md)** - Detailed instructions for running components locally.
- **[Architecture Walkthrough](docs/architecture-walkthrough.md)** - High-level system design.
- **[Codebase Map](docs/structure.md)** - Directory structure reference.
- **[Testing Guide](docs/testing-guide.md)** - How to test your changes.
- **[Security Policy](SECURITY.md)** - How to report vulnerabilities privately.
- **[Code of Conduct](CODE_OF_CONDUCT.md)** - Community behavior rules.

## Contributing

We welcome contributions. Start with [CONTRIBUTING.md](CONTRIBUTING.md), then
review the [Contributor Checklist](docs/contributor-checklist.md) before opening
a pull request. Check issues labeled `good first issue` to get started.

## License

This project is licensed under the MIT License.
