package loader

import (
	"errors"
	"reflect"
	"testing"
)

func TestParameterManagerConfig(t *testing.T) {
	tests := []struct {
		name    string
		loader  ParameterManager
		env     map[string]string
		want    parameterManagerConfig
		wantErr error
	}{
		{
			name:    "no project",
			wantErr: ErrNoClient,
		},
		{
			name:   "defaults",
			loader: ParameterManager{ProjectID: "field"},
			want:   parameterManagerConfig{ProjectID: "field", Location: "global", Version: "latest", Fallback: true},
		},
		{
			name:   "fields",
			loader: ParameterManager{ProjectID: "field", Location: "europe-west3", Version: "v1", DisableVersionFallback: true},
			want:   parameterManagerConfig{ProjectID: "field", Location: "europe-west3", Version: "v1", Fallback: false},
		},
		{
			name:   "env priority",
			loader: ParameterManager{ProjectID: "field", Location: "europe-west3", Version: "v1", DisableVersionFallback: true},
			env: map[string]string{
				"GCP_PROJECT_ID":         "env",
				"GCP_PARAMETER_LOCATION": "europe-west4",
				"GCP_PARAMETER_VERSION":  "v2",
				"GCP_PARAMETER_FALLBACK": "true",
			},
			want: parameterManagerConfig{ProjectID: "env", Location: "europe-west4", Version: "v2", Fallback: true},
		},
		{
			name:    "invalid fallback",
			env:     map[string]string{"GCP_PROJECT_ID": "env", "GCP_PARAMETER_FALLBACK": "nope"},
			wantErr: errors.New("invalid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range []string{"GCP_PROJECT_ID", "GCP_PARAMETER_LOCATION", "GCP_PARAMETER_VERSION", "GCP_PARAMETER_FALLBACK"} {
				t.Setenv(k, tt.env[k])
			}

			got, err := tt.loader.config()
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestSecretManagerConfig(t *testing.T) {
	tests := []struct {
		name   string
		loader SecretManager
		env    map[string]string
		want   secretManagerConfig
	}{
		{
			name:   "defaults",
			loader: SecretManager{ProjectID: "field"},
			want:   secretManagerConfig{ProjectID: "field", Location: "global", Version: "latest", Additional: []string{"generic"}},
		},
		{
			name:   "fields",
			loader: SecretManager{ProjectID: "field", Location: "europe-west3", Version: "3", AdditionalSecrets: []string{}},
			want:   secretManagerConfig{ProjectID: "field", Location: "europe-west3", Version: "3", Additional: []string{}},
		},
		{
			name:   "env priority",
			loader: SecretManager{ProjectID: "field", Location: "europe-west3", AdditionalSecrets: []string{"x"}},
			env: map[string]string{
				"GCP_PROJECT_ID":        "env",
				"GCP_SECRET_LOCATION":   "europe-west4",
				"GCP_SECRET_VERSION":    "5",
				"GCP_SECRET_ADDITIONAL": " generic , keycloak,,",
			},
			want: secretManagerConfig{ProjectID: "env", Location: "europe-west4", Version: "5", Additional: []string{"generic", "keycloak"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range []string{"GCP_PROJECT_ID", "GCP_SECRET_LOCATION", "GCP_SECRET_VERSION", "GCP_SECRET_ADDITIONAL"} {
				t.Setenv(k, tt.env[k])
			}

			got, err := tt.loader.config()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestSecretResourceName(t *testing.T) {
	cfg := secretManagerConfig{ProjectID: "p", Location: "global", Version: "latest"}
	if got := secretResourceName(cfg, gcpResourceName("/finops/br_sr/")); got != "projects/p/secrets/finops-br_sr/versions/latest" {
		t.Errorf("unexpected global name: %s", got)
	}

	cfg.Location = "europe-west3"
	if got := secretResourceName(cfg, "app"); got != "projects/p/locations/europe-west3/secrets/app/versions/latest" {
		t.Errorf("unexpected regional name: %s", got)
	}
}
