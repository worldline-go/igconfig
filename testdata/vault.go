/*
Turns out Vault provides testing functionality to set Vault server from code.
See: https://github.com/hashicorp/vault/tree/master/vault

Unfortunately it is not easy to get it working,
as Vault's dependencies(for it's own packages) in it's root go.mod are a bit screwed.

So once those dependencies will be fixed - it would be good to actually write implementation of it here.

For example see: https://stackoverflow.com/a/57773764
*/
package testdata

import (
	"os"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
)

func NewVaultClient(t *testing.T) *api.Client {
	t.Helper()

	if os.Getenv("VAULT_ADDR") == "" || os.Getenv("VAULT_TOKEN") == "" {
		// Please see testdata/vault.go for additional info
		t.Skip("test currently works with Vault instance")
	}
	conf := api.DefaultConfig()
	conf.Address = os.Getenv("VAULT_ADDR")

	cl, err := api.NewClient(conf)
	require.NoError(t, err)

	cl.SetToken(os.Getenv("VAULT_TOKEN"))

	return cl
}
