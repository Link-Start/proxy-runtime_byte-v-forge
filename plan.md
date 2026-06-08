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
