#!/bin/bash
set -e
ready_file=/data/gitea/conf/.p0-oidc-ready
if [ -L "$ready_file" ] || { [ -e "$ready_file" ] && ! [ -f "$ready_file" ]; }; then
    echo '[forgejo-bootstrap] refusing unrelated readiness path' >&2
    exit 1
fi
if [ -f "$ready_file" ] && ! grep -qx 'p0-forgejo-oidc-ready-v1' "$ready_file"; then
    echo '[forgejo-bootstrap] preserving unrelated readiness file' >&2
    exit 1
fi
rm -f "$ready_file"
/usr/bin/entrypoint 2> >(grep -v "Read-only file system" >&2) &
FORGEJO_PID=$!
/generate-runner-token.sh &

(
    while true; do
        if /init-forgejo.sh; then
            exit 0
        fi
        echo "[forgejo-entrypoint] Forgejo bootstrap failed; retrying in 30s" >&2
        sleep 30
    done
) &
wait "$FORGEJO_PID"
