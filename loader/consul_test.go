package loader

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/worldline-go/igconfig/codec"
	"github.com/worldline-go/igconfig/internal"
)

func TestLoadFromConsul(t *testing.T) {
	type inner struct {
		Slice []string `yaml:"slice"`
	}
	type res struct {
		UntaggedStr     string
		CamelCaseStr    string `cfg:"camelCaseStr"`
		CamelCaseInt    int64  `cfg:"camelCaseInt"`
		CamelCaseStruct inner  `cfg:"camelCaseStruct"`
		SnakeCaseInt    int64  `cfg:"snake_case_int"`
		SnakeCaseStruct inner  `cfg:"snake_case_struct"`
	}

	tests := []struct {
		name       string
		consulConf *ConsulMock
		to         res
		result     res
		err        string
	}{
		{
			name: "test-json",
			consulConf: &ConsulMock{kv: map[string][]byte{
				"test-json": []byte(`{untaggedStr: 'untag value', camelCaseStr: 'camel case value', camelCaseInt: 64, camelCaseStruct: {slice: [one, two]}, snake_case_int: 55, snake_case_struct: {slice: [one, two, three four]}}`),
			}},
			result: res{
				UntaggedStr:     "untag value",
				CamelCaseStr:    "camel case value",
				CamelCaseInt:    64,
				CamelCaseStruct: inner{[]string{"one", "two"}},
				SnakeCaseInt:    55,
				SnakeCaseStruct: inner{[]string{"one", "two", "three four"}},
			},
		},
		{
			name: "test-yaml",
			consulConf: &ConsulMock{kv: map[string][]byte{
				"test-yaml": []byte(`
untaggedStr: test
camelCaseStr: 'camel case value'
camelCaseInt: 64
camelCaseStruct:
  slice:
  - one
  - two
snake_case_int: 55
snake_case_struct:
  slice:
  - one
  - two
  - three four`),
			}},
			result: res{
				UntaggedStr:     "test",
				CamelCaseStr:    "camel case value",
				CamelCaseInt:    64,
				CamelCaseStruct: inner{[]string{"one", "two"}},
				SnakeCaseInt:    55,
				SnakeCaseStruct: inner{[]string{"one", "two", "three four"}},
			},
		},
		{
			name:       "no-key",
			consulConf: &ConsulMock{kv: map[string][]byte{}},
		},
		{
			name:       "error",
			consulConf: &ConsulMock{err: errors.New("test error")},
			err:        "test error",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Consul{Client: NewConsulMock(test.consulConf)}.Load(test.name, &test.to)

			if test.err == "" {
				assert.NoError(t, err)
			} else {
				// Errors from Consul client RoundTripper always wrapped with url.Error
				assert.EqualError(t, errors.Unwrap(err), test.err)
			}

			assert.Equal(t, test.result, test.to)
		})
	}
}

func TestNewConsuler_WrongAddr(t *testing.T) {
	c, err := NewConsul("locall:8787")

	assert.Nil(t, err)
	assert.NotNil(t, c)
}

func TestConsul_PathPrefix(t *testing.T) {
	tests := []struct {
		name       string
		pathPrefix string
	}{
		{
			name: "get_values_on_default_path",
		},
		{
			name:       "get_values_on_custom_path",
			pathPrefix: "finops/kv2",
		},
	}
	for _, scenario := range tests {
		func() {
			expectedBasePath := ConsulConfigPathPrefix
			if scenario.pathPrefix != "" {
				expectedBasePath = scenario.pathPrefix
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == fmt.Sprintf("/v1/kv/%s/hello-finops", expectedBasePath) {
					kv := []*api.KVPair{
						{
							Key:   "key",
							Value: []byte(`{"value":"test"}`),
						},
					}
					resp, err := json.Marshal(kv)
					assert.NoError(t, err)

					_, err = w.Write(resp)
					assert.NoError(t, err)
					return
				}

				t.Errorf("Invalid request path: %s", r.URL.Path)
				t.Fail()
				w.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			cl, err := NewConsul(server.URL)
			assert.NoError(t, err)

			loader := Consul{Client: cl}
			loader.Decoder = codec.JSON{}

			if scenario.pathPrefix != "" {
				err = os.Setenv(ConsulConfigPathPrefixEnv, scenario.pathPrefix)
				assert.NoError(t, err)
			}

			resp := map[string]interface{}{}
			assert.NoError(t, loader.Load("hello-finops", &resp))
		}()
	}
}

func NewConsulMock(mockConfig *ConsulMock) *api.Client {
	cl, _ := api.NewClient(&api.Config{
		HttpClient: &http.Client{
			Transport: mockConfig,
		},
	})

	return cl
}

type ConsulMock struct {
	kvFunc    func(keyPath string) (*api.KVPair, *api.QueryMeta, bool)
	kv        map[string][]byte
	lock      sync.RWMutex
	LastIndex uint64
	err       error
}

func (m *ConsulMock) RoundTrip(request *http.Request) (*http.Response, error) {
	reqURI := request.URL.RequestURI()

	switch {
	case strings.HasPrefix(reqURI, "/v1/kv/"):
		key := strings.TrimPrefix(reqURI, path.Join("/v1/kv",
			internal.GetEnvWithFallback(ConsulConfigPathPrefixEnv, ConsulConfigPathPrefix))+"/")

		kvResp, meta, err := m.Get(key, nil)

		bts, _ := json.Marshal(api.KVPairs{kvResp})

		httpResp := http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(bts)),
		}

		if kvResp == nil {
			httpResp.StatusCode = http.StatusNotFound
		}

		httpResp.Header = generateMetaHeader(meta)

		return &httpResp, err
	}

	return nil, fmt.Errorf("%s %s", request.Method, request.URL.RequestURI())
}

func (m *ConsulMock) Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error) {
	if m.err != nil {
		return nil, nil, m.err
	}

	var data = &api.KVPair{
		Key: key,
	}

	m.lock.RLock()
	var meta = &api.QueryMeta{
		LastIndex: m.LastIndex,
	}
	m.lock.RUnlock()

	var ok bool

	if inx := strings.Index(key, "?index"); inx != -1 {
		key = key[:inx]
	}

	data.Value, ok = m.GetKey(key)
	if !ok {
		if m.kvFunc == nil {
			return nil, nil, nil
		}

		data, meta, _ = m.kvFunc(key)
	}

	return data, meta, nil
}

func (m *ConsulMock) SetKey(key string, value []byte) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.kv[key] = value
	m.LastIndex++
}

func (m *ConsulMock) GetKey(key string) (val []byte, ok bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	val, ok = m.kv[key]

	return
}

func generateMetaHeader(meta *api.QueryMeta) http.Header {
	var h = http.Header{}

	if meta == nil {
		return h
	}

	h.Set("X-Consul-Index", strconv.FormatUint(meta.LastIndex, 10))

	return h
}
