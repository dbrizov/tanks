#!/usr/bin/env bash
# One-time host setup for the tanks deployment. Run as root from the repo root
# (which is expected to be /opt/tanks)
#     sudo ./setup.sh
#
# Creates the service user, the config/state directories, seeds the game server
# config, and writes both systemd unit files with a shared JWT secret. It does
# not enable/start the services (the venv and binary aren't ready yet — that's
# Parts 2 and 3). Idempotent — safe to re-run; existing files are left as-is.
set -euo pipefail

# --- Deployment-specific settings -------------------------------------------
# The origin the browser client is served from (used for CORS in the auth
# service). Change this to match where the client is hosted.
CORS_ORIGIN="https://denisrizov.com"
# ----------------------------------------------------------------------------

if [ "$(id -u)" -ne 0 ]; then
    echo "setup.sh must be run as root (use: sudo ./setup.sh)" >&2
    exit 1
fi

# The repo root is wherever this script lives.
repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

auth_unit="/etc/systemd/system/tanks-auth.service"
server_unit="/etc/systemd/system/tanks-server.service"

# 1. Service user — a system account with no login shell.
if id -u tanks >/dev/null 2>&1; then
    echo "user 'tanks' already exists"
else
    useradd --system --home /opt/tanks --shell /usr/sbin/nologin tanks
    echo "created user 'tanks'"
fi

# 2. Directories, one per role (Linux FHS):
#   /opt/tanks      code + binary                  — owned by the service user
#   /etc/tanks      static config (config.json)    — root-owned, admin-edited
#   /var/lib/tanks  app-written state (auth.db)    — owned by the service user
mkdir -p /etc/tanks /var/lib/tanks
chown -R tanks:tanks "$repo_dir" /var/lib/tanks
echo "directories ready: /etc/tanks, /var/lib/tanks"

# 3. Seed the game server config from the repo template if it isn't there yet.
#    The config holds no secrets (just game tuning), so it stays root-owned and
#    world-readable; JWT_SECRET lives only in the unit files (below).
if [ -f /etc/tanks/config.json ]; then
    echo "/etc/tanks/config.json already present — left untouched"
else
    cp "$repo_dir/server/config.json" /etc/tanks/config.json
    chown root:root /etc/tanks/config.json
    chmod 644 /etc/tanks/config.json
    echo "seeded /etc/tanks/config.json from server/config.json"
fi

# 4. Shared JWT secret. Reuse the one already in an existing unit so re-runs
#    don't invalidate live tokens; otherwise generate a fresh one.
jwt_secret=""
for unit in "$auth_unit" "$server_unit"; do
    if [ -f "$unit" ]; then
        jwt_secret="$(grep -oP '^Environment=JWT_SECRET=\K.*' "$unit" | head -n1 || true)"
        [ -n "$jwt_secret" ] && break
    fi
done
if [ -z "$jwt_secret" ]; then
    jwt_secret="$(openssl rand -base64 48)"
    echo "generated a new shared JWT secret"
fi

# 5. systemd unit files — written only if missing (never clobbered, so the
#    secret and any hand-edits survive re-runs).
if [ -f "$auth_unit" ]; then
    echo "$auth_unit already present — left untouched"
else
    cat > "$auth_unit" <<EOF
[Unit]
Description=Tanks auth service (Flask/Gunicorn)
After=network.target

[Service]
User=tanks
Group=tanks
WorkingDirectory=/opt/tanks/auth
Environment=APP_ENV=prod
Environment=JWT_SECRET=$jwt_secret
Environment=DB_PATH=/var/lib/tanks/auth.db
Environment=APP_CORS_ALLOWED_ORIGINS=$CORS_ORIGIN
ExecStart=/opt/tanks/auth/.venv/bin/gunicorn -w 2 -b 127.0.0.1:8100 wsgi:app
Restart=always

[Install]
WantedBy=multi-user.target
EOF
    chmod 644 "$auth_unit"
    echo "wrote $auth_unit"
fi

if [ -f "$server_unit" ]; then
    echo "$server_unit already present — left untouched"
else
    cat > "$server_unit" <<EOF
[Unit]
Description=Tanks game server (Go)
After=network.target

[Service]
User=tanks
Group=tanks
WorkingDirectory=/opt/tanks/server
Environment=CONFIG=/etc/tanks/config.json
Environment=PORT=8101
Environment=JWT_SECRET=$jwt_secret
ExecStart=/opt/tanks/tanks_server
Restart=always

[Install]
WantedBy=multi-user.target
EOF
    chmod 644 "$server_unit"
    echo "wrote $server_unit"
fi

systemctl daemon-reload

echo
echo "setup complete. Next:"
echo "  - Part 2: install the auth venv, then 'sudo systemctl enable --now tanks-auth'"
echo "  - Part 3: build the server binary, then 'sudo systemctl enable --now tanks-server'"
