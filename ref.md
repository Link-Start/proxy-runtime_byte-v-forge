对，你这个判断是对的：**不应该从头构建一整套代理内核**。更合理的做法是：

```text
Mihomo / Clash.Meta = 数据面
你自己的服务 = 控制面 + 动态 IP 适配层
```

也就是说，Clash/Mihomo 负责真实转发、规则、分组、健康检查、订阅源、静态节点、链式代理；你只负责把“动态 IP / 粘性 Session / 租约”这种 Clash 原生不理解的东西，**翻译成 Clash 能理解的静态节点或代理组**。

---

# 1. 为什么不用从头做

Mihomo 已经覆盖了你说的大部分“静态代理规划”能力。

它支持 `proxy-providers`，可以从 `http`、`file`、`inline` 三类来源加载节点，并且支持更新间隔、健康检查、过滤、覆盖字段等配置。订阅源、机场节点、自建 VPS 节点都可以收敛到这一层。([Metacubex][1])

它的 `proxy-groups` 可以引用普通代理、其他代理组，或者通过 `use` 引用 proxy-provider；并且支持健康检查 URL、检查间隔、失败阈值、过滤等通用字段。([Metacubex][2])

它还支持 `select`、`url-test`、`fallback`、`load-balance` 等策略组；`load-balance` 里还有 `round-robin`、`consistent-hashing`、`sticky-sessions` 这些策略。注意这里的 `sticky-sessions` 是“同源/同目标在 10 分钟缓存期内命中同一已配置节点”，不是代理商那种“新建一个动态住宅 IP session”。([Metacubex][3])

链式代理也不需要你自己实现。Mihomo 现在更推荐用 `dialer-proxy`，它允许一个代理节点通过另一个代理或代理组建立连接；官方文档也提示 `relay` 策略即将废弃，应使用 `dialer-proxy`。([Metacubex][4])

所以你的系统不应该是：

```text
应用
  ↓
你自研完整代理内核
  ↓
各种代理资源
```

而应该是：

```text
应用
  ↓
你的 Proxy Allocation API / Proxy Control Plane
  ↓
Mihomo / Clash.Meta
  ↓
静态代理 / 订阅源 / 自建 VPS / 动态代理商
```

---

# 2. 你真正需要建的东西

你只需要建一个**动态代理控制面**，而不是代理内核。

最小模型可以简化成 4 个对象。

---

## 1）ProviderConnector

表示代理来源。

```text
ProviderConnector
├── id
├── name
├── type
│   ├── subscription      # 机场订阅
│   ├── static_proxy      # 固定 HTTP/SOCKS 节点
│   ├── vps               # 自建 VPS 节点
│   └── dynamic_vendor    # 动态 IP 代理商
├── protocol
├── auth_config
├── fetch_config
└── capabilities
```

这里不要过度设计 `Endpoint`、`ResourcePool`、`Account`。你不是售卖，而是给内部应用用，核心是“某个 Provider 能提供什么能力”。

---

## 2）AppRoute

表示某个内部应用要怎么出网。

```text
AppRoute
├── id
├── app_id
├── inbound
│   ├── port
│   ├── username
│   └── listener_name
├── policy
│   ├── country
│   ├── network_type
│   ├── sticky_required
│   ├── chain_required
│   └── fallback
└── mihomo_group_name
```

Mihomo 的规则支持按 `IN-PORT`、`IN-TYPE`、`IN-USER`、`IN-NAME` 匹配入口流量，所以内部应用最好通过**不同端口、不同用户名、不同 listener name**来区分，而不是依赖进程名；进程名规则虽然也支持，但在容器和服务端环境里通常不如入口维度稳定。([Metacubex][5])

示例：

```yaml
listeners:
  - name: app-crawler
    type: http
    port: 18080
    listen: 0.0.0.0
    users:
      - username: crawler
        password: crawler-pass

rules:
  - IN-USER,crawler,APP_CRAWLER_US
  - MATCH,DEFAULT
```

Mihomo 的 HTTP/SOCKS listener 都支持用户名密码配置。([Metacubex][6])

---

## 3）DynamicLease

只给动态粘性 IP 用。

```text
DynamicLease
├── id
├── app_id
├── provider_id
├── session_key
├── country
├── city
├── isp
├── created_at
├── expires_at
├── state
│   ├── active
│   ├── expired
│   ├── revoked
│   └── failed
└── materialized_proxy_name
```

这个对象就是你之前说的“租约式”。
`new session` 的本质不是让 Clash 换 IP，而是：

```text
创建新的 session_key
  ↓
生成新的上游代理凭证 / 节点
  ↓
让 Mihomo 把它当成一个普通静态节点使用
```

---

## 4）MaterializedProxy

表示“控制面生成出来，交给 Mihomo 使用的节点”。

```text
MaterializedProxy
├── name
├── provider_id
├── lease_id
├── type
│   ├── http
│   ├── socks5
│   ├── ss
│   ├── vmess
│   └── vless
├── server
├── port
├── username
├── password
├── dialer_proxy
├── tags
└── expires_at
```

Mihomo 支持 HTTP 和 SOCKS 出站节点的 `server`、`port`、`username`、`password` 等字段，所以很多动态代理商可以被“伪装成普通静态 HTTP/SOCKS 节点”。([Metacubex][7])

---

# 3. 动态 IP 的三种适配方式

## 方式 A：动态代理商支持 username/session 编码

这是最推荐的方式。

很多动态代理不是每次给你一个固定 IP，而是给你一个固定网关：

```text
gateway.provider.com:10000
```

真正的国家、城市、session 信息通过用户名表达：

```text
user-country-us-session-abc123
```

那你只要把它 materialize 成 Mihomo 节点：

```yaml
proxies:
  - name: dyn-us-sess-abc123
    type: http
    server: gateway.provider.com
    port: 10000
    username: "user-country-us-session-abc123"
    password: "secret"
```

对 Mihomo 来说，这是一个普通 HTTP 代理；对代理商来说，这是一个动态粘性 session。

`new session` 时，你只需要生成：

```text
dyn-us-sess-def456
```

而不是重启 Mihomo 或修改复杂路由。

---

## 方式 B：你暴露一个动态 proxy-provider

你的控制面提供一个 provider YAML：

```text
GET /mihomo/providers/dynamic-us.yaml
```

返回内容：

```yaml
proxies:
  - name: dyn-us-sess-abc123
    type: http
    server: gateway.provider.com
    port: 10000
    username: "user-country-us-session-abc123"
    password: "secret"

  - name: dyn-us-sess-def456
    type: http
    server: gateway.provider.com
    port: 10000
    username: "user-country-us-session-def456"
    password: "secret"
```

Mihomo 配置：

```yaml
proxy-providers:
  dynamic-us:
    type: http
    url: "http://proxy-control.internal/mihomo/providers/dynamic-us.yaml"
    path: ./providers/dynamic-us.yaml
    interval: 60
    health-check:
      enable: true
      url: https://www.gstatic.com/generate_204
      interval: 300
      timeout: 5000

proxy-groups:
  APP_CRAWLER_US:
    type: select
    use:
      - dynamic-us
```

Mihomo 的 proxy-provider 原生支持 HTTP provider、更新间隔、健康检查、过滤和 override；你可以把动态租约定期渲染成 provider 文件。([Metacubex][1])

当你创建或释放 lease 后，可以通过 Mihomo 的 External Controller 更新 provider；API 文档里有 `/providers/proxies/{providers_name}` 的更新接口，也有 `/configs` reload、`/proxies` 查询、`/connections` 连接管理等接口。([Metacubex][8])

---

## 方式 C：本地动态代理适配器

当代理商不是简单 username/session 模式，而是必须调用 API 才能分配 IP，或者需要复杂鉴权时，可以在本地放一个很薄的 adapter：

```text
Mihomo
  ↓
dynamic-adapter:18081
  ↓
代理商 API / 动态代理网关
```

对 Mihomo 暴露：

```yaml
proxies:
  - name: dyn-adapter-us
    type: http
    server: 127.0.0.1
    port: 18081
```

adapter 内部做：

```text
app_id / session_id
  ↓
找到 DynamicLease
  ↓
选择上游代理商 session
  ↓
CONNECT 到真实代理商
```

这个方式比 A/B 更灵活，但它要求你实现或引入一个 HTTP CONNECT/SOCKS 转发器。只有在动态供应商太复杂时才需要。

---

# 4. 链式代理怎么处理

不要再基于 `relay` 设计新系统。Mihomo 文档明确提示 `relay` 即将废弃，并建议使用 `dialer-proxy`。`dialer-proxy` 的语义是：当前代理节点通过指定代理或代理组建立网络连接。([Metacubex][4])

例如你想：

```text
先走自建 VPS
再连接动态住宅代理
最终目标看到动态住宅 IP
```

可以这样：

```yaml
proxies:
  - name: vps-jp
    type: socks5
    server: vps.example.com
    port: 1080
    username: vps-user
    password: vps-pass

  - name: dyn-us-sess-abc123
    type: http
    server: gateway.provider.com
    port: 10000
    username: "user-country-us-session-abc123"
    password: "secret"
    dialer-proxy: vps-jp
```

这个链路大致是：

```text
App
  ↓
Mihomo
  ↓
vps-jp
  ↓
dynamic provider gateway
  ↓
target site
```

最终目标看到的是动态住宅代理的出口 IP，而不是 VPS IP。Mihomo 文档里的例子也说明了 `dialer-proxy` 下“最终可见 IP”和“连接路径”的区别。([Metacubex][4])

反过来，如果你想：

```text
先走动态住宅
再连接 VPS
最终目标看到 VPS IP
```

那就把 `dialer-proxy` 放在 VPS 节点上：

```yaml
proxies:
  - name: dyn-us-sess-abc123
    type: http
    server: gateway.provider.com
    port: 10000
    username: "user-country-us-session-abc123"
    password: "secret"

  - name: vps-final
    type: socks5
    server: vps.example.com
    port: 1080
    username: vps-user
    password: vps-pass
    dialer-proxy: dyn-us-sess-abc123
```

---

# 5. 推荐的最终架构

我建议你的系统拆成这样：

```text
                    ┌────────────────────┐
                    │ Internal Apps       │
                    │ crawler / worker    │
                    └─────────┬──────────┘
                              │
                              │ HTTP/SOCKS proxy
                              │ or allocation API
                              ▼
                    ┌────────────────────┐
                    │ Proxy Control API   │
                    │ - app route         │
                    │ - lease manager     │
                    │ - provider renderer │
                    └─────────┬──────────┘
                              │
                              │ update provider / reload / select
                              ▼
                    ┌────────────────────┐
                    │ Mihomo / Clash.Meta │
                    │ - routing rules     │
                    │ - proxy groups      │
                    │ - health check      │
                    │ - dialer-proxy      │
                    └─────────┬──────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
  static proxy          subscription nodes      dynamic vendor
  VPS / SOCKS           airport / provider      sticky / rotating
```

Mihomo 的 External Controller 可以通过 REST API 控制内核，配置里有 `external-controller` 和 `secret`；生产环境建议只绑定内网或 `127.0.0.1`，并配置 secret。([Metacubex][9])

---

# 6. 应用侧 API 设计

你可以对内部应用暴露一个非常简单的 API。

## 申请一个代理

```http
POST /proxy/leases
```

请求：

```json
{
  "app_id": "crawler-a",
  "country": "US",
  "type": "sticky",
  "ttl_seconds": 1800,
  "chain": ["vps-jp", "dynamic-us"]
}
```

返回：

```json
{
  "lease_id": "lease_01",
  "session_key": "abc123",
  "proxy": {
    "scheme": "http",
    "host": "mihomo.internal",
    "port": 18080,
    "username": "crawler",
    "password": "crawler-pass"
  },
  "mihomo_group": "APP_CRAWLER_US",
  "materialized_proxy": "dyn-us-sess-abc123",
  "expires_at": "2026-06-04T12:30:00-07:00"
}
```

---

## 申请 new session

```http
POST /proxy/leases/lease_01/rotate
```

内部做：

```text
old: dyn-us-sess-abc123
new: dyn-us-sess-def456
```

然后更新 provider：

```text
dynamic-us.yaml
```

再调用 Mihomo provider update API。

---

## 释放租约

```http
DELETE /proxy/leases/lease_01
```

内部做：

```text
标记 lease expired
从 provider YAML 移除节点
关闭相关连接
```

Mihomo API 支持查看和关闭连接，`/connections` 可以获取连接信息，`DELETE /connections` 可以关闭连接。([Metacubex][8])

---

# 7. 能力边界表

| 能力                                 |          交给 Mihomo |             你自己做 |
| ---------------------------------- | -----------------: | ---------------: |
| 静态 HTTP/SOCKS 节点                   |                  是 |            只生成配置 |
| 机场订阅源                              |                  是 |       拉取、分类、清洗可选 |
| 自建 VPS                             |                  是 |           维护节点信息 |
| 规则路由                               |                  是 |             生成规则 |
| 按应用区分出口                            |                  是 |       设计入口用户名/端口 |
| 健康检查                               |                  是 |        收集结果、降级策略 |
| fallback / url-test / load-balance |                  是 |             选择策略 |
| 链式代理                               | 是，用 `dialer-proxy` |           生成链式配置 |
| rotating IP                        |         半交给 Mihomo |       你把它建模成普通节点 |
| sticky dynamic IP                  |                  否 | 你做 lease/session |
| new session                        |                  否 | 你生成新 session_key |
| 租约过期                               |                  否 |    你清理 lease 和节点 |
| 供应商 API 调用                         |                  否 |       你做 adapter |

---

# 8. 最小可落地版本

你的 MVP 可以这样做：

```text
1. 部署 Mihomo
2. 把机场订阅、自建 VPS、固定代理都接成 proxy-provider
3. 用 IN-USER / IN-PORT 区分内部应用
4. 自己做一个 proxy-control 服务
5. dynamic-control 只负责生成 dynamic provider YAML
6. sticky IP = DynamicLease + session_key + materialized proxy
7. new session = 生成新节点 + 更新 provider
8. 链式 = 给节点加 dialer-proxy
```

这个方案的关键是：**动态 IP 在你的系统里是 lease，在 Mihomo 里只是一个普通 proxy node**。

---

# 9. 什么时候不要强行塞进 Mihomo

有三种情况不适合完全走 provider YAML：

第一，动态 session 数量巨大，比如几万、几十万个短租约 session。频繁渲染 YAML 和更新 provider 会变重。

第二，你需要“每个 HTTP 请求都显式指定 session”，而不是“每个应用/任务绑定一个 session”。这时用一个本地 dynamic-adapter 更合适。

第三，代理商的动态 IP 必须通过复杂 API 预分配、续租、释放，且不能简单通过 username/session 表达。这时还是用 adapter 封装供应商差异。

除此之外，**绝大多数静态能力和中等规模动态 sticky session，都可以基于 Mihomo 做，不需要从头造轮子**。

[1]: https://wiki.metacubex.one/en/config/proxy-providers/?utm_source=chatgpt.com "proxy-providers configuration - mihomo docs"
[2]: https://wiki.metacubex.one/en/config/proxy-groups/ "proxy-groups configuration - mihomo docs"
[3]: https://wiki.metacubex.one/en/config/proxy-groups/load-balance/ "Load-Balance - mihomo docs"
[4]: https://wiki.metacubex.one/en/config/proxies/dialer-proxy/ "dialer-proxy - mihomo docs"
[5]: https://wiki.metacubex.one/en/config/rules/ "Route Rules - mihomo docs"
[6]: https://wiki.metacubex.one/en/config/inbound/listeners/http/ "http - mihomo docs"
[7]: https://wiki.metacubex.one/en/config/proxies/http/ "HTTP - mihomo docs"
[8]: https://wiki.metacubex.one/en/api/ "APIs - mihomo docs"
[9]: https://wiki.metacubex.one/en/config/general/ "General configuration - mihomo docs"
