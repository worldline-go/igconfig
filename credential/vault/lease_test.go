package vault_test

import (
	"context"
	"testing"
	"time"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/credential/vault"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/testdata"
)

func TestGetCredentialsWithClient(t *testing.T) {
	cl := testdata.NewVaultClient(t)

	s, err := vault.GetCredentialsWithClient(cl, vault.Database("test_one"))
	assert.NoError(t, err)
	require.NotNil(t, s)
	assert.NotEmpty(t, *s)

	//t.Logf("%#v", s)
}

func TestKeepRenewed(t *testing.T) {
	cl := testdata.NewVaultClient(t)

	s, err := vault.GetCredentialsWithReader(cl.Logical().Read, vault.Database("test_one"))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	start := time.Now()
	assert.NoError(t, vault.KeepRenewed(ctx, cl, s))

	assert.WithinDuration(t, time.Now(), start.Add(1*time.Second), 300*time.Millisecond)
}

func TestKeepRenewedError(t *testing.T) {
	cl := testdata.NewVaultClient(t)

	s, err := vault.GetCredentialsWithReader(cl.Logical().Read, vault.Database("test_one"))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cl.ClearToken()
	assert.Equal(t, &api.ResponseError{
		HTTPMethod: "PUT",
		URL:        cl.Address() + "/v1/sys/leases/renew",
		StatusCode: 400,
		RawError:   false,
		Errors:     []string{"missing client token"},
	}, vault.KeepRenewed(ctx, cl, s))
}
