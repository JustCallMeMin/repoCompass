# Local Setup Guide

This guide explains how to set up the RepoCompass development environment locally.

## Prerequisites

- **Go 1.22+**
- **Node.js 20+**
- **Git**
- **Docker** and **Docker Compose**
- `make` utility

## 1. Environment Configuration

Copy the example environment files for both the backend and frontend:

```bash
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env.local
```

## 2. Database Setup

RepoCompass uses PostgreSQL. You can spin up a local database instance using Docker:

```bash
# Start PostgreSQL in the background
make db-up

# Apply database migrations
make migrate-up

# (Optional) Seed the database with sample data
make db-seed
```

### Useful Database Commands
- `make migrate-status`: Check the current migration version.
- `make migrate-down`: Revert the last migration.
- `make db-reset`: Drop the database and re-apply all migrations.
- `make db-down`: Stop and remove the PostgreSQL container.

*(Note: `golang-migrate` is used under the hood in the `backend/scripts/dev` scripts.)*

## 3. Running the Backend

With the database running, you can start the API server:

```bash
make server
```

The server listens on `http://localhost:8080`.

## 4. Running the Frontend

In a separate terminal, install dependencies and start the Next.js development server:

```bash
make frontend-install
make frontend-dev
```

The dashboard will be available at `http://localhost:3000`.

## 5. Docker Runtime

If you want to run the entire stack (API, Frontend, Database) via Docker Compose without installing Go or Node.js locally:

```bash
make docker-build
make docker-up
```

To stop the stack:
```bash
make docker-down
```

## Common Friction Points

- **Missing `golang-migrate`**: The migration scripts (`make migrate-*`) rely on `golang-migrate`. If it is not installed globally, the scripts will attempt to use Docker. Ensure Docker is running.
- **Port Conflicts**: Ensure ports `8080` (API), `3000` (Frontend), and `5432` (PostgreSQL) are not occupied by other services.
