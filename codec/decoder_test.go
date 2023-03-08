package codec

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Decode_struct(t *testing.T) {
	sYAML := `
name: hoofddrop
age: 1000
typeVehicle: train`

	sJSON := `
{
	"name": "hoofddrop",
	"age": 1000,
	"typeVehicle": "train"
}`

	sTOML := `
name = "hoofddrop"
age = 1000
typeVehicle = "train"
`

	type args struct {
		sData map[string]string
		to    interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		match   interface{}
	}{
		{
			name: "Empty struct for strict mode",
			args: args{
				sData: map[string]string{
					"JSON": sJSON,
					"YAML": sYAML,
					"TOML": sTOML,
				},
				to: &struct{}{},
			},
			match:   &struct{}{},
			wantErr: false,
		},
		{
			name: "Full struct",
			args: args{
				sData: map[string]string{
					"JSON": sJSON,
					"YAML": sYAML,
					"TOML": sTOML,
				},
				to: &struct {
					Name        string `yaml:"name" json:"name"`
					Age         int    `yaml:"age" json:"age"`
					TypeVehicle string `yaml:"typeVehicle" json:"typeVehicle"`
				}{
					Name:        "undef",
					Age:         106,
					TypeVehicle: "undef",
				},
			},
			match: &struct {
				Name        string
				Age         int
				TypeVehicle string
			}{
				Name:        "hoofddrop",
				Age:         1000,
				TypeVehicle: "train",
			},
			wantErr: false,
		},
		{
			name: "Less struct",
			args: args{
				sData: map[string]string{
					"JSON": sJSON,
					"YAML": sYAML,
					"TOML": sTOML,
				},
				to: &struct {
					Name        string `yaml:"name" json:"name"`
					TypeVehicle string `yaml:"typeVehicle" json:"typeVehicle"`
				}{
					Name:        "undef",
					TypeVehicle: "undef",
				},
			},
			match: &struct {
				Name        string
				TypeVehicle string
			}{
				Name:        "hoofddrop",
				TypeVehicle: "train",
			},
			wantErr: false,
		},
		{
			name: "More struct",
			args: args{
				sData: map[string]string{
					"JSON": sJSON,
					"YAML": sYAML,
					"TOML": sTOML,
				},
				to: &struct {
					Name        string `yaml:"name" json:"name"`
					Age         int    `yaml:"age" json:"age"`
					TypeVehicle string `yaml:"typeVehicle" json:"typeVehicle"`
					SpeedLimit  int    `yaml:"speedLimit" json:"speedLimit"`
				}{
					Name:        "undef",
					Age:         0,
					TypeVehicle: "undef",
					SpeedLimit:  999,
				},
			},
			match: &struct {
				Name        string
				Age         int
				TypeVehicle string
				SpeedLimit  int
			}{
				Name:        "hoofddrop",
				Age:         1000,
				TypeVehicle: "train",
				SpeedLimit:  999,
			},
			wantErr: false,
		},
		{
			name: "Mix struct",
			args: args{
				sData: map[string]string{
					"JSON": sJSON,
					"YAML": sYAML,
					"TOML": sTOML,
				},
				to: &struct {
					Name        string `yaml:"name" json:"name"`
					TypeVehicle string `yaml:"typeVehicle" json:"typeVehicle"`
					SpeedLimit  int    `yaml:"speedLimit" json:"speedLimit"`
				}{
					Name:        "undef",
					TypeVehicle: "undef",
					SpeedLimit:  1000,
				},
			},
			match: &struct {
				Name        string
				TypeVehicle string
				SpeedLimit  int
			}{
				Name:        "hoofddrop",
				TypeVehicle: "train",
				SpeedLimit:  1000,
			},
			wantErr: false,
		},
	}

	decoders := map[string]Decoder{
		"YAML": YAML{},
		"JSON": JSON{},
		"TOML": TOML{},
	}

	for _, tt := range tests {
		for k, v := range decoders {
			t.Run(tt.name+" "+k, func(t *testing.T) {
				if err := v.Decode(strings.NewReader(tt.args.sData[k]), tt.args.to); (err != nil) != tt.wantErr {
					t.Errorf("%s.Decode() error = %v, wantErr %v", k, err, tt.wantErr)
				}

				if !assert.ObjectsAreEqualValues(tt.match, tt.args.to) {
					t.Errorf("%s.Decode() unmatch %+v to %+v", k, tt.match, tt.args.to)
				}
			})
		}
	}
}

func Test_Decode_map(t *testing.T) {
	sYAML := `
name: hoofddrop
age: 1000.0
typeVehicle: train`

	sJSON := `
{
	"name": "hoofddrop",
	"age": 1000.0,
	"typeVehicle": "train"
}`

	sTOML := `
name = "hoofddrop"
age = 1000.0
typeVehicle = "train"
`

	type args struct {
		sData map[string]string
		to    interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		match   interface{}
	}{
		{
			name: "Mapped",
			args: args{
				sData: map[string]string{
					"JSON": sJSON,
					"YAML": sYAML,
					"TOML": sTOML,
				},
				to: &map[string]interface{}{},
			},
			match: &map[string]interface{}{
				"name":        "hoofddrop",
				"age":         1000.0,
				"typeVehicle": "train",
			},
			wantErr: false,
		},
	}

	decoders := map[string]Decoder{
		"YAML": YAML{},
		"JSON": JSON{},
		"TOML": TOML{},
	}

	for _, tt := range tests {
		for k, v := range decoders {
			t.Run(tt.name+" "+k, func(t *testing.T) {
				if err := v.Decode(strings.NewReader(tt.args.sData[k]), tt.args.to); (err != nil) != tt.wantErr {
					t.Errorf("%s.Decode() error = %v, wantErr %v", k, err, tt.wantErr)
				}

				assert.EqualValues(t, tt.match, tt.args.to)
			})
		}
	}
}

func TestMapDecoder(t *testing.T) {
	type inner struct {
		Field2 string `secret:"field_2"`
	}

	type testStruct struct {
		Field1   string  `secret:"field_1"`
		Value    float64 `secret:"value"`
		ValueInt int     `secret:"valueInt"`
		Untagged int64
		NoSet    string `secret:"-"`
		NoData   string `secret:"missing"`
		Time     time.Time
		Duration time.Duration `cfg:"duration"`
		Inner    inner         `secret:"other"`
	}

	type innerDef struct {
		Field2 string `cfg:"field_2"`
	}

	type testStructDef struct {
		Field1   string  `cfg:"field_1"`
		Value    float64 `secret:"value"`
		ValueInt int     `secret:"valueInt"`
		Untagged int64
		NoSet    string `secret:"-"`
		NoData   string `cfg:"missing"`
		Time     time.Time
		Inner    innerDef `secret:"other"`
	}

	type args struct {
		input  interface{}
		output interface{}
		tag    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    interface{}
	}{
		{
			args: args{
				input: &map[string]interface{}{
					"field_1":  "one",
					"value":    64,
					"valueInt": 64,
					"noset":    "not_empty",
					"other": map[string]interface{}{
						"field_2": "other",
					},
					"untagged": 54,
					"duration": "2m",
				},
				output: &testStruct{},
				tag:    "secret",
			},
			wantErr: false,
			want: &testStruct{
				Field1:   "one",
				Value:    64,
				ValueInt: 64,
				Untagged: 54,
				Inner: inner{
					Field2: "other",
				},
				Duration: 2 * time.Minute,
			},
		},
		{
			args: args{
				input: &map[string]interface{}{
					"field_1":  "one",
					"value":    64,
					"valueInt": 64,
					"noset":    "not_empty",
					"other": map[string]interface{}{
						"field_2": "other",
					},
					"untagged": 54,
					"missing":  "diff",
				},
				output: &testStructDef{},
				tag:    "secret",
			},
			wantErr: false,
			want: &testStructDef{
				Field1:   "one",
				Value:    64,
				ValueInt: 64,
				Untagged: 54,
				Inner: innerDef{
					Field2: "other",
				},
				NoData: "diff",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, MapDecoder(tt.args.input, tt.args.output, tt.args.tag))
			assert.Equal(t, tt.want, tt.args.output)
		})
	}
}

func TestMapDecoder_Nil(t *testing.T) {
	type inner struct {
		Field2 string `secret:"field_2"`
	}

	type testStruct struct {
		Field1   string  `secret:"field_1"`
		Value    float64 `secret:"value"`
		ValueInt int     `secret:"valueInt"`
		Untagged int64
		NoSet    string `secret:"-"`
		NoData   string `secret:"missing"`
		Time     time.Time
		Inner    inner `secret:"other"`
	}

	type args struct {
		input  interface{}
		output interface{}
		tag    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    interface{}
	}{
		{
			args: args{
				input: nil,
				output: &testStruct{
					Field1: "don't change this",
					Value:  1234,
				},
				tag: "secret",
			},
			wantErr: false,
			want: &testStruct{
				Field1:   "don't change this",
				Value:    1234,
				ValueInt: 0,
				Untagged: 0,
				Inner: inner{
					Field2: "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, MapDecoder(tt.args.input, tt.args.output, tt.args.tag))
			assert.Equal(t, tt.want, tt.args.output)
		})
	}
}

func TestLoadReaderWithDecoder(t *testing.T) {
	type testStructWithCfgTag struct {
		UntaggedStr string
		ValueInt    int      `cfg:"valueInt"`
		ValueFloat  float64  `cfg:"valueFloat"`
		ValueStr    string   `cfg:"valueStr"`
		ValueSlice  []string `cfg:"valueArr"`
		IgnoreStr   string   `cfg:"-"`
	}

	tests := map[string]struct {
		input   string
		output  interface{}
		decoder Decoder
		tag     string
		want    interface{}
	}{
		"test with json decoder": {
			input: `
{
	"untaggedStr": "untagged string",
	"valueInt": 64,
	"valueFloat": 6.4,
	"valueStr": "value string",
	"valueArr": ["one", "two"],
	"ignoreStr": "ignore string"
}`,
			output:  &testStructWithCfgTag{},
			decoder: JSON{},
			tag:     "cfg",
			want: &testStructWithCfgTag{
				UntaggedStr: "untagged string",
				ValueInt:    64,
				ValueFloat:  6.4,
				ValueSlice:  []string{"one", "two"},
				ValueStr:    "value string",
			},
		},
		"test with yaml decoder": {
			input: `untaggedStr: "untagged string"
valueInt: 64
valueFloat: 6.4
valueStr: "value string"
valueArr: 
  - one
  - two
ignoreStr: 'ignore string'
`,
			output:  &testStructWithCfgTag{},
			decoder: YAML{},
			tag:     "cfg",
			want: &testStructWithCfgTag{
				UntaggedStr: "untagged string",
				ValueInt:    64,
				ValueFloat:  6.4,
				ValueSlice:  []string{"one", "two"},
				ValueStr:    "value string",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NoError(t, LoadReaderWithDecoder(strings.NewReader(tc.input), tc.output, tc.decoder, tc.tag))
			assert.Equal(t, tc.want, tc.output)
		})
	}
}
