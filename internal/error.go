package internal

import (
	"errors"
	"net"
	"os"
	"strings"
	"syscall"
)

// IsLocalNetworkError will check if provided error is local connection error.
//
// It will return true if error is connection failed to host "127.0.0.1".
func IsLocalNetworkError(err error) bool {
	// Check if error is network one.
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		// Check if network error is well-known "Connection Refused"
		var sErr *os.SyscallError
		if !errors.As(err, &sErr) || !errors.Is(sErr.Err, syscall.ECONNREFUSED) {
			// If it is not a connection refused - return it.
			return false
		}

		// If host is 127.0.0.1 - it means that no hostname was provided in environment.
		// Please use "localhost" if you want to receive an error instead.
		if strings.HasPrefix(netErr.Addr.String(), "127.0.0.1") {
			return true
		}
	}

	return false
}
