# Asylum Wall App

Go server + embedded dashboard for the [Bancamp Asylum Wall](https://github.com/khannover/bancamp-asylum-wall) censorship log.

**This repo** (`asylum-wall-app`) is the application code. **Case data** lives separately in the public ledger repo [`bancamp-asylum-wall`](https://github.com/khannover/bancamp-asylum-wall).

## Quick Start

**From GHCR (no local build):**

```bash
cp .env.example .env
# Edit .env with your REPO_URL and GITHUB_TOKEN
docker pull ghcr.io/khannover/asylum-wall-app:latest
docker compose up -d
```

Set `image: ghcr.io/khannover/asylum-wall-app:latest` in `docker-compose.yml`, or build locally:

```bash
cp .env.example .env
# Edit .env with your REPO_URL and GITHUB_TOKEN
docker compose up -d --build
```

If a previous startup left the Git volume in a bad state, reset it once:

```bash
docker compose down -v && docker compose up -d --build
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `REPO_URL` | Yes | Public ledger repo, e.g. `https://github.com/khannover/bancamp-asylum-wall.git` |
| `GITHUB_TOKEN` | Yes (HTTPS) | PAT with `repo` scope for automated commit/push |
| `GIT_BRANCH` | No | Target branch (default: `main`) |
| `GIT_USER_NAME` | No | Git commit author name |
| `GIT_USER_EMAIL` | No | Git commit author email |
| `PORT` | No | Host port mapping (default: `8080`) |

### GitHub Token Setup

1. Use the public ledger repo (e.g. `bancamp-asylum-wall`).
2. Go to **Settings → Developer settings → Personal access tokens → Fine-grained tokens** (or classic).
3. Grant **Contents: Read and write** on the target repository.
4. Set `GITHUB_TOKEN` in `.env`.

### Publish Docker image to GHCR

Pushes run from CI (`.github/workflows/docker-publish.yml`) on each push to `main`, or locally:

```bash
# One-time: grant gh CLI package + workflow scopes
gh auth refresh -h github.com -s write:packages,workflow

# Push workflow (first time only, if not on GitHub yet)
git add .github/workflows/docker-publish.yml && git commit -m "Add GHCR publish workflow" && git push

# Or build and push manually
./scripts/publish-image.sh latest
```

Image: `ghcr.io/khannover/asylum-wall-app:latest`

After the first push, set the package to **public** under GitHub → Packages → asylum-wall-app → Package settings.

### SSH Deploy Key (alternative)

Mount a read-write deploy key and use an SSH `REPO_URL`:

```yaml
REPO_URL=git@github.com:your-org/asylum-wall-ledger.git
```

## Frontend

Open **http://localhost:8080** after starting Docker. The dashboard is embedded in the Go binary — no separate frontend server.

Features: case grid, search/filter/sort, detail modal with proof viewer, submit form with drag-and-drop upload.

## API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/` | Dashboard UI |
| `GET` | `/api/templates` | 10 standard case templates + signal counts |
| `GET` | `/api/cases` | All entries (newest first) |
| `GET` | `/api/proof/{filename}` | Serve proof file (sanitized) |
| `POST` | `/api/signal` | Quick “same issue here” signal (artist + platform + template) |
| `POST` | `/api/submit-case` | Full report for unusual cases (multipart) |
| `GET` | `/health` | Health check |