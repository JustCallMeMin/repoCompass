# Public Demo Readiness

This document outlines the resources and script for demonstrating RepoCompass in public (e.g., conferences, YouTube, team presentations).

## Target Public Repositories (T7-035)
For a realistic demo, we use the following public repositories as candidates:
1. **kubernetes/kubernetes** (Large, complex, shows scalability of the scan engine)
2. **expressjs/express** (Node.js ecosystem, shows language-agnostic capabilities)
3. **pallets/flask** (Python ecosystem)
4. **JustCallMeMin/repoCompass** (Self-scan to show "eating our own dog food")

## Demo Fixture Set (T7-034)
Instead of relying on live network clones which can be slow during a live demo, we provide a script to pre-fetch demo repositories:

```bash
#!/usr/bin/env bash
# backend/scripts/dev/prepare-demo.sh
mkdir -p /tmp/repocompass-demo
git clone --depth 1 https://github.com/kubernetes/kubernetes /tmp/repocompass-demo/kubernetes
git clone --depth 1 https://github.com/expressjs/express /tmp/repocompass-demo/express
git clone --depth 1 https://github.com/pallets/flask /tmp/repocompass-demo/flask
echo "Demo fixtures prepared at /tmp/repocompass-demo"
```
*(You can create this script in `backend/scripts/dev/prepare-demo.sh` to quickly set up the demo environment).*

## 5-10 Minute Demo Script (T7-036)

**0:00 - 1:00: Introduction**
- "Welcome to RepoCompass. Today I'm going to show you how we can instantly understand the health of any codebase."
- Explain the problem: Joining a new team or maintaining hundreds of repos is hard. You need a fast way to check if a repo follows best practices.

**1:00 - 3:00: The CLI Experience**
- "Let's start with the developer experience. I've already cloned a few popular repositories."
- Run `repocompass scan /tmp/repocompass-demo/express`
- Show the Markdown output in the terminal. Point out the score, the findings, and the recommendations (e.g., missing CONTRIBUTING.md).
- Run `repocompass scan /tmp/repocompass-demo/flask` to show Python support.

**3:00 - 5:00: The Self-Scan (Dogfooding)**
- "RepoCompass can scan itself."
- Run `make demo`.
- Show that RepoCompass scores highly, but maybe point out a simulated warning if we removed a file temporarily.

**5:00 - 7:00: The Web Dashboard**
- "CLI is great for CI, but managers and platform teams want a bird's eye view."
- Run `make docker-up`.
- Open `http://localhost:3000`.
- Show the history of scans, the organization overview, and how scores trend over time.

**7:00 - 9:00: Extensibility**
- "What if your company has custom rules?"
- Open `backend/internal/analyzers/example/analyzer.go`.
- Show how simple the `Analyzer` interface is. Explain that anyone can write a custom analyzer in 50 lines of Go.

**9:00 - 10:00: Conclusion**
- "RepoCompass is completely open source. Check out our GitHub, read the Start Here guide, and try it on your own repositories today."
