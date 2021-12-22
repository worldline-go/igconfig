package loader

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/codec"
)

func TestConsul_DynamicValue(t *testing.T) {
	type fields struct {
		consulMock *ConsulMock
		Decoder    codec.Decoder
		Plan       Planer
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      []byte
		wantErr   bool
		changeVal [][]byte
	}{
		{
			name: "testing",
			fields: fields{
				consulMock: &ConsulMock{kv: map[string][]byte{
					"test": []byte(`some values`),
				}},
			},
			args: args{
				ctx: context.Background(),
				key: "test",
			},
			want:      []byte("some values"),
			wantErr:   false,
			changeVal: [][]byte{[]byte(`some values 2`), []byte(`other value`)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consulMock := tt.fields.consulMock
			if consulMock == nil {
				consulMock = &ConsulMock{kv: make(map[string][]byte)}
			}

			client := NewConsulMock(consulMock)

			c := Consul{
				Client:  client,
				Decoder: tt.fields.Decoder,
				Plan:    tt.fields.Plan,
			}

			ctx, cancel := context.WithCancel(tt.args.ctx)
			defer cancel()

			got, err := c.DynamicValue(ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Consul.DynamicValue() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(<-got, tt.want) {
				t.Errorf("Consul.DynamicValue() = %v, want %v", got, tt.want)
			}

			// change
			for i := range tt.changeVal {
				tt.fields.consulMock.SetKey(tt.args.key, tt.changeVal[i])

				v := <-got
				fmt.Printf("%s\n", tt.changeVal[i])
				fmt.Printf("%s\n", v)
				if !reflect.DeepEqual(v, tt.changeVal[i]) {
					t.Errorf("got = %s, want %s", v, tt.changeVal[i])
				}
			}
		})
	}
}
