# proxy-runtime

`proxy-runtime` is the proxy control plane for Byte-V Forge. It exposes a stable proxy entry and uses Mihomo as the only data plane.

## Responsibilities

- Run and reconcile Mihomo configuration for the fixed entry listener, proxy users, fixed proxy nodes, subscription providers, proxy groups, and health checks.
- Manage proxy provider adapters, dynamic provider endpoints, provider accounts, dynamic IP sessions, and sticky dynamic leases.
- Present an internal dynamic-proxy-provider style interface: clients use one entry address and proxy username/password to select an egress policy.
- Keep provider control-plane access separate from business data-plane egress.
- Provide runtime observations for providers, source nodes, leases, exit IP checks, IP fraud checks, and edge canary checks.

`proxy-runtime` does not implement a proxy kernel. Mihomo performs traffic forwarding, routing, grouping, and health checks.

## Model

- `ProxyUserRoute`: a registered proxy username that maps inbound traffic to an egress policy.
- `EgressGateway`: the fixed entry exposed to internal applications.
- `ProxyDynamicIPProviderSettings`: dynamic provider endpoint settings owned by `proxy-runtime`.
- `EgressProfileSettings`: two-layer egress profile with a proxy route layer and an exit layer.
- `ProxyDynamicLease`: a sticky dynamic IP session managed by the control plane.
- `MaterializedProxy`: an upstream proxy node rendered into Mihomo for a fixed/source/dynamic route.
- `EgressRoutePlan`: the selected dynamic provider gateway for a dynamic lease.

Traffic enters one Mihomo mixed listener. Mihomo `IN-USER` rules map proxy usernames to proxy groups. Dynamic provider credentials and session parameters are rendered only into Mihomo upstream proxy nodes and are not exposed as business-facing proxy addresses.

Detailed design: `docs/egress-gateway-design.md`.

## Runtime Shape

```text
Internal app
  -> fixed entry host:port
  -> proxy username/password
  -> Mihomo IN-USER rule
  -> egress profile
  -> proxy route
  -> route exit, static IP, or dynamic IP exit
```

Chained egress is configuration-only: `proxy-runtime` stores Egress Profiles as `route + exit` and renders them into Mihomo-native proxy groups, provider overrides, and `dialer-proxy`. `exit=direct` means using the selected route's own exit; it does not create a hidden second hop. There is no GOST runtime and no separate chain proxy kernel.

## Configuration

Runtime environment variables are intentionally small. Most provider, source, proxy user, IP fraud, and canary settings are control-plane data managed through API/UI.

Required bootstrap:

- `PROXY_RUNTIME_POSTGRES_DSN` or `PG_DSN`: proxy-runtime control-plane PostgreSQL DSN.
- `PLATFORM_REDIS_URL`: Redis URL for lease runtime locks.
- `PROXY_RUNTIME_ENCRYPTION_KEY`: encryption key for stored provider credentials and settings secrets.

Optional bootstrap with defaults:

- `PROXY_RUNTIME_ADDR`: HTTP control plane address. Default `:8080`.
- `PROXY_RUNTIME_MIHOMO_PATH`: Mihomo executable path. Default `mihomo`.
- `PROXY_RUNTIME_MIHOMO_CONFIG_DIR`: generated Mihomo config directory. Default `/var/lib/byte-v-forge/proxy-runtime/mihomo`.
- `PROXY_RUNTIME_MIHOMO_API_ADDR`: Mihomo external-controller address. Default `127.0.0.1:18901`.
- `PROXY_RUNTIME_LOCAL_ADDR`: fixed entry listener address. Default `:1080`.
- `PROXY_RUNTIME_LOCAL_PROTOCOL`: fixed entry protocol, `http` or `socks5`. Default `http`.
- `PROXY_RUNTIME_SESSION_ADVERTISED_HOST`: optional host returned when the entry binds to loopback or all interfaces.

Optional local bootstrap:

- `PROXY_RUNTIME_LOCAL_USERNAME` / `PROXY_RUNTIME_LOCAL_PASSWORD`: default entry proxy user.
- `PROXY_RUNTIME_PROXY_USERS_JSON`: initial registered proxy users for fixed entry routing. Prefer API/UI after boot.

Example proxy users:

```sh
PROXY_RUNTIME_PROXY_USERS_JSON='[
  {"id":"crawler-us","username":"crawler-us","password":"crawler-pass","route":"provider"},
  {"id":"crawler-profile","username":"crawler-profile","password":"crawler-pass","route":"profile","profile_id":"profile-us"},
  {"id":"direct","username":"direct","password":"direct-pass","route":"direct"}
]'
```

Provider bootstrap is optional and kept for local bring-up. Production dynamic provider accounts and endpoints should be stored through API/UI:

- `PROXY_RUNTIME_PROVIDER`: base provider, `1024proxy`, `static`, or `none`. Default `1024proxy`.
- `PROXY_RUNTIME_SIMPLE_PROXIES`: static base provider proxies when `PROXY_RUNTIME_PROVIDER=static`.
- `PROXY_RUNTIME_PROVIDER_HTTP_PROXY`: optional HTTP proxy used only for provider control-plane calls.
- `PROXY_RUNTIME_REQUEST_TIMEOUT_SECONDS`: provider HTTP timeout. Default `10`.
- `PROXY_RUNTIME_REFRESH_SECONDS`: reconcile interval. Default `300`.

1024Proxy username/session mode:

- `PROXY_RUNTIME_1024_PROXY_ADDR`: provider endpoint address, for example `us.1024proxy.io:3000`.
- `PROXY_RUNTIME_1024_USERNAME` / `PROXY_RUNTIME_1024_PASSWORD`: provider credentials.
- `PROXY_RUNTIME_1024_PROTOCOL`: upstream protocol, `http` or `socks5`. Default `http`.

1024Proxy API extraction mode:

- `PROXY_RUNTIME_1024_API_URL`: API URL copied from the provider console.
- `PROXY_RUNTIME_1024_API_REGION` / `PROXY_RUNTIME_1024_API_FORMAT` / `PROXY_RUNTIME_1024_API_TIME` / `PROXY_RUNTIME_1024_API_NUM` / `PROXY_RUNTIME_1024_API_TYPE`: API query overrides.

IP fraud, Cloudflare canary, dynamic provider endpoints, and proxy exit IP check settings are managed through `GET/PUT /proxy/settings`. Secret values are write-only inputs and are stored in the service-owned secret store.

## HTTP Endpoints

All endpoints are exposed under both `/proxy/*` and `/api/proxy-runtime/*`.

- `GET /healthz`: process liveness.
- `GET /readyz`: Mihomo data plane readiness.
- `GET /proxy-runtime`: full-page entry that redirects to the forked MetaCubeXD frontend.
- `GET /proxy/providers`: provider capability descriptors.
- `GET /proxy/gateway`: fixed entry, listeners, runtime overview, routes, and pool snapshot.
- `GET /proxy/pool`: current proxy pool snapshot.
- `POST /proxy/refresh`: reconcile providers, sources, active leases, and Mihomo config.
- `GET /proxy/provider-accounts` / `PUT /proxy/provider-accounts` / `DELETE /proxy/provider-accounts`: upstream provider accounts.
- `GET /proxy/sources` / `PUT /proxy/sources` / `DELETE /proxy/sources`: control-plane source metadata for runtime planning.
- `GET /proxy/sources/nodes`: source nodes as observed from Mihomo provider state.
- `GET /proxy/leases`: dynamic IP leases.
- `POST /proxy/leases/acquire`: create or replace a sticky dynamic IP lease; returns the fixed entry endpoint and proxy user route.
- `POST /proxy/leases/release`: release a dynamic IP lease idempotently.
- `POST /proxy/proxies/resolve`: resolve an internal proxy resource.
- `POST /proxy/proxy_exit_ip`: check the exit IP through a configured listener.
- `POST /proxy/proxy_exit_geo`: lookup geo for an IP without proxy egress.
- `POST /proxy/ip_fraud_check`: check IP fraud risk.
- `POST /proxy/check_cf_access_risk`: check edge access risk through the selected egress.
- `POST /proxy/target_connectivity_check`: check target connectivity through the selected egress.
- `GET /proxy/settings` / `PUT /proxy/settings`: runtime settings; responses do not echo token/API key values.
- `GET /proxy/settings/egress-profiles` / `PUT /proxy/settings/egress-profiles`: configured egress profiles rendered into Mihomo native groups and `dialer-proxy` chains.
- `GET /proxy/mihomo/dashboard`: forked MetaCubeXD main frontend bootstrap. It registers the loopback Mihomo controller endpoint and opens the MetaCubeXD `proxies` page.
- `/proxy/mihomo/ui/*`: same-origin reverse proxy to Mihomo `external-ui`.
- `/proxy/mihomo/controller/*`: same-origin reverse proxy to Mihomo external-controller for MetaCubeXD.

## Dashboard

The dashboard uses a project-owned MetaCubeXD fork as the main frontend:

- Upstream MetaCubeXD pages remain the Mihomo operations surface: overview, proxies, proxy providers, rules, connections, logs, config, fixed proxies, subscriptions, and provider updates.
- The project overlay adds `动态代理提供者` and `来源配置` inside MetaCubeXD `proxies` for dynamic provider endpoints, provider accounts, fixed sources, subscriptions, and egress profiles.
- The project overlay adds `动态租约` inside MetaCubeXD `connections` for active dynamic lease runtime state.

The forked MetaCubeXD assets are built into the `proxy-runtime` image and served full-page through same-origin routes. The Byte-V dashboard no longer loads a `proxy-runtime` module-federation frontend. Browsers do not need direct access to the loopback-only Mihomo API.

## Generation

Public proto contracts and generated Go/TypeScript types are owned by `common-lib`:

- Source: `common-lib/proto/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime.proto`
- Go generated package: `common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1`
- TypeScript generated package: `common-lib/ui/src/proto/byte/v/forge/contracts/proxyruntime/v1`

Do not edit generated files manually.

## Verification

Preferred local source-level checks:

```sh
gofmt -w ./cmd ./internal
go vet ./...
```

The aggregate Byte-V Forge environment does not use the Mac for business image builds or deployment validation. Runtime deployment, image build, and environment validation happen on the remote host environment.
