#!/bin/bash
set -e
prepare_forgejo_data_root() {
    local path="$1" uid="${USER_UID:-1000}" gid="${USER_GID:-1000}"
    if ! [[ "$uid" =~ ^[0-9]+$ && "$gid" =~ ^[0-9]+$ ]] || [ -L "$path" ] || ! [ -d "$path" ]; then
        echo '[forgejo-bootstrap] refusing invalid data root or server identity' >&2
        return 1
    fi
    # The root wrapper/token writer is not the git server identity. Podman's
    # private fresh volume can be root0700; the upstream entrypoint prepares
    # children but the dropped server UID must also traverse this parent.
    # Change only this one mounted directory, never recursively or /runner-token.
    chown -- "$uid:$gid" "$path"
}
prepare_forgejo_data_root /data
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
