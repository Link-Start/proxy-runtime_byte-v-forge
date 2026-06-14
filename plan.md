# proxy-runtime Mihomo-only Target Architecture

## Summary

`proxy-runtime` adopts the mature dynamic proxy provider shape: fixed entry plus proxy username/password parameters controlling egress.

Mihomo is the only data plane. `proxy-runtime` is the control plane for dynamic IP provider adapters, provider accounts/endpoints, proxy user routes, dynamic IP leases, Egress Profiles, Mihomo config rendering, and runtime observations.

MetaCubeXD becomes the main `proxy-runtime` frontend through a project-owned fork. The fork keeps upstream Mihomo operations as-is. Mihomo-native fixed proxies, subscriptions, proxy providers, rules, groups, and config editing stay in upstream pages. The project overlay adds `入口用户` inside `proxies` for proxy username/password plus line/exit bindings, `动态IP提供商` inside `proxies` for proxy-runtime-only dynamic provider endpoints and provider accounts, and `动态租约` inside `connections` for active dynamic lease runtime state.

There is one desired-state model: service-owned control-plane facts in the database. The generated Mihomo config is only a runtime projection of those facts.

## Architecture

```text
Internal app
  -> fixed entry host:port
  -> proxy username/password
  -> Mihomo IN-USER rule
  -> egress profile
  -> line layer
  -> direct, static IP, or dynamic IP exit
```

No GOST runtime is retained. Chain-style egress is represented only as two-layer Egress Profile configuration and rendered into Mihomo-native groups and dynamic IP nodes with `dialer-proxy` when the exit is proxy-runtime-owned dynamic IP. Mihomo-native proxies and proxy providers are referenced by name and are not cloned by proxy-runtime.

## Key Decisions

- Business clients consume proxy user routes, not upstream provider URLs.
- The fixed entry uses one Mihomo mixed listener.
- Proxy usernames map to registered egress policies.
- Dynamic IP leases materialize as ordinary Mihomo HTTP/SOCKS proxy nodes.
- Egress Profiles materialize as a line group and an exit group.
- Provider credentials and dynamic session parameters stay inside the control plane and generated Mihomo config.
- Mihomo-native fixed proxies, subscriptions, proxy providers, rules, groups, and node health are managed by Mihomo and MetaCubeXD.
- A forked MetaCubeXD is the primary UI. It is served full-page, not embedded as an iframe and not duplicated as a Byte-V dashboard tab.
- The existing Byte-V dashboard module is removed. The proxy-runtime frontend is served as a standalone app on its own host root, for example `proxy-runtime.<byte-v-forge-host>/`, and redirects to the forked MetaCubeXD frontend.
- The added project tabs edit only proxy-runtime-owned facts through same-origin HTTP APIs. The backend renders profiles, ingress rules, dynamic provider endpoints, accounts, and leases into Mihomo config.
- Mihomo config changes hot-reload through the external-controller whenever possible. Restart is only a fallback when the process is not running or the listener endpoint changes.

## Clash Verge Rev Lessons

Clash Verge Rev is a desktop client, so its Tauri process model, system proxy controls, TUN service mode, local file UX, and JavaScript script engine are not copied.

The useful pattern is its config lifecycle:

```text
profile/runtime input
  -> optional business overlay
  -> generated runtime Mihomo YAML/JSON
  -> validate/apply
  -> hot reload core
```

Borrowed decisions:

- Treat proxy-runtime-owned input as control-plane data, not as live UI state inside MetaCubeXD.
- Generate a complete desired Mihomo config from stored facts instead of patching fragments ad hoc.
- Keep only the last accepted generated Mihomo runtime config bytes as a rollback buffer.
- Validate before committing a profile switch where practical.
- Preserve the active runtime when a profile/provider update fails.

Rejected decisions:

- No Fluxo dependency.
- No desktop profile UX, WebDAV sync, JS script enhancement, or local file editor.
- No separate custom proxy-runtime dashboard that duplicates the forked MetaCubeXD UI.
- No hand-editing of the generated Mihomo runtime projection as a second source of truth. The forked UI edits `proxy-runtime` configuration, and reconcile renders the Mihomo config.

## Implementation Scope

- Keep a single Mihomo data plane and remove standalone source/route control surfaces.
- Remove GOST driver, GOST config generation, and GOST process management from the main path.
- Render base provider pool, proxy user routes, dynamic provider endpoints, and dynamic leases into one Mihomo config.
- Render egress profiles into hidden Mihomo line/exit groups. Reference Mihomo-native proxies/providers by name; clone only proxy-runtime-owned dynamic IP materialized nodes when a selected route requires `dialer-proxy`.
- Use `IN-USER` rules for proxy user routing.
- Vendor or build a project-owned MetaCubeXD fork and serve it as the proxy-runtime main frontend.
- Keep upstream MetaCubeXD runtime pages intact: overview, proxies, rules, connections, logs, config, and provider update operations.
- Add the `动态IP提供商` tab inside the forked MetaCubeXD `proxies` page for dynamic provider endpoints and provider accounts.
- Add Egress Profile and ingress-rule management inside the forked MetaCubeXD `rules` page.
- Add the `动态租约` tab inside the forked MetaCubeXD `connections` page for active lease runtime state.
- Remove the old standalone proxy-runtime dashboard business pages and module-federation entry.
- Apply profile, dynamic provider, proxy user, and lease changes by rendering a full desired config and hot-reloading Mihomo via `PUT /configs?force=true`.
- On hot reload failure, restore only the last accepted generated Mihomo runtime config bytes. This is a rollback buffer, not a second config model or compatibility track.
- Keep only the current proxy-runtime-owned HTTP APIs; Mihomo-native node/provider state is read from Mihomo.

## Hot Reload Contract

- Render config to a temp file and atomically replace the runtime config path.
- Hot reload with Mihomo external-controller `PUT /configs?force=true` using the absolute config path.
- Wait for the entry listener and controller observations after reload.
- Record the applied runtime revision only after Mihomo accepts the new projection.
- Roll back to the last accepted generated runtime config bytes on reload failure.
- If a persisted desired-state change cannot be applied immediately, mark reconcile failed and retry from the same desired state. Do not create a second config model or compatibility path.
- Restart Mihomo only when it is not running, the fixed entry endpoint changes, or hot reload cannot recover.
- Never log provider credentials, subscription URLs with tokens, proxy passwords, or full upstream proxy URLs during reload errors.

## Verification

- Format Go source with `gofmt`.
- Search for stale GOST, proxy-source, gateway/pool, and resolver references.
- Run static checks in the allowed environment.
- Validate runtime scenarios: empty entry, proxy user routing, Mihomo-native proxy/provider observation, dynamic lease acquire/release, and restore.

## Runtime Reliability and Clean Code Refactor Plan

### Target Outcome

`proxy-runtime` should move from a large `Runtime` object plus synchronous slow paths to a usecase-oriented runtime:

```text
transport -> usecase -> ports -> adapters
```

The refactor must preserve one source of truth:

- Durable control-plane facts live in the service-owned database.
- Hot/ephemeral state, locks, leases, and idempotency windows use TTL-backed runtime state where needed.
- Mihomo config remains a generated runtime projection, never a second editable model.

Success criteria:

- Login and dashboard loading are not blocked by lease history scans.
- Playground lease acquire does not wait for inactive lease history.
- Service startup is not blocked by provider calls or lease restore.
- HTTP handlers only adapt requests/responses; they do not run business orchestration directly.
- Usecases do not depend on `*Runtime`, `gin.Context`, or concrete store implementations.
- Store queries do not load all JSON rows and filter business state in Go.
- External HTTP/gRPC/SDK calls have explicit timeout, bounded retry/backoff, and visible failure handling.
- Logs, metrics, traces, and errors never expose provider passwords, session material, API tokens, cookies, Mihomo secret, or full proxy URLs.

### Phase 0: Migration Boundary

Scope:

- Touch only `proxy-runtime` unless deployment values or scripts are required.
- Keep cross-repository collaboration through proto, HTTP, events, or deploy config.
- Do not add deprecated wrappers, compatibility aliases, temporary adapters, or dual business flows.
- Do not add tests in this project phase; validate with formatting, static checks, focused grep, generated-code checks, and remote/deploy validation where artifacts are required.

Required before each implementation batch:

- Check `proxy-runtime` git status.
- Read `proxy-runtime/AGENTS.md`.
- Keep each migration batch independently committable.

### Phase 1: Fix User-Visible Latency

#### Lease list query contract

Current defect:

- Main UI calls `GET /api/leases?include_inactive=true`.
- Backend reads full `lease_json` history and unmarshals rows before filtering.
- Large history makes Playground and dashboard appear stuck.

Target API shape:

```text
GET /api/leases?status=active&limit=50
GET /api/leases?status=recent&limit=50
GET /api/leases?status=history&cursor=<cursor>&limit=50
GET /api/leases/{lease_id}
```

Rules:

- `active` means `status = ACTIVE` and `expires_at > now`.
- `recent` is bounded by `limit`.
- `history` is cursor-paginated.
- Full JSON/proto detail is fetched only by lease ID.
- Main UI must not default to inactive history.

Store target methods:

```text
ListActiveLeases(ctx, filter)
ListRecentLeases(ctx, page)
ListLeaseHistory(ctx, page)
GetLeaseFact(ctx, leaseID)
HasBlockingLease(ctx, providerAccount)
FindActiveLeaseBySession(ctx, providerAccount, sessionID)
```

Data access requirements:

- Use SQL predicates for active, blocking, session, provider account, and expiry checks.
- Replace full-row JSON scans with projected columns.
- Add or verify indexes for:
  - `(status, expires_at)`
  - `(provider_key, account_key, status, expires_at)`
  - `(acquired_at DESC, updated_at DESC)`
  - `session_id` if session lookup is required.

Acceptance:

- Active lease list is bounded and does not scan inactive history.
- `HasBlockingLease` uses an existence query, not full row loading.
- Playground opens without waiting on full history.

#### Playground acquire path

Current defect:

```text
save playground rule -> leases.load() -> acquire lease -> leases.load()
```

Target flow:

```text
save playground rule without lease refresh
acquire lease
refresh active lease only
```

Rules:

- Save failure blocks acquire.
- Acquire failure shows the provider/runtime error.
- Refresh failure does not hide a successful acquire.
- Active lease check must include expiry, not only status.

Acceptance:

- Clicking acquire enters acquire loading immediately.
- No inactive history request is made before acquire.
- Expired active rows do not block UI state.

#### Frontend request policy

Target:

- Centralize proxy-runtime API requests.
- Add timeout and cancellation.
- Normalize errors into:
  - `timeout`
  - `unauthorized`
  - `backend_unreachable`
  - `validation_error`
  - `provider_error`
  - `internal_error`

Default timeouts:

- Lease refresh: short bounded timeout.
- Normal API calls: bounded default timeout.
- Long apply/reconcile operations: moved to background operation instead of relying on long HTTP requests.

Acceptance:

- Backend slowness is displayed as a clear timeout or unavailable state.
- Component unmount cancels pending requests.
- Repeated refresh does not create uncontrolled concurrent requests.

### Phase 2: Decouple Startup and Background Work

Current defect:

```text
refresh(ctx)
restoreActiveLeases(ctx)
serveHTTP()
```

Provider or lease restore latency can delay HTTP startup.

Target lifecycle:

```text
load minimal config
initialize stores and dataplane controller
start HTTP
start workers
restore leases in background
publish runtime status
```

Health model:

```text
/healthz              process liveness
/readyz               dataplane and minimal runtime readiness
/api/runtime/status   UI-facing restore/apply/worker status
```

Background lease worker responsibilities:

- Restore active leases.
- Cleanup expired leases.
- Cleanup failed cleanup-pending leases.
- Release orphan dynamic slots.
- Reconcile provider session state where required.

Worker requirements:

- Bounded concurrency.
- Per-task timeout.
- Bounded retry with backoff.
- Idempotent state transitions.
- Structured logs with request or operation correlation.
- No secret or full proxy URL output.

Acceptance:

- HTTP starts even when provider restore is slow or failing.
- UI sees `restoring` or `degraded` state instead of backend unreachable.
- Lease cleanup no longer depends on user-triggered requests.

### Phase 3: Extract Lease Application

Target package:

```text
internal/app/lease/
  application.go
  repository.go
  worker.go
  active_predicate.go
  slot_allocator.go
  errors.go
```

Target application dependencies:

```text
LeaseRepository
ProviderSessionService
DataPlaneLeaseApplier
LockManager
Clock
Logger
```

Target operations:

```text
Acquire(ctx, req)
Release(ctx, req)
ListActive(ctx, req)
ListRecent(ctx, req)
ListHistory(ctx, req)
Get(ctx, req)
Restore(ctx)
CleanupExpired(ctx)
```

Rules:

- `lease.Application` must not depend on `*Runtime`.
- Business predicates must be centralized.
- Store implementations must not decide business workflow.
- Go-layer filtering of all historical JSON rows is not allowed on hot paths.

Acceptance:

- `Runtime` owns lifecycle wiring, not lease business logic.
- HTTP handlers call lease usecase methods only.
- Lease query behavior is testable by reading usecase and repository interfaces without following the whole runtime.

### Phase 4: Refactor Settings and Reconcile

Current defect:

- Settings read path can mutate persisted settings.
- Settings update synchronously performs reconcile and rollback.

Target model:

```text
LoadSettings       pure read
NormalizeSettings  pure function
MigrateSettings    explicit migration/startup operation
UpdateSettings     persist desired state
ApplySettings      background operation
```

State fields should remain minimal and represent current need:

```text
desired_version
applied_version
apply_status
last_error
updated_at
applied_at
```

Target package:

```text
internal/app/settings/
  application.go
  repository.go
  normalizer.go
  validator.go
  projector.go
  apply_worker.go
```

Rules:

- GET settings cannot write storage.
- Desired state is persisted before apply.
- Apply failure records status and is retryable from desired state.
- Rollback buffer remains only the last accepted generated Mihomo config bytes.
- Do not create a second config model.

Acceptance:

- Settings read is side-effect free.
- Save settings returns quickly.
- Dataplane apply slowness is observable through operation status.
- Apply failure does not leave invisible partial state.

### Phase 5: Split Mihomo Config Projection

Current defect:

- Large config files mix validation, projection, provider data, egress profile logic, rendering, and runtime apply concerns.

Target structure:

```text
internal/app/mihomo/
  projection/
    projector.go
    egress_profiles.go
    dns.go
    routing.go
  render/
    renderer.go
  validate/
    validator.go
```

Pipeline:

```text
settings + provider facts + lease facts
  -> projection
  -> validation
  -> render Mihomo config
  -> dataplane apply
```

Rules:

- Projection and render must not directly read storage.
- Render must not call provider APIs.
- Validation must produce user-actionable errors without secrets.
- Proto remains the source for contract models.

Acceptance:

- No single Mihomo config file owns the whole pipeline.
- Each stage has one clear responsibility.
- Runtime apply can report which stage failed.

### Phase 6: Provider Account and Provider Adapter Boundary

Current defect:

- Deleting a provider account may synchronously release or cleanup many blocking leases.

Target flow:

```text
request delete provider account
mark deleting
background release blocking leases
delete account
mark deleted or failed
```

Provider adapter rules:

- Provider-specific branches stay inside provider adapters or capability registry.
- Business usecases consume stable capability interfaces.
- Provider account validation belongs to the plugin boundary.
- Secrets are decrypted only at adapter boundaries.
- Logs use provider/account identifiers only, not reusable credentials.

Acceptance:

- Provider account delete does not block on unbounded lease cleanup.
- Blocking lease checks use efficient repository methods.
- Provider failures are classified and visible without leaking secret data.

### Phase 7: HTTP, Auth, and Dashboard Separation

Target structure:

```text
internal/app/httpapi/
  server.go
  middleware.go
  error.go
  protojson.go
  routes.go

internal/app/auth/
  application.go
  session.go
  cookie.go
  ws_token.go

internal/app/dashboard/
  reverse_proxy.go
  bootstrap.go
  sanitizer.go
```

HTTP rules:

- Handler parses request, calls one usecase, writes response.
- Handler does not access stores, providers, dataplane internals, or global runtime state directly.
- HTTP errors map from application errors in one place.

Auth rules:

```text
public:
  /api/auth/login
  /api/auth/logout
  /api/auth/session

authenticated:
  /api/auth/ws-token
```

Dashboard proxy rules:

- Reverse proxy is initialized once, not created per request.
- Query token/session stripping is centralized.
- Mihomo secret injection is centralized.
- Upstream errors are sanitized.

Acceptance:

- Route ownership is obvious.
- Auth route naming matches real access control.
- Dashboard proxy behavior is reusable and auditable.

### Phase 8: Frontend Module Cleanup

Target structure:

```text
metacubexd-fork/src/proxy-runtime/
  api/
    client.ts
    errors.ts
    leases.ts
    settings.ts
    providers.ts
  composables/
    useProxyRuntimeLeases.ts
    useProxyRuntimePlayground.ts
    useProxyRuntimeSettings.ts
    useProxyRuntimeAuth.ts
  components/
    Playground.vue
    LeasePanel.vue
    ProviderAccountPanel.vue
```

Rules:

- API client owns base URL, credentials, timeout, cancellation, and error normalization.
- Server state and action state are separated.
- Playground lease acquire is independent from history loading.
- Refresh buttons refresh explicit resources only.
- Prefer existing UI library components instead of handwritten replacements.

Acceptance:

- Loading indicators map to exact operations.
- Active and history lease views are independent.
- Backend timeout does not make UI look inert.

### Phase 9: Observability

Required metrics:

```text
proxy_runtime_http_request_duration_seconds
proxy_runtime_lease_list_duration_seconds
proxy_runtime_lease_list_rows_total
proxy_runtime_lease_acquire_duration_seconds
proxy_runtime_lease_release_duration_seconds
proxy_runtime_lease_worker_runs_total
proxy_runtime_lease_worker_failures_total
proxy_runtime_provider_request_duration_seconds
proxy_runtime_dataplane_apply_duration_seconds
proxy_runtime_settings_apply_duration_seconds
```

Required structured log fields:

```text
request_id
operation_id
lease_id
provider_key
provider_account_key
runtime_status
duration_ms
error_code
```

Forbidden output:

```text
token
cookie
mihomo secret
provider password
full proxy URL
session material
subscription URL with token
```

Acceptance:

- A slow UI operation can be traced to lease query, provider request, dataplane apply, or settings apply.
- Backend unavailable and backend slow are distinguishable.
- Sensitive proxy/auth/session material does not appear in logs, metrics, traces, or client errors.

### Phase 10: Postgres and SQLite Store Decision

Current defect:

- Postgres and SQLite stores duplicate dynamic lease query and persistence behavior.

Decision path:

1. Verify whether deployed runtime uses SQLite.
2. Verify whether standalone SQLite mode is still a supported product requirement.
3. If not required, remove SQLite store from the runtime path.
4. If required, keep SQLite as an adapter but prevent duplicated business predicates.

Rules:

- There must be one semantic definition for active, blocking, cleanup-pending, restorable, and session-bound leases.
- Store adapters implement repository contracts; they do not invent independent business behavior.

Acceptance:

- Fixing lease behavior requires changing one semantic layer, not two drifting store implementations.

### Current Execution Status

Snapshot date: 2026-06-15.

Completed user-visible/runtime batches:

- Lease list hot path is bounded and no longer defaults to inactive history.
- Playground acquire no longer waits for full lease history before requesting a lease.
- Frontend proxy-runtime requests have bounded timeout and clearer backend-unreachable handling.
- HTTP startup is decoupled from active lease restore.
- Lease restore and cleanup attempts are bounded.
- Lease list semantics were extracted from ad hoc handler filtering.
- Settings reads are side-effect free.
- Runtime settings apply, Mihomo-native config apply, and provider-account delete now persist desired state and continue through asynchronous reconcile/background work.
- WebSocket token route is treated as authenticated.
- Dashboard proxy creation was centralized/reused.
- Runtime status is exposed by backend and surfaced in the Playground UI.
- First frontend/backend split steps are done for lease API and Mihomo-native update logic.
- Mihomo-native config model, projection persistence, settings mapping, and URI rendering helpers are split into focused files.
- Mihomo-native projection orchestration is separated from native JSON file path/load/save IO.
- Mihomo-native settings mapping is split into focused conversion, normalization, indexing, and subscription-provider rendering files.
- Mihomo-native update now separates pure update-plan construction and validation from persistence, file save, reference replacement, and reconcile scheduling side effects.
- Mihomo-native native JSON file IO now depends on an explicit config directory instead of direct `*Runtime` access.
- Mihomo-native projection orchestration now uses an explicit settings repository port and config directory; `Runtime` only wires dependencies.
- Mihomo-native update now uses explicit repository/config-dir/after-apply dependencies; runtime cache clearing and reconcile scheduling are wiring callbacks instead of embedded update logic.
- Mihomo sourceplane egress-profile rendering is split into line, exit, native-resource, dynamic-exit, and naming helpers.
- Mihomo driver model, reconcile flow, and hot-reload candidate apply logic are separated into focused files.
- Mihomo reconcile now uses explicit rendered-config projection and base-config apply helpers.
- Source-plane dataplane config projection is split into a pure explicit-input builder; runtime wiring only loads settings, builds dynamic pool, and stores the snapshot.
- Source-plane proxy-user route merge/dedup projection moved with the explicit-input projection builder, keeping runtime dataplane config assembly as wiring only.
- Listener default config and proto/local-service projection helpers are split from Runtime route wiring into explicit-input projection functions.
- Lease application now owns list/acquire/release response orchestration through repository/coordinator ports, and HTTP request details are reduced to an advertised host before entering lease orchestration.
- Lease detail lookup is exposed through the lease application repository port and `GET /api/leases/{lease_id}`, so full lease detail can be fetched by ID instead of through list hot paths.
- Lease application construction now uses an explicit dependency object with repository, coordinator, worker, logger, and clock ports; lease list duration/row logging lives in the lease application.
- Lease acquire and release application boundaries now log duration and stable lease/account/provider identifiers without emitting provider/session secrets or raw error text.
- Dynamic lease retry and Mihomo connection cleanup paths no longer log or return raw provider/controller error bodies; logs use safe error types and HTTP status summaries.
- Lease list query parsing for status, legacy inactive mode, and bounded limits is centralized in the lease package; the Gin handler only adapts query values to application input.
- Listener resolution and reserved-listener checks no longer use the old unbounded `ListLeaseFacts`; they use bounded active lease queries plus cleanup-pending facts, and the unbounded store port was removed.
- SQLite provider-account blocking and cleanup-pending lease queries now use SQL/JSON predicates, aligning with the Postgres existence-query behavior instead of loading rows only to filter in Go.
- Active lease lookup by provider session now uses SQL/JSON session-id predicates in both Postgres and SQLite instead of loading all active account leases and filtering in Go.
- Lease package is split into application, repository/coordinator ports, list options, operations, and predicates.
- Lease restore/expire/cleanup worker entrypoints now go through the lease application worker port instead of direct Runtime coordinator calls.
- Lease cleanup label mutation is centralized in the lease package together with cleanup predicates.
- Settings application service forwarding, update usecases, apply scheduling, Mihomo-native handlers, and validation are split into focused files.
- Runtime settings application now carries explicit logger, settings-store, proxy-user, provider-descriptor, Mihomo-native, and apply-scheduler dependencies instead of reaching through `*Runtime` inside usecase methods.
- Runtime settings application now depends on a narrow settings repository port instead of the concrete `runtimeSettingsStore` type.
- Runtime settings store load/save and Mihomo-native persistence facades are split from settings update logic.
- HTTP control-plane route declarations are split by auth, runtime status, provider, lease, check, and settings ownership.
- Runtime auth session signing, verification, token matching, and safe redirects are extracted into `internal/app/auth`.
- Mihomo dashboard/controller URL, query sanitization, cache header, and error redaction helpers are extracted into `internal/app/dashboard`.
- HTTP request helpers for path-prefix matching, forwarded protocol, and request IDs are extracted into `internal/app/httpapi`.
- HTTP route declarations now use the shared `httpapi.Route` model instead of an app-local route shape.
- HTTP request-body and proto JSON codec helpers are extracted into `internal/app/httpapi`, while app-level error mapping stays at the adapter boundary.
- Runtime auth required/public-path rules are centralized in `internal/app/auth`.
- HTTP JSON error response writing is extracted into `internal/app/httpapi`, with app-specific error-to-status mapping kept at the adapter edge.
- Public HTTP route registration now uses the shared `httpapi.Route` model and is split from HTTP server setup.
- Runtime login request parsing is extracted into `internal/app/auth`, leaving the Gin handler as body-read and response adapter.
- Runtime auth session cookie construction and clearing are centralized in `internal/app/auth`.
- MetaCubeXD bootstrap HTML, reverse proxy construction, controller auth injection, query sanitization, cache headers, and upstream error redaction are centralized in `internal/app/dashboard`.
- HTTP request ID, authorization handoff, panic recovery, and request logging middleware are extracted into `internal/app/httpapi`; panic logs no longer include the recovered payload.
- Runtime auth secret handling, required-path checks, login token matching, session verification, session cookies, and WebSocket token minting are routed through `internal/app/auth.Application`.
- Runtime login page rendering and login redirect URL construction are owned by `internal/app/auth`, leaving Gin handlers to set headers and write responses.
- Lease coordinator wiring now uses explicit store, settings, lock, dataplane, provider-session factory, concurrency limiter, logger, and runtime-callback dependencies instead of holding `*Runtime`; acquire/release/restore/expire/cleanup paths no longer dereference the large runtime object directly.
- Lease coordinator time-dependent lease predicates and timestamps now use an injected clock port instead of direct `time.Now()` calls in lease acquire, failed-acquire recording, restore, and expiry logic.
- Provider session creation in lease orchestration is now behind a lease-owned factory port; registry and HTTP client details are confined to the runtime wiring adapter.
- Dynamic lease data-plane operations now depend on a narrow session-route applier port instead of the full dataplane driver surface.
- Dynamic lease orchestration now depends on a narrow lock-manager port that exposes only account, provider-account, and listener-allocation critical sections.
- Lease coordinator logging now depends on the lease logger port instead of the concrete slog logger.
- Dynamic lease failed-acquire, released, expired, cleanup-failure, and cleanup-retry status mutations are centralized in `internal/app/lease` lifecycle helpers; app-level persistence code now saves already-mutated lease facts instead of owning status transitions.
- Dynamic lease concurrency holder, concurrency policy, and dynamic-provider-id extraction are centralized in `internal/app/lease`, keeping lease fact interpretation out of runtime provider concurrency helpers.
- Active dynamic lease fact construction and active/released status checks are centralized in `internal/app/lease`, removing direct status-enum writes and checks from lease runtime orchestration.
- Dynamic IP endpoint health scoring now uses lease status predicates from `internal/app/lease` instead of interpreting active/expired/released/failed enums locally.
- Acquire request session-id label parsing is centralized in `internal/app/lease`, so lease orchestration no longer owns sticky-session label aliases.
- Acquire request account/purpose label injection is centralized in `internal/app/lease`; dynamic IP policy normalization remains in the runtime dynamic-IP layer.
- Lease orchestration store, provider-session factory, data-plane applier, and lock-manager ports are now defined in `internal/app/lease`; the app layer only adapts runtime registry and lock implementations to those ports.
- Provider-session release and stateless-session detection are centralized in `internal/app/lease`, keeping provider session cleanup semantics out of app-level acquire/release failure handling.
- Lease concurrency mode/text and slot TTL calculation are centralized in `internal/app/lease`; provider-account concurrency adapters now use lease policy interpretation from the lease package.
- Dynamic lease endpoint-id extraction is centralized in `internal/app/lease`; dynamic IP endpoint health scoring no longer reads selection/egress/session labels directly.
- Dynamic lease ID generation now goes through an injected lease ID generator port; acquire and failed-acquire persistence no longer call the random package directly.
- Dynamic lease account, purpose, session, provider-account, selection, endpoint, dynamic-provider, and concurrency-holder label keys are centralized in `internal/app/lease`, removing duplicated lease label strings from runtime orchestration.
- Lease coordinator runtime adapters for provider-session factory, lock manager, and ID generation are split from coordinator dependency wiring.
- Source-plane proxy-user configured-route merge and dedup helpers are split from the top-level source-plane dataplane config builder.
- Dynamic lease acquire/release request validation, account/purpose normalization, release lookup parsing, and release lease/account/purpose match validation are centralized in `internal/app/lease`.
- Dynamic lease acquire attempt label writing is centralized in `internal/app/lease`, so runtime acquire retry logic no longer edits raw policy label keys directly.
- Dynamic lease release lookup, release state transition orchestration, provider-session release locking, and dataplane route deletion are split from the acquire service file into a focused release file.
- Dynamic lease egress-profile policy resolution and playground replacement checks are split from the acquire service file into a focused profile helper file.
- Dynamic lease listener allocation, endpoint materialization, dataplane route upsert, active fact persistence, and playground connection cleanup are split from the acquire service file into a focused route-apply file.
- Provider account application now receives explicit repository, settings, descriptor, lock, lease-operation, and logger dependencies; provider usecase methods no longer dereference `*Runtime` directly.
- Runtime check application now receives explicit settings, HTTP-client, exit-IP probe, geo lookup, IP-fraud, edge-canary, and cache dependencies instead of reaching through `*Runtime`.
- Runtime lease application construction now accepts the lease application dependency object directly; `RuntimeService` owns only wiring from `Runtime` to lease ports.
- Runtime settings application construction now uses an explicit dependency object, and settings usecases guard missing repositories/loggers instead of dereferencing app fields directly.

Still open:

- Fully extract lease application into `internal/app/lease`; remaining work is to move provider-session creation, data-plane route apply/delete, listener allocation, locks, and concurrency-slot behavior behind lease-owned ports instead of the current app-level coordinator.
- Continue splitting Mihomo sourceplane projection, validation, render, and apply stages so no single file owns the whole config pipeline.
- Move settings orchestration into an explicit settings application package.
- Separate `httpapi`, `auth`, and `dashboard` packages and keep handlers as thin transport adapters.
- Finish provider adapter capability boundaries and secret-handling audit.
- Expand metrics and structured operation logging for slow paths.
- Decide whether SQLite remains a supported adapter; if it stays, remove duplicated business predicates from store implementations.

### Next Implementation Batches

1. **Finish Mihomo projection split, no behavior change**
   - Move sourceplane projection helpers, native update orchestration, validation, and render/apply boundaries into focused files/packages.
   - Keep generated config output semantically identical.
   - Validate with focused diffs, `gofmt`, stale-reference search, and remote build/deploy validation only when artifacts are required.

2. **Lease application extraction**
   - Introduce `internal/app/lease` ports and usecase methods.
   - Move acquire/release/list/restore/cleanup orchestration out of `Runtime`.
   - Keep active/blocking/restorable predicates in one semantic layer.

3. **Settings application extraction**
   - Separate pure load/normalize/validate from persist/apply.
   - Keep GET side-effect free.
   - Model apply state as desired/applied version plus status/error, without adding a second config source.

4. **HTTP/auth/dashboard cleanup**
   - Move routing, middleware, proto JSON, and error mapping into `internal/app/httpapi`.
   - Move login/session/cookie/ws-token logic into `internal/app/auth`.
   - Move dashboard reverse proxy bootstrapping, Mihomo secret injection, and URL/session sanitization into `internal/app/dashboard`.

5. **Provider boundary and secret hygiene**
   - Keep provider-specific branches inside adapters/capability registry.
   - Decrypt secrets only at adapter boundary.
   - Audit logs, metrics, traces, and client errors for provider credentials, Mihomo secret, proxy passwords, full proxy URLs, and session material.

6. **Observability and store decision**
   - Add missing slow-path metrics for lease list/acquire/release, workers, provider requests, dataplane apply, and settings apply.
   - Verify deployed store mode and remove SQLite from runtime path if it is no longer a product requirement.

### Recommended Commit Sequence

```text
1. optimize lease list query contract and pagination
2. decouple playground acquire from lease history refresh
3. add frontend timeout, cancellation, and error normalization
4. start HTTP before background lease restore
5. move lease restore and cleanup into bounded worker
6. extract lease application from Runtime
7. remove settings load side effects
8. make settings apply and reconcile operation-based
9. split Mihomo projection, validation, and rendering
10. isolate provider account delete as a background operation
11. separate httpapi, auth, and dashboard proxy packages
12. consolidate or remove duplicated SQLite store behavior
13. add runtime status, metrics, and structured slow-path logs
```

### First Implementation Batch

The first batch should be small and user-visible:

1. Replace default full lease history loading with bounded active/recent queries.
2. Remove pre-acquire `leases.load()` from Playground save flow.
3. Add frontend timeout/cancellation in the proxy-runtime API client.
4. Make active lease checks include `expires_at > now`.

Expected impact:

- Playground acquire no longer waits behind history scans.
- Refresh button has visible bounded behavior.
- Login/dashboard no longer appears unavailable because of full lease list latency.
