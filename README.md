# RepoCompass

RepoCompass is an open-source repository analysis and onboarding engine. It helps engineering teams instantly understand a new codebase, identify missing best practices, and assess the health of their repositories.

## Quickstart

Start the full stack (Dashboard + API + Database) via Docker:

```bash
git clone https://github.com/JustCallMeMin/repoCompass.git
cd repoCompass

# Build and start the stack
make docker-up
```

- **Dashboard**: [http://localhost:3000](http://localhost:3000)
- **API Health**: [http://localhost:8080/healthz](http://localhost:8080/healthz)

To run a scan using the CLI on a public repository:
```bash
make build
./backend/bin/repocompass scan https://github.com/kubernetes/kubernetes
```

To run a quick demo scan on RepoCompass itself:
```bash
make demo
```

## Documentation

Whether you want to learn how RepoCompass works or contribute to its development, our documentation covers everything:

- **[Start Here: Contributor Guide](docs/start-here.md)** - Entry point for new contributors.
- **[Local Setup](docs/local-setup.md)** - Detailed instructions for running components locally.
- **[Architecture Walkthrough](docs/architecture-walkthrough.md)** - High-level system design.
- **[Codebase Map](docs/structure.md)** - Directory structure reference.
- **[Testing Guide](docs/testing-guide.md)** - How to test your changes.

## Contributing

We welcome contributions! Please review the [Contributor Checklist](docs/contributor-checklist.md) before opening a Pull Request. Check out the issues labeled `good first issue` to get started.

## License

This project is licensed under the MIT License.
