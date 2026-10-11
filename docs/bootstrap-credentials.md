# Private Forgejo bootstrap credentials

The derived image keeps the exact Forgejo 14.0.2 base manifest
`sha256:d198af9a52220f2c933c73a8ab8e58e388387c9fb217bbd9cb792964521ac6d7`.
It adds one static, source-built PostgreSQL configuration helper, not another
service or a replacement Forgejo version. The Go builder and driver are pinned
by immutable manifest and `go.sum` respectively.

Forgejo 14's native `admin auth add-oauth/update-oauth --secret` exposes secrets
in process arguments; it has no private stdin/environment equivalent. Bootstrap
therefore uses the native CLI for source creation and public configuration,
reconciles only `ClientSecret` using the schema-pinned helper, and invokes native
`update-oauth` again without `--secret` to register the reconciled source.

The helper reads the existing `FORGEJO_OAUTH_SECRET` and native PostgreSQL inputs
from the service's private environment. It locks exactly one active OAuth2
`login_source` by positive ID and expected Keycloak name, verifies provider,
client ID and the exact declared discovery URL, preserves every other JSON field,
and uses a parameterized compare-and-swap transaction. Mismatched identity,
unsupported database, malformed input, missing rows and stale updates fail
closed. Driver errors, DSNs and configuration are never printed. PostgreSQL
transport follows the native application's declared SSL mode; OIDC always uses
verified TLS with the live public CA and declared host-gateway alias.

Existing users and passwords are preserved. Fresh native bootstrap users receive
an unsaved random password through Forgejo's own `--random-password` mode, with
its credential output discarded; bootstrap API tokens do not require that
password. Existing duplicate authentication sources and all user references are
preserved rather than deleted. The legacy bootstrap password file is not removed.

API authorization headers pass through private file descriptors, not argv.
Readiness requires completion of private OIDC configuration during the current
container startup; the non-secret marker is removed before startup and written
only after successful native registration. Immutable installation, functional
SSO and rollback verification remain separate acceptance steps.
