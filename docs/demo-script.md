# Public Demo Readiness

This document outlines the resources and script for demonstrating RepoCompass in public (e.g., conferences, YouTube, team presentations).

## Target Public Repositories (T7-035)
For a realistic demo, we use the following public repositories as candidates:
1. **kubernetes/kubernetes** (Large, complex, shows scalability of the scan engine)
2. **expressjs/express** (Node.js ecosystem, shows language-agnostic capabilities)
3. **pallets/flask** (Python ecosystem)
4. **JustCallMeMin/repoCompass** (Self-scan to show "eating our own dog food")

## Demo Fixture Set (T7-034)
The default demo is offline-safe and uses the bundled fixture:

```bash
make demo
```

This scans `./backend/testdata/fixtures/local-repositories/good-onboarding-repo`
and does not require network access.

For public-repository demos, pre-fetch repositories before the presentation:

```bash
make demo-prepare
```

The helper script is `backend/scripts/dev/prepare-demo.sh`. It checks for `git`,
skips repositories that already exist, applies a clone timeout when the host has
`timeout`, and reports clear failures.

### Offline Fallback Fixture
If you are doing a live demo without internet access, or if the GitHub clone fails, use the built-in, offline-ready deterministic fixture:
- **Path**: `./backend/testdata/fixtures/local-repositories/good-onboarding-repo`
- **Output Expectation**: It will successfully scan and yield a deterministic passing score because all best practices are included.

## 5-10 Minute Demo Script (T7-036)

**0:00 - 1:00: Introduction**
- "Welcome to RepoCompass. Today I'm going to show you how we can instantly understand the health of any codebase."
- Explain the problem: Joining a new team or maintaining hundreds of repos is hard. You need a fast way to check if a repo follows best practices.

**1:00 - 3:00: The CLI Experience**
- "Let's start with the developer experience. This first run is offline-safe."
- Run `make demo`.
- Show the Markdown output in the terminal. Point out the score, findings, and recommendations.
- If public repositories were prepared with `make demo-prepare`, run `./backend/bin/repocompass scan /tmp/repocompass-demo/express`.

**3:00 - 5:00: The Self-Scan (Dogfooding)**
- "RepoCompass can scan itself."
- Run `./backend/bin/repocompass scan .` after `make build`.
- Compare self-scan output with the offline fixture output.

**5:00 - 7:00: The Web Dashboard**
- "CLI is great for CI, but managers and platform teams want a bird's eye view."
- Run `make docker-up`.
- Open `http://localhost:3000`.
- Show the history of scans, the organization overview, and how scores trend over time.

**7:00 - 9:00: Extensibility**
- "What if your company has custom rules?"
- Open `backend/internal/analyzers/example/analyzer.go`.
- Show the `Analyzer` interface in `backend/internal/analyzer/analyzer.go` and the runnable example analyzer.

**9:00 - 10:00: Conclusion**
- "RepoCompass is completely open source. Check out our GitHub, read the Start Here guide, and try it on your own repositories today."
