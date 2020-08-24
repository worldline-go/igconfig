package loader

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
)

type inner struct {
	Field2 string `secret:"field_2"`
}

type testStruct struct {
	Field1   string `secret:"field_1"`
	Untagged int64
	NoSet    string `secret:"-"`
	NoData   string `secret:"missing"`
	Time     time.Time
	Inner    inner `secret:"other"`
}

func TestVault_Load(t *testing.T) {
	mock := VaultMock{
		data: map[string]map[string]interface{}{
			"test": {
				"field_1":  "one",
				"untagged": "54",
				"noset":    "not_empty",
			},
			"test/other": {
				"field_2": "other",
			},
		},
	}

	v := Vault{
		Client: mock,
	}

	var s testStruct

	assert.NoError(t, v.Load("test", &s))
	assert.Equal(t, testStruct{
		Field1:   "one",
		Untagged: 54,
		Inner: inner{
			Field2: "other",
		},
	}, s)
}

func TestVault_LoadMissingData(t *testing.T) {
	mock := VaultMock{
		data: map[string]map[string]interface{}{
			"test": {
				"field_1":  "one",
				"untagged": "54",
			},
		},
	}

	v := Vault{
		Client: mock,
	}

	var s testStruct

	assert.NoError(t, v.Load("test", &s))
	assert.Equal(t, testStruct{
		Field1:   "one",
		Untagged: 54,
	}, s)
}

func TestVault_LoadGeneric(t *testing.T) {
	mock := VaultMock{
		data: map[string]map[string]interface{}{
			"generic": {
				"field_1":  "one",
				"untagged": "54",
			},
			"generic/other": {
				"field_2": "other",
			},
		},
	}

	v := Vault{
		Client: mock,
	}

	var s testStruct

	assert.NoError(t, v.LoadGeneric(&s))
	assert.Equal(t, testStruct{
		Field1:   "one",
		Untagged: 54,
		Inner: inner{
			Field2: "other",
		},
	}, s)
}

func TestSimpleVaultLoad(t *testing.T) {
	// This test requires for Vault to be running and token to be known
	addr, token := os.Getenv("VAULT_HOST"), os.Getenv("VAULT_TOKEN")

	if addr == "" || token == "" {
		t.Skip("vault address and token must be provided")
	}

	var s testStruct

	require.NoError(t, SimpleVaultLoad(addr, token, "test", &s))
	assert.Equal(t, testStruct{
		Field1: "one",
		Inner: inner{
			Field2: "other",
		},
	}, s)
}

type VaultMock struct {
	data map[string]map[string]interface{}
	err  error
}

func (v VaultMock) Read(path string) (*api.Secret, error) {
	if v.err != nil {
		return nil, v.err
	}

	path = strings.TrimPrefix(path, VaultSecretBasePath)

	data := v.data[path]
	if data == nil {
		return nil, nil
	}

	secret := api.Secret{
		Data: map[string]interface{}{
			"data": data,
		},
	}

	return &secret, nil
}

func TestNewVaulterer_RoleID(t *testing.T) {
	// This test requires for Vault to be running and token to be known
	addr, roleID := os.Getenv("VAULT_HOST"), os.Getenv("ROLE_ID")

	if addr == "" || roleID == "" {
		t.Skip("vault address and role_id must be provided")
	}

	v, err := NewVaulterer(addr, SetAppRole(roleID, ""))
	assert.NoError(t, err)

	var s testStruct

	require.NoError(t, v.Load("test", &s))
	assert.Equal(t, testStruct{
		Field1: "one",
		Inner: inner{
			Field2: "other",
		},
	}, s)
}

func TestNewVaulterer_Token(t *testing.T) {
	// This test requires for Vault to be running and token to be known
	addr, token := os.Getenv("VAULT_HOST"), os.Getenv("VAULT_TOKEN")

	if addr == "" || token == "" {
		t.Skip("vault address and token must be provided")
	}

	v, err := NewVaulterer(addr, SetToken(token))
	assert.NoError(t, err)

	var s testStruct

	require.NoError(t, v.Load("test", &s))
	assert.Equal(t, testStruct{
		Field1: "one",
		Inner: inner{
			Field2: "other",
		},
	}, s)
}
