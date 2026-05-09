package loader

import (
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// gcpResourceName converts a name that may contain "/" (e.g. Consul/Vault-style
// paths like "finops/br_sr") into a GCP-safe resource name by replacing "/" with "-".
func gcpResourceName(name string) string {
	return strings.ReplaceAll(name, "/", "-")
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
