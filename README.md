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
git clone https://github.com/JustCallMeMin/repoCompass.git
cd repocompass
```

### Review the project structure

The project is still in its initial setup phase. Start by reviewing the current documentation:

- [docs/structure.md](./docs/structure.md): Initial repository and backend directory structure
- [docs/contributor-guide.md](./docs/contributor-guide.md): Contributor workflow and repository conventions
- [backend/db/README.md](./backend/db/README.md): Database directory usage guide
- [backend/.env.example](./backend/.env.example): Local backend and database environment template

### Backend workspace

The backend workspace has already been prepared with the expected top-level folders:

- `backend/cmd`
- `backend/internal`
- `backend/db`
- `backend/testdata`
- `backend/scripts`

The backend now includes a minimal Go module and CLI entrypoint:

```bash
cd backend
go run ./cmd/repocompass --help
go run ./cmd/repocompass scan
```

The backend also includes:

- a Cobra-based CLI skeleton at `backend/cmd/repocompass`
- internal package scaffolding for core modules under `backend/internal`
- documented database layout under `backend/db`
- script-based migration workflow under `backend/scripts/dev`
- standard local `fmt`, `vet`, and `test` scripts under `backend/scripts/dev`
- Docker Compose-based local PostgreSQL bootstrap

As backend modules and runnable services are introduced, this README should be updated with:

- dependency installation steps
- database startup instructions
- higher-level workflow commands if new wrappers are introduced

### Backend local workflow

Configure local backend and database values:

```bash
cp backend/.env.example backend/.env
```

Then use the standard backend scripts:

```bash
./backend/scripts/dev/db-up.sh
./backend/scripts/dev/fmt.sh
./backend/scripts/dev/vet.sh
./backend/scripts/dev/test.sh
./backend/scripts/dev/migrate-status.sh
```

## Documentation

The repository currently uses local documentation under `docs/` as the repository-facing source of truth.

- [docs/structure.md](./docs/structure.md): Repository and backend structure reference
- [docs/contributor-guide.md](./docs/contributor-guide.md): Contributor workflow and repository conventions
- [backend/db/README.md](./backend/db/README.md): Database migration and seed directory guide
- [backend/.env.example](./backend/.env.example): Local environment variable template

## Current Status

- Initial repository structure is in place
- Backend Go module is in place
- Cobra-based CLI skeleton is in place
- Internal package scaffolding is in place
- Database migration and seed organization is documented
- Script-based migration workflow is in place
- Standard local `fmt`, `vet`, and `test` scripts are in place
- Docker Compose-based local PostgreSQL bootstrap is in place
- Project documentation has started in English
- CI workflow is not implemented yet
- Real business logic is not implemented yet

## Contributing Notes

- Keep documentation in English
- Prefer following the existing backend layout before adding new top-level directories
- Update `README.md` and related docs when the setup process becomes more concrete
