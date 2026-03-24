#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WASM_EXEC_JS="$(go env GOROOT)/lib/wasm/wasm_exec.js"

if [[ ! -f "${WASM_EXEC_JS}" ]]; then
  WASM_EXEC_JS="$(go env GOROOT)/misc/wasm/wasm_exec.js"
fi

cp "${WASM_EXEC_JS}" "${ROOT_DIR}/wasm_exec.js"
GOOS=js GOARCH=wasm go build \
  -trimpath \
  -ldflags='-s -w' \
  -o "${ROOT_DIR}/car-extractor.wasm" \
  ./myapp/cmd/carwasm

printf 'Built %s and %s\n' "${ROOT_DIR}/wasm_exec.js" "${ROOT_DIR}/car-extractor.wasm"
