package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPatchOnlySelectedSecret(t *testing.T) {
	issuer := "https://keycloak.example.test/realms/webservices/.well-known/openid-configuration"
	raw := `{"Provider":"openidConnect","ClientID":"forgejo","ClientSecret":"old","OpenIDConnectAutoDiscoveryURL":"` + issuer + `","AdminGroup":"owners","UnknownFutureField":{"keep":true}}`
	updated, err := patchConfig(raw, issuer, strings.Repeat("x", 32))
	if err != nil {
		t.Fatal("valid native source refused")
	}
	var before, after map[string]json.RawMessage
	json.Unmarshal([]byte(raw), &before)
	json.Unmarshal([]byte(updated), &after)
	for key, value := range before {
		if key != "ClientSecret" && string(value) != string(after[key]) {
			t.Fatal("unrelated source field changed")
		}
	}
	if len(before) != len(after) {
		t.Fatal("unrelated source field added or removed")
	}
}

func TestRefuseUnrelatedSourcesAndInvalidInputs(t *testing.T) {
	issuer := "https://keycloak.example.test/realms/webservices/.well-known/openid-configuration"
	raw := `{"Provider":"openidConnect","ClientID":"forgejo","OpenIDConnectAutoDiscoveryURL":"` + issuer + `"}`
	for _, bad := range []string{`{`, strings.Replace(raw, "forgejo", "another", 1), strings.Replace(raw, "openidConnect", "github", 1), strings.Replace(raw, "example.test", "other.test", 1)} {
		if _, err := patchConfig(bad, issuer, strings.Repeat("x", 32)); err == nil {
			t.Fatal("unrelated or malformed native source accepted")
		}
	}
	for _, secret := range []string{"", "short", strings.Repeat("x", 5000), strings.Repeat("x", 32) + "\n"} {
		if _, err := patchConfig(raw, issuer, secret); err == nil {
			t.Fatal("invalid private input accepted")
		}
	}
}
