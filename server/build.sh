#!/usr/bin/env bash
# Rebuild the tanks_server executable from the Go sources in this directory.
# Run as the service user so the binary is owned by tanks:
#     sudo -u tanks /opt/tanks/server/build.sh
#
# Host setup (service user, directories, config) is done once by setup.sh
set -euo pipefail

# Resolve the directory this script lives in (the server/ source dir), so the
# script works no matter what the current working directory is.
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Emit the binary one level up: when the sources live in /opt/tanks/server
output="$script_dir/../tanks_server"

echo "building $output ..."
go build -o "$output" "$script_dir"
echo "done: $output"
