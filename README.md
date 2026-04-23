# RepoCompass

RepoCompass is an onboarding repository tool for new contributors.

This repository is still in the bootstrap phase. The current goal is to establish the initial project structure, project documentation, and local development conventions.

## Repository Layout

- `backend/`: Backend application code, database assets, scripts, and test fixtures.
- `frontend/`: Frontend application area.
- `deployments/`: Deployment manifests and environment-specific setup.
- `docs/`: Project documentation written in English.

For more details about the backend folder structure, see [docs/structure.md](./docs/structure.md).

## Local Setup

### Prerequisites

- Go 1.22+ installed
- Git installed
- A POSIX-compatible shell such as `zsh` or `bash`

### Clone the repository

```bash
git clone <your-repository-url>
cd repocompass
```

### Review the project structure

The project is still in its initial setup phase. Start by reviewing the current documentation:

- [docs/structure.md](./docs/structure.md): Initial repository and backend directory structure
- [docs/contributor-guide.md](./docs/contributor-guide.md): Contributor workflow and repository conventions

### Backend workspace

The backend workspace has already been prepared with the expected top-level folders:

- `backend/cmd`
- `backend/internal`
- `backend/db`
- `backend/testdata`
- `backend/scripts`

As backend modules and runnable services are introduced, this README should be updated with:

- dependency installation steps
- environment variable setup
- database startup instructions
- run, test, and migration commands

## Documentation

The repository currently uses local documentation under `docs/` as the repository-facing source of truth.

- [docs/structure.md](./docs/structure.md): Repository and backend structure reference
- [docs/contributor-guide.md](./docs/contributor-guide.md): Contributor workflow and repository conventions

## Current Status

- Initial repository structure is in place
- Backend foundational directories are present
- Project documentation has started in English
- Application code and executable setup are not yet added

## Contributing Notes

- Keep documentation in English
- Prefer following the existing backend layout before adding new top-level directories
- Update `README.md` and related docs when the setup process becomes more concrete
