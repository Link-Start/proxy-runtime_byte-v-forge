# proxy-runtime

`proxy-runtime` is an independent proxy control plane and gateway. It exposes a stable proxy entry and uses Mihomo as the only data plane.

## Responsibilities

- Run and reconcile Mihomo configuration for the fixed entry listener, proxy users, dynamic IP lease materialization, and egress profiles.
- Manage proxy provider adapters, dynamic provider endpoints, provider accounts, dynamic IP sessions, and sticky dynamic leases.
- Present an internal dynamic-proxy-provider style interface: clients use one entry address and proxy username/password to select an egress policy.
- Keep provider control-plane access separate from business data-plane egress.
- Provide runtime observations for dynamic IP providers, leases, exit IP checks, IP fraud checks, and edge canary checks.

`proxy-runtime` does not implement a proxy kernel. Mihomo performs traffic forwarding, routing, grouping, and health checks.

## Model

- `ProxyUserRoute`: a registered proxy username that maps inbound traffic to an egress policy.
- Fixed entry listener: the Mihomo mixed listener exposed to internal applications.
- `ProxyDynamicIPProviderSettings`: dynamic provider endpoint settings owned by `proxy-runtime`.
- `EgressProfileSettings`: two-layer egress profile with a line layer and an exit layer.
- `ProxyDynamicLease`: a dynamic IP lease managed by the control plane; PlayGround uses a single fixed IN-USER user plus one active lease at most.
- `ProxyDynamicIPSelectionPlan`: the selected provider account and endpoint for a dynamic lease.

Traffic enters one Mihomo mixed listener. Mihomo `IN-USER` rules map proxy usernames to proxy groups. PlayGround persists one fixed IN-USER user, `playground`; when dynamic IP is needed it acquires one active lease and routes that same user to the lease. Dynamic provider credentials and session parameters are rendered only into Mihomo upstream proxy nodes and are not exposed as business-facing proxy addresses.

Detailed design: `docs/egress-gateway-design.md`.

## Runtime Shape

```text
Internal app
  -> fixed entry host:port
  -> proxy username/password
  -> Mihomo IN-USER rule
  -> egress profile
  -> line layer
  -> route exit, static IP, or dynamic IP exit
```

Chained egress is configuration-only: `proxy-runtime` stores Egress Profiles as `route + exit` and renders them into Mihomo-native proxy groups, provider overrides, and `dialer-proxy`. `exit=direct` means using the selected route's own exit; it does not create a hidden second hop. There is no GOST runtime and no separate chain proxy kernel.

## Configuration

Runtime environment variables are intentionally small. Dynamic IP provider, proxy user, IP fraud, and canary settings are control-plane data managed through API/UI. Mihomo-native static proxies, proxy-providers, rules, groups, and node operations remain Mihomo config/UI concerns.

Required bootstrap:

- `PROXY_RUNTIME_ENCRYPTION_KEY`: encryption key for stored provider credentials and settings secrets.

Storage and runtime coordination:

- `PROXY_RUNTIME_POSTGRES_DSN` or `PG_DSN`: optional PostgreSQL DSN. When omitted, `proxy-runtime` uses its embedded SQLite control store.
- `PROXY_RUNTIME_DATA_DIR`: local data directory for embedded SQLite and runtime files when PostgreSQL is omitted. Default `/var/lib/proxy-runtime`.
- `PROXY_RUNTIME_REDIS_URL`: optional Redis URL for distributed lease locks and provider-account concurrency slots.
- When Redis is omitted, lease locks and concurrency slots use an in-process adapter. This is suitable for standalone single-replica runtime; multi-replica deployments should configure Redis.

Optional bootstrap with defaults:

- `PROXY_RUNTIME_ADDR`: HTTP control plane address. Default `:8080`.
- `PROXY_RUNTIME_MIHOMO_PATH`: Mihomo executable path. Default `mihomo`.
- `PROXY_RUNTIME_MIHOMO_CONFIG_DIR`: generated Mihomo config directory. Default `/var/lib/proxy-runtime/mihomo`.
- `PROXY_RUNTIME_MIHOMO_API_ADDR`: Mihomo external-controller address. Default `127.0.0.1:18901`.
- `PROXY_RUNTIME_LOCAL_ADDR`: fixed entry listener address. Default `:1080`.
- `PROXY_RUNTIME_LOCAL_PROTOCOL`: fixed entry protocol, `http` or `socks5`. Default `http`.
- `PROXY_RUNTIME_SESSION_ADVERTISED_HOST`: optional host returned when the entry binds to loopback or all interfaces.

Optional local bootstrap:

- `PROXY_RUNTIME_LOCAL_PASSWORD`: password used only by the explicit lease API for temporary proxy users. Business applications should prefer fixed gateway IN-USER rules.
- `PROXY_RUNTIME_PROXY_USERS_JSON`: initial registered proxy users for fixed entry routing. Only `direct` and `profile` routes are accepted; prefer API/UI after boot.

Example proxy users:

```sh
PROXY_RUNTIME_PROXY_USERS_JSON='[
  {"id":"crawler-profile","username":"crawler-profile","password":"crawler-pass","route":"profile","profile_id":"profile-us"},
  {"id":"direct","username":"direct","password":"direct-pass","route":"direct"}
]'
```

Provider bootstrap is optional and kept for local bring-up. Production dynamic provider accounts and endpoints should be stored through API/UI:

- `PROXY_RUNTIME_PROVIDER`: base provider, `1024proxy` or `none`. Default `1024proxy`.
- `PROXY_RUNTIME_PROVIDER_HTTP_PROXY`: optional HTTP proxy used only for provider control-plane calls.
- `PROXY_RUNTIME_REQUEST_TIMEOUT_SECONDS`: provider HTTP timeout. Default `10`.
- `PROXY_RUNTIME_REFRESH_SECONDS`: reconcile interval. Default `300`.

1024Proxy username/session mode:

- `PROXY_RUNTIME_1024_USERNAME` / `PROXY_RUNTIME_1024_PASSWORD`: provider credentials.
- `PROXY_RUNTIME_1024_PROTOCOL`: upstream protocol, `http` or `socks5`. Default `http`.

1024Proxy API extraction mode:

- `PROXY_RUNTIME_1024_API_URL`: API URL copied from the provider console.
- `PROXY_RUNTIME_1024_API_REGION` / `PROXY_RUNTIME_1024_API_FORMAT` / `PROXY_RUNTIME_1024_API_TIME` / `PROXY_RUNTIME_1024_API_NUM` / `PROXY_RUNTIME_1024_API_TYPE`: API query overrides.

IP fraud, Cloudflare canary, dynamic provider endpoints, and proxy exit IP check settings are managed through `GET/PUT /api/settings`. Secret values are write-only inputs and are stored in the service-owned secret store.


## Standalone Image

The repository Dockerfile is self-contained and does not require a sibling checkout. Build and deployment validation for the Byte-V Forge environment still happen on the remote host, but a standalone image can run with SQLite by default:

```sh
docker run --rm \
  -p 8080:8080 \
  -p 1080:1080 \
  -v proxy-runtime-data:/var/lib/proxy-runtime \
  -e PROXY_RUNTIME_ENCRYPTION_KEY=change-me-at-least-32-characters \
  proxy-runtime:local
```

## HTTP Endpoints

The standalone HTTP surface uses root UI, `/api/*` control-plane routes, and `/mihomo/*` Mihomo reverse-proxy routes.

- `GET /`: MetaCubeXD fork frontend entry.
- `GET /healthz`: process liveness.
- `GET /readyz`: Mihomo data plane readiness.
- `GET /api/providers`: provider capability descriptors.
- `GET /api/provider-accounts` / `PUT /api/provider-accounts` / `DELETE /api/provider-accounts`: upstream provider accounts.
- `GET /api/leases`: dynamic IP leases.
- `POST /api/leases/acquire`: explicit lease tooling endpoint. Business applications and PlayGround sticky profiles should not depend on it for normal egress.
- `POST /api/leases/release`: release an explicit lease idempotently.
- `POST /api/proxy_exit_ip`: check the exit IP through a configured listener.
- `POST /api/proxy_exit_geo`: lookup geo for an IP without proxy egress.
- `POST /api/ip_fraud_check`: check IP fraud risk.
- `POST /api/check_cf_access_risk`: check edge access risk through the selected egress.
- `POST /api/target_connectivity_check`: check target connectivity through the selected egress.
- `GET /api/settings` / `PUT /api/settings`: runtime settings; responses do not echo token/API key values.
- `GET /api/settings/in-user-rules` / `PUT /api/settings/in-user-rules`: proxy username/password rules with line and exit settings.
- `GET /mihomo/dashboard`: MetaCubeXD controller bootstrap.
- `/mihomo/ui/*`: same-origin reverse proxy to Mihomo `external-ui`.
- `/mihomo/controller/*`: same-origin reverse proxy to Mihomo external-controller for MetaCubeXD.

## Dashboard

The dashboard uses a project-owned MetaCubeXD fork as the main frontend:

- Upstream MetaCubeXD pages remain the Mihomo operations surface: overview, proxies, proxy providers, rules, connections, logs, config, fixed proxies, subscriptions, and provider updates.
- The project overlay adds `入口用户` and `动态IP提供商` inside MetaCubeXD `proxies` for proxy username/password routing plus dynamic provider endpoints/accounts.
- The project overlay adds `动态租约` inside MetaCubeXD `connections` for active dynamic lease runtime state. PlayGround also shows its own single active lease in the PlayGround page.

The forked MetaCubeXD assets are built into the `proxy-runtime` image and served full-page through same-origin routes. Browsers do not need direct access to the loopback-only Mihomo API.

## Generation

Proto contracts and generated Go/TypeScript types are owned by this repository:

- Source: `proto/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime.proto`
- Go generated package: `gen/go/byte/v/forge/contracts/proxyruntime/v1`
- TypeScript generated package: `metacubexd-fork/types/byte/v/forge/contracts/proxyruntime/v1`

Do not edit generated files manually. Run `sh scripts/generate-proto.sh` after proto changes; run `sh scripts/generate-web-proto.sh` when the frontend type output must be regenerated.

## Verification

Preferred local source-level checks:

```sh
gofmt -w ./cmd ./internal
git diff --check
```

The aggregate Byte-V Forge environment does not use the Mac for business image builds or deployment validation. Runtime deployment, image build, and environment validation happen on the remote host environment.
