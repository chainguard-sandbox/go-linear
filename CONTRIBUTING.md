# Contributing to go-linear

## Setup

```bash
git clone https://github.com/chainguard-sandbox/go-linear
cd go-linear
make dev    # installs tools + downloads deps
```

## Testing

| Tier | Command | API key | What it tests |
|------|---------|---------|---------------|
| Mock | `make test` | No | Unit tests, filters, parsing |
| Read | `make test-read` | Read | Live queries against Linear |
| Write | `make test-write` | Write | Creates/updates/deletes real data |

Mock tests run in CI. Read and write tests require `LINEAR_API_KEY`.

## Code Style

```bash
make check   # fmt + vet + lint + test + tidy
```

This is the same gate CI runs. Fix any issues before opening a PR.

## Pull Requests

1. Fork and branch from `main`
2. Keep changes focused — one concern per PR
3. Add tests for new functionality (mock tier at minimum)
4. Run `make check` before pushing
5. PR description should explain *why*, not just *what*

## Security

Report vulnerabilities privately to mark.esler@chainguard.dev. See [SECURITY.md](SECURITY.md).
