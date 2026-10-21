---
name: moca-git-repos-backup
description: Check git-repos-backup changes with make lint, make test, gofmt and docker build; live-test Forgejo API compatibility with a real container.
auto: true
---

# git-repos-backup: verified checks

- **Before calling a git-repos-backup change done**: run `cd git-repos-backup && gofmt -l . && go vet ./... && go test ./...` plus `make lint`; a `grep` sweep asserts nothing (×1, 2026-10-21)
- **After editing Dockerfile, .dockerignore or backup_worker.sh**: run `cd git-repos-backup && docker build -t grb-test .`; .dockerignore filters COPY sources, hiding a `COPY --from=builder`: not found failure (×1, 2026-10-21)
- **Checking Forgejo/Gitea API compatibility**: do not infer from docs; start forgejo:11 with `FORGEJO__security__INSTALL_LOCK=true`, add the user via `forgejo admin user create --admin`, probe `/api/v1/repos/search` (×1, 2026-10-21)
- **When reviewing dependency freshness**: `cd git-repos-backup && go list -m -u all && go mod tidy -diff` (×1, 2026-10-21)
- **When cleaning scratch files here**: `rm -rf` is refused by the user's permission rule; use `rm -f` and report dirs left behind (×1, 2026-10-21)
