# proxy-runtime Refactor Plan

## Current Architecture

`proxy-runtime` is a Mihomo-only unified egress gateway.

```text
client -> fixed Mihomo mixed listener -> proxy username/password -> egress profile -> line -> exit
```

Control-plane facts are owned by `proxy-runtime` storage. Mihomo config is a generated runtime projection only; it is not an editable second source of truth.

## Runtime Contract

- Business clients consume proxy user routes, not upstream provider URLs.
- Dynamic IP provider accounts, endpoint selection, session acquisition, route apply/delete, restore, cleanup, and lease list semantics are owned by the lease application boundary.
- SQLite remains a supported standalone runtime path; PostgreSQL remains the primary multi-instance deployment store.
- PostgreSQL and SQLite store adapters share repository semantics for active, blocking, cleanup-pending, restorable, and session-bound leases.
- MetaCubeXD fork uses same-origin `/api/*` and `/mihomo/*` routes; no backend URL is user-configured in browser storage.
- Auth is cookie/session based for the UI; Mihomo controller secret injection is centralized in the backend reverse proxy.
- `/metrics` exposes Prometheus-compatible operation counters/durations for slow-path runtime visibility without secret-bearing labels.
- Logs, errors, metrics, and traces must not expose reusable tokens, cookies, Mihomo secret, provider credentials, proxy passwords, full proxy URLs, or dynamic session material.

## Deployment Validation

Deploy through the remote KVM/k3s environment only:

```bash
cd deploy
scripts/deploy-remote.sh proxy-runtime webui
```

Expected target:

- host: `pood1e@192.168.0.126`
- namespace: `byte-v-forge`
- Helm release: `byte-v-forge`
- kubeconfig: `/tmp/self-hosted-business-kubeconfigs/byte-v-forge.yaml`

## Remaining Plan

No active refactor backlog remains in this plan.
