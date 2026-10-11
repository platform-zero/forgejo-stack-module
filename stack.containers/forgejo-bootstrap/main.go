// Reconcile only the secret of a native Forgejo 14 OAuth source.
// The native CLI owns source creation/registration; credentials never enter argv.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func patchConfig(raw, issuer, secret string) (string, error) {
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return "", errors.New("invalid source configuration")
	}
	for key, expected := range map[string]string{"Provider": "openidConnect", "ClientID": "forgejo", "OpenIDConnectAutoDiscoveryURL": issuer} {
		var actual string
		if json.Unmarshal(cfg[key], &actual) != nil || actual != expected {
			return "", errors.New("source identity mismatch")
		}
	}
	if len(secret) < 32 || len(secret) > 4096 || strings.ContainsAny(secret, "\r\n\x00") {
		return "", errors.New("invalid secret input")
	}
	value, _ := json.Marshal(secret)
	cfg["ClientSecret"] = value
	updated, err := json.Marshal(cfg)
	return string(updated), err
}

func reconcile() error {
	if len(os.Args) != 2 {
		return errors.New("source id required")
	}
	id, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil || id <= 0 {
		return errors.New("invalid source id")
	}
	domain := os.Getenv("DOMAIN")
	if domain == "" || strings.ContainsAny(domain, "/:@?#\\\r\n\x00") {
		return errors.New("invalid domain")
	}
	if os.Getenv("FORGEJO__database__DB_TYPE") != "postgres" {
		return errors.New("unsupported database")
	}
	host := os.Getenv("FORGEJO__database__HOST")
	user := os.Getenv("FORGEJO__database__USER")
	password := os.Getenv("FORGEJO__database__PASSWD")
	name := os.Getenv("FORGEJO__database__NAME")
	if host == "" || user == "" || password == "" || name == "" {
		return errors.New("missing private database input")
	}
	if !strings.Contains(host, ":") {
		host += ":5432"
	}
	dsn := url.URL{Scheme: "postgres", Host: host, User: url.UserPassword(user, password), Path: "/" + name}
	query := url.Values{}
	sslmode := os.Getenv("FORGEJO__database__SSL_MODE")
	if sslmode == "" {
		sslmode = "disable"
	} // Same native private-network default; OIDC TLS remains verified.
	query.Set("sslmode", sslmode)
	query.Set("connect_timeout", "10")
	dsn.RawQuery = query.Encode()
	db, err := sql.Open("postgres", dsn.String())
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	sourceName := os.Getenv("FORGEJO_EXPECTED_AUTH_SOURCE_NAME")
	if sourceName != "Keycloak" && sourceName != "Keycloak migrated" {
		return errors.New("invalid source name")
	}
	var raw string
	if err = tx.QueryRowContext(ctx, `SELECT cfg FROM login_source WHERE id=$1 AND type=6 AND name=$2 AND is_active=true FOR UPDATE`, id, sourceName).Scan(&raw); err != nil {
		return err
	}
	issuer := "https://keycloak." + domain + "/realms/webservices/.well-known/openid-configuration"
	updated, err := patchConfig(raw, issuer, os.Getenv("FORGEJO_OAUTH_SECRET"))
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `UPDATE login_source SET cfg=$1, updated_unix=EXTRACT(EPOCH FROM NOW())::bigint WHERE id=$2 AND cfg=$3 AND type=6 AND name=$4 AND is_active=true`, updated, id, raw, sourceName)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return errors.New("source compare-and-swap refused")
	}
	return tx.Commit()
}

func main() {
	if reconcile() != nil {
		// Never print driver errors, DSNs, configuration, or private values.
		fmt.Fprintln(os.Stderr, "[forgejo-bootstrap] private credential reconciliation refused")
		os.Exit(1)
	}
}
