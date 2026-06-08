# Contributing

- 保持本仓只承载代理基础设施能力。
- provider 差异通过 adapter、配置和契约表达，不写入业务分支。
- secret、token、代理密码和可复用会话材料不得写入日志、示例或提交历史。
- 代理契约位于本仓 `proto/`；修改后运行 `sh scripts/generate-proto.sh`，需要前端类型时运行 `sh scripts/generate-web-proto.sh`。
- 优先使用 Go 标准库、官方 SDK 或成熟社区 SDK；不要复制外部仓的通用 helper。
