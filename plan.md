# proxy-runtime Refactor Plan

## Target Architecture

`proxy-runtime` is a Mihomo-only unified egress gateway.

```text
client -> fixed Mihomo mixed listener -> proxy username/password -> egress profile -> line -> exit
```

Control-plane facts are owned by `proxy-runtime` storage. Mihomo config is a generated runtime projection only; it is not an editable second source of truth.

## Current Design Contract

- Business clients consume proxy user routes, not upstream provider URLs.
- Dynamic IP provider accounts, endpoint selection, session acquisition, route apply/delete, restore, cleanup, and lease list semantics are owned by the lease application boundary.
- SQLite remains a supported standalone runtime path; PostgreSQL remains the primary multi-instance deployment store.
- PostgreSQL and SQLite store adapters must share repository semantics for active, blocking, cleanup-pending, restorable, and session-bound leases.
- MetaCubeXD fork is the main frontend. It uses same-origin `/api/*` and `/mihomo/*` routes; no backend URL should be user-configured in browser storage.
- Auth is cookie/session based for the UI; Mihomo controller secret injection is centralized in the backend reverse proxy.
- Logs, errors, metrics, and traces must not expose reusable tokens, cookies, Mihomo secret, provider credentials, proxy passwords, full proxy URLs, or dynamic session material.

## Deployment-Ready Scope

The current deployment candidate is limited to:

- `proxy-runtime`
- `webui`

Deployment must be executed through `deploy/scripts/deploy-remote.sh` on the remote KVM/k3s environment. The local Mac is only the source editing and orchestration entrypoint.

## Validation Checklist

Before deployment:

- `proxy-runtime` git status is clean.
- Go source touched by the refactor is formatted with `gofmt`.
- `rtk git diff --check` passes.
- Focused stale-reference greps pass for removed coordinator inline factories, raw sensitive error logging, and old lease-list request patterns.

Remote deployment validation:

```bash
cd deploy
scripts/deploy-remote.sh proxy-runtime webui
```

Expected target:

- host: `pood1e@192.168.0.126`
- namespace: `byte-v-forge`
- Helm release: `byte-v-forge`
- kubeconfig: `/tmp/self-hosted-business-kubeconfigs/byte-v-forge.yaml`

## Post-Deploy Follow-up Backlog

These are non-blocking hardening items after the current deployment:

1. Add Prometheus-compatible slow-path metrics for lease list/acquire/release, worker runs/failures, provider requests, dataplane apply, and settings apply.
2. Continue moving Mihomo native projection/render/apply helpers into smaller package boundaries without changing generated config output.
3. Continue shrinking `runtimeSettingsStore` by separating pure normalization/validation from persistence and apply scheduling.
4. Keep HTTP handlers thin by moving any remaining transport-neutral logic into `internal/app/httpapi`, `internal/app/auth`, `internal/app/dashboard`, or usecase packages.
5. Keep auditing logs and client errors for sensitive provider/session material.
