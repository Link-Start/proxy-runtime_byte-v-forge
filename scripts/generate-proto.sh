#!/usr/bin/env sh
set -eu

ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
PATH="$(go env GOPATH)/bin:$PATH"

rm -rf "$ROOT/gen"
mkdir -p "$ROOT/gen/go"

protoc -I "$ROOT/proto" \
  --go_out="$ROOT" \
  --go_opt=module=github.com/byte-v-forge/proxy-runtime \
  --go-grpc_out="$ROOT" \
  --go-grpc_opt=module=github.com/byte-v-forge/proxy-runtime \
  "$ROOT/proto/byte/v/forge/contracts/common/v1/common.proto" \
  "$ROOT/proto/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime.proto"

gofmt -w "$ROOT/gen/go"
