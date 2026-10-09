package loader

import (
	"os"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GCPDefaultLocation is used when no location is configured.
const GCPDefaultLocation = "global"

// GCPDefaultVersion is used when no version is configured.
const GCPDefaultVersion = "latest"

// gcpResourceName converts a name that may contain "/" (e.g. Consul/Vault-style
// paths like "finops/br_sr") into a GCP-safe resource name by replacing "/" with "-".
// Leading and trailing "/" are trimmed.
func gcpResourceName(name string) string {
	return strings.ReplaceAll(strings.Trim(name, "/"), "/", "-")
}

// envOr returns the value of the environment variable if it is set and not empty,
// otherwise the first non-empty value of fallbacks.
func envOr(env string, fallbacks ...string) string {
	if v := os.Getenv(env); v != "" {
		return v
	}

	for _, v := range fallbacks {
		if v != "" {
			return v
		}
	}

	return ""
}

// splitComma parses a comma-separated list, trimming whitespace and dropping empty entries.
func splitComma(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))

	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}

	return out
}

// isGCPNotFound returns true when a GCP API call returns a NotFound status.
// Used by ParameterManager and SecretManager loaders to distinguish "resource does not exist"
// (silent skip) from other errors (hard failure).
func isGCPNotFound(err error) bool {
	if err == nil {
		return false
	}

	s, ok := status.FromError(err)

	return ok && s.Code() == codes.NotFound
}

// isGCPFailedPrecondition returns true when a GCP API call fails with FailedPrecondition.
// Parameter Manager returns this when the requested version is disabled.
func isGCPFailedPrecondition(err error) bool {
	if err == nil {
		return false
	}

	s, ok := status.FromError(err)

	return ok && s.Code() == codes.FailedPrecondition
}
