#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
output="$script_dir/../tanks_server"

echo "building $output ..."
go build -o "$output" "$script_dir"
echo "done: $output"
