# MetaCubeXD Fork Overlay

`proxy-runtime` uses MetaCubeXD as its main frontend. The Docker build clones the upstream MetaCubeXD source at a pinned commit, applies `patches/`, overlays the files in this directory, and builds static assets for Mihomo `external-ui`.

Only project-owned additions live here:

- `patches/001-dynamic-ip-providers-proxies-tab.patch`
- `patches/002-dynamic-leases-connections-tab.patch`
- `patches/003-disable-google-fonts.patch`
- `components/proxy-runtime/ProxyRuntimeDynamicIPProviders.vue`
- `components/proxy-runtime/ProxyRuntimeInUserRules.vue`
- `components/proxy-runtime/DynamicIPProvider*.vue`
- `composables/useProxyRuntime*.ts`
- generated proto contracts copied from `common-lib` during Docker build

The fork keeps upstream MetaCubeXD as the Mihomo UI. Mihomo-native fixed proxies, subscriptions, proxy providers, rules, groups, and config editing stay in upstream pages.

The project overlay adds:

- `入口用户` inside MetaCubeXD `proxies` for proxy username/password, line, and exit bindings.
- `动态IP提供商` inside MetaCubeXD `proxies` for proxy-runtime-only dynamic provider instances, endpoints, and provider accounts.
- `动态租约` inside MetaCubeXD `connections` for active dynamic lease runtime state.

Both use `/api/proxy-runtime/*`. They do not hand-edit the generated Mihomo runtime projection.
