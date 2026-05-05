# Architecture Walkthrough

This document provides a high-level overview of how RepoCompass works, helping new contributors understand the flow of data and where to make changes.

## Core Concepts

RepoCompass revolves around scanning a codebase and producing structured reports. The workflow follows a pipeline:

1. **Repository**: The source code to be analyzed.
2. **Snapshot**: A point-in-time capture of the repository state (files, commits).
3. **Scan**: The process of running rules against the Snapshot.
4. **Findings**: The individual issues or observations detected during the Scan.
5. **Report**: The aggregated Findings formatted for user consumption (e.g., Markdown, JSON).
6. **Dashboard**: The UI to visualize Reports and metrics.

## Component Map

### 1. Providers (`internal/integration`)
Providers fetch the code. A `RepositoryProvider` knows how to download or access the repository (e.g., GitHub, Local Filesystem).

### 2. Snapshots (`internal/snapshot`)
Snapshots abstract the file system. They provide a unified interface for Analyzers to read files, regardless of where the repository came from.

### 3. Analyzers & Rules (`internal/rules`, `internal/assessment`)
Analyzers inspect the Snapshot. Each ecosystem (Go, Node.js, Python, etc.) or concept (CI, Documentation) has an Analyzer. Analyzers apply Rules and generate Findings.

### 4. Scan Engine (`internal/scan`)
The Scan Engine coordinates the process. It takes a Snapshot, runs the registered Analyzers in parallel, and collects the Findings.

### 5. Renderers (`internal/report`)
Renderers format the Findings into a final Report (e.g., Markdown, JSON, HTML).

### 6. Persistence & History (`internal/history`, `internal/repository`)
Scan results are saved to PostgreSQL. The History read models serve this data to the API.

### 7. Product Surface (`frontend`, `internal/api`)
The HTTP API exposes the data, and the Next.js Dashboard consumes it.

## Where to Change What?

- **Adding a new language check?** -> Add a new Analyzer in `internal/rules/`.
- **Changing the API response?** -> Update `internal/api/` and `internal/history/`.
- **Modifying the Database Schema?** -> Add a new migration in `backend/db/migrations/` and update `internal/repository/`.
- **Changing the UI?** -> Edit the Next.js app in `frontend/`.
