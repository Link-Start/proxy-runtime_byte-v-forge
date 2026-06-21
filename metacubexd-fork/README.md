# MetaCubeXD Fork Overlay

`proxy-gateway` uses MetaCubeXD as its main frontend. The Docker build clones the upstream MetaCubeXD source at a pinned commit, applies `patches/`, overlays the files in this directory, and builds static assets for Mihomo `external-ui`.

Only project-owned additions live here:

- `patches/001-dynamic-ip-providers-proxies-tab.patch`
- `patches/002-dynamic-leases-connections-tab.patch`
- `patches/003-disable-google-fonts.patch`
- `components/proxy-gateway/ProxyGatewayDynamicIPProviders.vue`
- `components/proxy-gateway/ProxyGatewayInUserRules.vue`
- `components/proxy-gateway/ProxyGatewayMetrics.vue`
- `components/proxy-gateway/DynamicIPProvider*.vue`
- `composables/useProxyGateway*.ts`
- generated proto contracts from `metacubexd-fork/types/`

The fork keeps upstream MetaCubeXD as the Mihomo UI. Mihomo-native fixed proxies, subscriptions, proxy providers, rules, groups, and config editing stay in upstream pages.

The project overlay adds:

- `入口用户` inside MetaCubeXD `proxies` for proxy username/password, line, and exit bindings.
- `动态IP提供商` inside MetaCubeXD `proxies` for proxy-gateway-only dynamic provider instances, endpoints, and provider accounts.
- `动态租约` inside MetaCubeXD `connections` for active dynamic lease runtime state.
- `观测` tab inside MetaCubeXD `overview` for proxy-gateway operation metrics.

Both use `/api/*`. They do not hand-edit the generated Mihomo runtime projection.
