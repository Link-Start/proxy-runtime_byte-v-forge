# proxy-runtime

`proxy-runtime` 是统一出口代理网关，为业务服务提供稳定入口地址，并把代理商账号、动态 IP 租约、出口策略和 Mihomo 数据面配置集中在本服务内管理。

## 核心能力

- 以“固定网关地址 + 代理用户名/密码”向业务侧提供出口选择能力。
- 统一管理代理用户、出口 profile、动态 IP provider、provider account、租约和并发槽位。
- 生成并协调 Mihomo 配置，让 Mihomo 负责真实转发、认证、规则路由、proxy group 和健康检查。
- 提供出口 IP、地理位置、风控风险、Cloudflare canary 和目标连通性检查。
- 基于项目自有 MetaCubeXD fork 提供代理运维 dashboard，覆盖入口用户、动态 provider、原生配置和运行连接观察。

## 使用方式

业务仓只通过 proxy ref、固定网关账号或契约调用本服务，不接触上游 provider 代理地址、密码、session material 或动态租约细节。provider 控制面访问与业务数据面出口在本服务内分离建模。

## 控制面鉴权

设置 `PROXY_RUNTIME_CONTROL_AUTH_TOKEN` 后，MetaCubeXD 静态 UI、后台配置 API 与 `/mihomo/controller/*` 需要先通过 `/login` 登录。登录成功后服务端下发 `HttpOnly` session cookie；浏览器不保存 Mihomo controller secret，proxy-runtime 在反向代理到 Mihomo 时内部注入 controller 鉴权。`/api/leases/acquire`、`/api/leases/release` 和出口检测类 API 保持服务间机器调用入口，不依赖浏览器登录态。

## 入口

- 服务入口：`cmd/proxy-runtime`
- 契约真源：`proto/byte/v/forge/contracts/proxyruntime/v1/`
- 控制面实现：`internal/app/`
- Mihomo 数据面适配：`internal/dataplane/`、`internal/sourceplane/`
- Dashboard fork：`metacubexd-fork/`

## 常用检查

```sh
sh scripts/generate-proto.sh
sh scripts/generate-web-proto.sh
gofmt -w ./cmd ./internal
git diff --check
```
