#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROTO_DIR="${PROTO_DIR:-${ROOT}/proto}"
OUT_DIR="${OUT_DIR:-${ROOT}/metacubexd-fork/types}"
PLUGIN="${PROTOC_GEN_TS_PROTO:-}"

if [[ -z "${PLUGIN}" ]]; then
  if [[ -x "${ROOT}/node_modules/.bin/protoc-gen-ts_proto" ]]; then
    PLUGIN="${ROOT}/node_modules/.bin/protoc-gen-ts_proto"
  elif [[ -x "${ROOT}/../webui/node_modules/.bin/protoc-gen-ts_proto" ]]; then
    PLUGIN="${ROOT}/../webui/node_modules/.bin/protoc-gen-ts_proto"
  fi
fi

if [[ -z "${PLUGIN}" || ! -x "${PLUGIN}" ]]; then
  printf 'ts-proto plugin not found; set PROTOC_GEN_TS_PROTO or install protoc-gen-ts_proto\n' >&2
  exit 1
fi

rm -rf "${OUT_DIR}/byte"
mkdir -p "${OUT_DIR}"

protoc -I "${PROTO_DIR}" \
  --plugin="protoc-gen-ts_proto=${PLUGIN}" \
  --ts_proto_out="${OUT_DIR}" \
  --ts_proto_opt=onlyTypes=true,outputServices=none,esModuleInterop=true,useJsonWireFormat=true,snakeToCamel=false \
  "${PROTO_DIR}/byte/v/forge/contracts/common/v1/common.proto" \
  "${PROTO_DIR}/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime.proto"
