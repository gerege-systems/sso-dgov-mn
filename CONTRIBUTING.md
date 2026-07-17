# Contributing to DAN-Government SSO

The full contribution guide lives in [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) —
local setup, branch/commit conventions, the `make pre-push` quality gate (lint +
race tests + swag drift + build), and the PR checklist.

Quick start:

```bash
# backend
cd backend && make pre-push      # lint + test -race + swag check + build

# frontend
cd frontend && npm ci && npm run lint && npm run build
```

By contributing you agree your work is licensed under the project's
[MIT License](LICENSE).
