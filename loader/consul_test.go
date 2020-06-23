package loader

import (
	"errors"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

func TestLoadFromConsul(t *testing.T) {
	type inner struct {
		Slice []string `yaml:"slice"`
	}
	type res struct {
		FirstName string
		Base      int64 `yaml:"base_int"`
		Inner     inner `yaml:"inner"`
	}

	tests := []struct {
		name     string
		consuler Consuler
		to       res
		result   res
		err      error
	}{
		{
			name: "test",
			consuler: ConsulMock{kv: map[string][]byte{
				"config/test": []byte(`{firstname: test, base_int: 55, inner: {slice: [one, two, three four]}}`),
			}},
			result: res{
				FirstName: "test",
				Base:      55,
				Inner:     inner{[]string{"one", "two", "three four"}},
			},
		},
		{
			name:     "no-key",
			consuler: ConsulMock{kv: map[string][]byte{}},
		},
		{
			name:     "error",
			consuler: ConsulMock{err: errors.New("test error")},
			err:      errors.New("test error"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			err := Consul{Client: test.consuler}.Load(test.name, &test.to)

			if test.err == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, test.err, err)
			}

			assert.Equal(t, test.result, test.to)
		})
	}
}

func TestNewConsuler_WrongAddr(t *testing.T) {
	c, err := NewConsuler("locall:8787")

	assert.Nil(t, err)
	assert.NotNil(t, c)
}

type ConsulMock struct {
	kv  map[string][]byte
	err error
}

func (m ConsulMock) Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error) {
	if m.err != nil {
		return nil, nil, m.err
	}

	val, ok := m.kv[key]
	if !ok {
		return nil, nil, nil
	}

	data := api.KVPair{
		Key:   key,
		Value: val,
	}

	return &data, nil, nil
}
