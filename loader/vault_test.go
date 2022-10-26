package loader

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	consulApi "github.com/hashicorp/consul/api"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
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
	secrets := map[string]interface{}{
		"field_1": "one",
		"noset":   "not_empty",
		"other": map[string]interface{}{
			"field_2": "other",
		},
		"untagged": 54,
	}

	mock := VaultMock{
		data: map[string]interface{}{
			"data/generic": secrets,
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

func TestVault_LoadUndefined(t *testing.T) {
	mock := VaultMock{
		data: nil,
	}

	v := Vault{
		Client: mock,
	}

	var s testStruct

	assert.NoError(t, v.Load("test", &s))
	assert.Equal(t, testStruct{
		Field1:   "",
		Untagged: 0,
		Inner: inner{
			Field2: "",
		},
	}, s)
}

func TestVault_LoadDestroyed(t *testing.T) {
	mock := VaultMock{
		metadata: map[string]interface{}{
			"destroyed": true,
		},
	}

	v := Vault{
		Client: mock,
	}

	var s testStruct

	assert.NoError(t, v.Load("test", &s))
	assert.Equal(t, testStruct{
		Field1:   "",
		Untagged: 0,
		Inner: inner{
			Field2: "",
		},
	}, s)
}

func TestVault_LoadMissingData(t *testing.T) {
	secrets := map[string]interface{}{
		"field_1":  "one",
		"untagged": 54,
	}

	mock := VaultMock{
		data: map[string]interface{}{
			"data/generic": secrets,
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
	secrets := map[string]interface{}{
		"field_1": "one",
		"other": map[string]interface{}{
			"field_2": "other",
		},
		"untagged": 54,
	}

	mock := VaultMock{
		data: map[string]interface{}{
			"data/generic": secrets,
		},
	}

	v := Vault{
		Client: mock,
	}

	var s testStruct

	assert.NoError(t, v.LoadGeneric(context.Background(), &s))
	assert.Equal(t, testStruct{
		Field1:   "one",
		Untagged: 54,
		Inner: inner{
			Field2: "other",
		},
	}, s)
}

func TestVault_LoadAdditional(t *testing.T) {
	secrets := map[string]interface{}{
		"other": map[string]interface{}{
			"field_2": "other",
		},
		"untagged": 54,
	}

	secretAdditional := map[string]interface{}{
		"field_1": "one",
	}

	mock := VaultMock{
		data: map[string]interface{}{
			"data/generic":    secrets,
			"data/additional": secretAdditional,
		},
	}

	defVaultSecretAdditionalPaths := VaultSecretAdditionalPaths
	defer func() {
		VaultSecretAdditionalPaths = defVaultSecretAdditionalPaths
	}()

	VaultSecretAdditionalPaths = append(VaultSecretAdditionalPaths,
		AdditionalPath{Map: "", Name: "additional"},
	)

	v := Vault{
		Client: mock,
	}

	var s testStruct

	assert.NoError(t, v.LoadGeneric(context.Background(), &s))
	assert.Equal(t, testStruct{
		Field1:   "one",
		Untagged: 54,
		Inner: inner{
			Field2: "other",
		},
	}, s)
}

func TestVault_WithoutGeneric(t *testing.T) {
	secrets := map[string]interface{}{
		"other": map[string]interface{}{
			"field_2": "other",
		},
		"untagged": 54,
	}

	secretAdditional := map[string]interface{}{
		"field_1": "one",
	}

	mock := VaultMock{
		data: map[string]interface{}{
			"data/generic":    secrets,
			"data/additional": secretAdditional,
		},
	}

	defVaultSecretAdditionalPaths := VaultSecretAdditionalPaths
	defer func() {
		VaultSecretAdditionalPaths = defVaultSecretAdditionalPaths
	}()

	VaultSecretAdditionalPaths = []AdditionalPath{}

	v := Vault{
		Client: mock,
	}

	var s testStruct

	assert.NoError(t, v.LoadGeneric(context.Background(), &s))
	assert.Equal(t, testStruct{}, s)
}

func TestVault_LoadList(t *testing.T) {
	keycloak := map[string]interface{}{
		"field_1": "one",
		"other": map[string]interface{}{
			"field_2": "other",
		},
		"untagged": 54,
	}

	gen := map[string]interface{}{
		"field_1": "two",
		"other": map[string]interface{}{
			"field_2": "other2",
		},
		"untagged": 77,
	}

	mock := VaultMock{
		data: map[string]interface{}{
			"data/generic/keycloak": keycloak,
			"data/generic/gen":      gen,
		},
		list: map[string][]interface{}{
			"metadata/generic": {"keycloak", "gen"},
		},
	}

	v := Vault{
		Client: mock,
	}

	type structRest struct {
		Gen      testStruct
		Keycloak testStruct
	}

	s := structRest{}

	assert.NoError(t, v.LoadGeneric(context.Background(), &s))
	assert.Equal(t, structRest{
		Keycloak: testStruct{
			Field1: "one",
			Inner: inner{
				Field2: "other",
			},
			Untagged: 54,
		},
		Gen: testStruct{
			Field1: "two",
			Inner: inner{
				Field2: "other2",
			},
			Untagged: 77,
		},
	}, s)
}

func TestFetchVaultAddrFromConsul(t *testing.T) {
	cl, err := api.NewClient(&api.Config{})
	require.NoError(t, err)

	require.NoError(t, FetchVaultAddrFromConsul(context.Background(), cl, func(ctx context.Context, name string, tags []string) ([]*consulApi.ServiceEntry, error) {
		return []*consulApi.ServiceEntry{
			{Service: &consulApi.AgentService{Address: "set_me", Port: 9090}},
		}, nil
	}))

	assert.Equal(t, "https://set_me:9090", cl.Address())
}

func TestFetchVaultAddrFromConsul_DoNotUpdate(t *testing.T) {
	cl, err := api.NewClient(&api.Config{Address: "do_not_change"})
	require.NoError(t, err)

	require.NoError(t, FetchVaultAddrFromConsul(context.Background(), cl, func(ctx context.Context, name string, tags []string) ([]*consulApi.ServiceEntry, error) {
		return nil, nil
	}))

	assert.Equal(t, "do_not_change", cl.Address())
}

func TestFetchVaultAddrFromConsul_RandomDistribution(t *testing.T) {
	cl, err := api.NewClient(&api.Config{})
	require.NoError(t, err)

	distribution := map[string]int{}

	services := []*consulApi.ServiceEntry{
		{Service: &consulApi.AgentService{Address: "set_me", Port: 9090}},
		{Service: &consulApi.AgentService{Address: "set_me2", Port: 9090}},
		{Service: &consulApi.AgentService{Address: "set_me3", Port: 9090}},
		{Service: &consulApi.AgentService{Address: "set_me4", Port: 9090}},
		{Service: &consulApi.AgentService{Address: "set_me5", Port: 9090}},
	}

	const numTimes = 20000
	for i := 0; i < numTimes; i++ {
		_ = FetchVaultAddrFromConsul(context.Background(), cl, func(ctx context.Context, name string, tags []string) ([]*consulApi.ServiceEntry, error) {
			return services, nil
		})

		distribution[cl.Address()]++
	}

	require.Len(t, distribution, len(services))

	minPercent := 80.0 / float64(len(services)) // 80 == 100 * 0.8, which means 20% of deviation per option. Chosen arbitrary.
	// This checks that no one distribution
	for addr, nums := range distribution {
		percent := 100 * float64(nums) / numTimes
		assert.Greater(t, percent, minPercent, "not evenly distributed:", addr)
	}
}

func TestSimpleVaultLoad(t *testing.T) {
	// This test requires for Vault to be running and token to be known
	addr, token := os.Getenv(api.EnvVaultAddress), os.Getenv(api.EnvVaultToken)

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
	data     map[string]interface{}
	list     map[string][]interface{}
	err      error
	metadata map[string]interface{}
}

func (v VaultMock) Read(path string) (*api.Secret, error) {
	if v.err != nil {
		return nil, v.err
	}

	path = strings.TrimPrefix(strings.TrimPrefix(path, VaultSecretBasePath), "/")

	data, ok := v.data[path]
	if !ok {
		return nil, nil
	}

	secret := api.Secret{
		Data: map[string]interface{}{
			"data": data,
		},
	}

	return &secret, nil
}

func (v VaultMock) List(path string) (*api.Secret, error) {
	path = strings.TrimPrefix(strings.TrimPrefix(path, VaultSecretBasePath), "/")

	if keys, ok := v.list[path]; ok {
		return &api.Secret{
			Data: map[string]interface{}{
				"keys": keys,
			},
		}, nil
	}

	return nil, nil
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

func TestVault_LoginPath(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		expected string
	}{
		{
			name:     "login_on_default_path",
			env:      "",
			expected: VaultAppRoleBasePath,
		},
		{
			name:     "login_on_custom_path",
			env:      "auth/finops/approle/login",
			expected: "auth/finops/approle/login",
		},
	}
	for _, scenario := range tests {
		func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, scenario.expected, r.URL.Path)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			err := os.Setenv("VAULT_ADDR", server.URL)
			assert.NoError(t, err)

			err = os.Setenv(VaultAppRoleBasePathEnv, scenario.env)
			assert.NoError(t, err)

			SetAppRole("dummy-role-id", "dummy-secret-id")
		}()
	}
}

func TestVault_SecretPath(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
	}{
		{
			name: "get_secrets_on_default_path",
		},
		{
			name:     "get_secrets_on_custom_path",
			basePath: "finops/kv2",
		},
	}
	for _, scenario := range tests {
		func() {
			expectedBasePath := VaultSecretBasePath
			if scenario.basePath != "" {
				expectedBasePath = scenario.basePath
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case fmt.Sprintf("/v1/%s/metadata/generic", expectedBasePath):
					fallthrough
				case fmt.Sprintf("/v1/%s/data/generic", expectedBasePath):
					fallthrough
				case fmt.Sprintf("/v1/%s/metadata/hello-finops", expectedBasePath):
					fallthrough
				case fmt.Sprintf("/v1/%s/data/hello-finops", expectedBasePath):
					w.WriteHeader(http.StatusOK)
				default:
					t.Errorf("Invalid request path: %s", r.URL.Path)
					t.Fail()
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			v, err := NewVaulterer(server.URL, func(*api.Client) error {
				return nil
			})
			assert.NoError(t, err)

			if scenario.basePath != "" {
				err = os.Setenv(VaultSecretBasePathEnv, scenario.basePath)
				assert.NoError(t, err)
			}

			resp := map[string]interface{}{}
			assert.NoError(t, v.Load("hello-finops", &resp))
		}()
	}
}
