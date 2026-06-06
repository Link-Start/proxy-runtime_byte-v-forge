# Mihomo-only Egress Design

`proxy-runtime` is a proxy control plane. Mihomo is the data plane.

## Product Shape

The runtime follows the common dynamic proxy provider model:

```text
fixed entry host:port + proxy username/password -> egress policy
```

Applications do not receive upstream provider proxy URLs. They receive a stable entry endpoint and a proxy user. The username selects a registered route, and the password authenticates access to the entry.

## Data Plane

Mihomo owns:

- inbound mixed listener
- proxy user authentication
- `IN-USER` routing rules
- static proxies
- proxy providers
- proxy groups
- health checks
- connection lifecycle through the external controller
- forked MetaCubeXD main frontend through same-origin routes

The control plane renders Mihomo config and reloads through the external controller. It never implements HTTP CONNECT/SOCKS forwarding itself.

The external controller remains loopback-only. `proxy-runtime` exposes same-origin controller routes for the forked MetaCubeXD frontend, so browser clients use the service HTTP origin instead of directly reaching Mihomo.

## Control Plane

`proxy-runtime` owns:

- provider adapters and provider account credentials
- dynamic provider endpoint settings
- dynamic provider session creation and release
- sticky dynamic lease facts
- dynamic IP endpoint selection
- Mihomo config rendering
- runtime observations and control-plane APIs
- project-owned MetaCubeXD fork as the main frontend
- same-origin business APIs used by the dynamic provider tab in that fork

The forked frontend keeps Mihomo-native runtime operations in upstream MetaCubeXD pages. The project overlay edits dynamic IP provider endpoints/accounts, IN-USER rules, and active leases through `proxy-runtime` APIs; reconcile renders only the proxy-runtime-owned facts into Mihomo.

## Proxy User Routes

`ProxyUserRoute` maps a proxy username to an egress policy.

Supported MVP route targets:

- `direct`: Mihomo `DIRECT`
- `profile`: a configured egress profile rendered as Mihomo-native groups and `dialer-proxy`
- dynamic session routes created by `AcquireProxyLease`

Mihomo renders these as:

```yaml
listeners:
  - name: proxy-runtime-gateway
    type: mixed
    listen: 0.0.0.0
    port: 1080
    users:
      - username: crawler-us
        password: crawler-pass

rules:
  - IN-USER,crawler-us,bvf-profile-profile-us
  - MATCH,REJECT
```

## Egress Profiles

An Egress Profile is the only supported chain model. It has two layers:

```text
route: direct or a Mihomo-selected node
exit: route exit, Mihomo-selected node, or dynamic IP provider pool
```

At render time, `proxy-runtime` projects the profile into Mihomo:

- a selected Mihomo node becomes a hidden route proxy group
- `exit=direct` selects the route group, or Mihomo `DIRECT` when the route is direct
- static IP exits are Mihomo-native nodes and require a direct route; proxy-runtime does not clone Mihomo-native proxies or proxy providers
- dynamic IP exits use the dynamic provider pool; when a route node is selected, pool nodes are rendered with `dialer-proxy`
- the final exit group name is used by `IN-USER` rules for `route=profile`

There is no hidden second hop. Additional hops must be explicit profile configuration and are rendered as Mihomo-native groups or `dialer-proxy`, not as a separate chain resolver API or second runtime model.

## Dynamic Lease Materialization

Dynamic IP is a lease in the control plane and a normal proxy node in Mihomo.

Flow:

```text
AcquireProxyLease
  -> choose provider account/provider endpoint
  -> create provider session
  -> fetch upstream HTTP/SOCKS node
  -> create Mihomo materialized proxy
  -> add IN-USER rule for the lease proxy user
```

Release removes the materialized route and releases the provider session. Restore fetches active lease sessions from provider adapters and recreates the Mihomo route.

Fixed proxies, subscriptions, proxy providers, rules, and groups stay as Mihomo-native resources.

## Reliability

- Reconcile is idempotent and renders the whole desired data-plane state.
- Mihomo config changes prefer external-controller hot reload. Restart is only a fallback when the process is not running, the gateway endpoint changes, or hot reload cannot recover.
- The last accepted generated Mihomo runtime config bytes are kept only as a rollback buffer for failed reloads. Control-plane facts in the database remain the single desired-state model.
- Active leases are persisted in PostgreSQL and restored after restart.
- Lease/runtime cleanup failures are persisted and retried by reconcile.
- Provider control-plane HTTP calls use configured timeouts and proxy settings.
- Logs and errors must not include provider passwords, API tokens, cookies, or full upstream proxy URLs.
