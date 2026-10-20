# forgejo stack module

- Module id: `forgejo`
- Module repo: `forgejo-stack-module`
- Source repo: none declared
- Lifecycle: `active`

## Owned overlays
- `stack.runtime.yaml`
- `stack.config/forgejo`

## Dependencies
- `stack-foundation`

## Validation

```sh
./tests/validate.sh
```

## Data-root ownership

The root bootstrap/token writer and dropped git server are different identities.
Before starting either, the wrapper assigns only the mounted `/data` parent to
`USER_UID:USER_GID` (default1000:1000). This lets the server traverse a private0700
fresh volume; upstream still prepares its children. No recursive chown, permission
widening or `/runner-token` ownership change occurs. Invalid identities or a linked/
non-directory root refuse. Fixtures cover ownership/order/preserved children and
negative paths; installed Forgejo/OIDC proof remains separate.

## Lifecycle

`active` modules are expected to keep `stack.module.json`, owned overlays, and `tests/validate.sh` in sync.
