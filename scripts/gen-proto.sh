#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROTO_DIR="${ROOT_DIR}/proto"
GO_OUT_DIR="${ROOT_DIR}/services/api/proto"

mkdir -p "${GO_OUT_DIR}"
rm -f "${GO_OUT_DIR}"/*.pb.go

protoc \
  -I "${PROTO_DIR}" \
  --go_out="${GO_OUT_DIR}" \
  --go_opt=paths=source_relative \
  "${PROTO_DIR}"/*.proto

if command -v protoc-go-inject-tag >/dev/null 2>&1; then
  protoc-go-inject-tag -input="${GO_OUT_DIR}"/*.pb.go
fi

echo "Generated Go proto files into ${GO_OUT_DIR}"
