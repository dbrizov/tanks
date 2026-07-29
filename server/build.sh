#!/usr/bin/env bash
# Rebuild the tanks_server executable from the Go sources in this directory.
set -euo pipefail

# Resolve the directory this script lives in (the server/ source dir), so the
# script works no matter what the current working directory is.
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Emit the binary one level up: when the sources live in /opt/tanks/server the
# binary lands at /opt/tanks/tanks_server, matching _docs/PUBLISH_HOME.md.
output="$script_dir/../tanks_server"

echo "building $output ..."
go build -o "$output" "$script_dir"
echo "done: $output"
