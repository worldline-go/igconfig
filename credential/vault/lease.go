package vault

import (
	"context"
	pathpkg "path"

	"github.com/hashicorp/vault/api"
)

//nolint:golint
type LogicReader func(string) (*api.Secret, error)

// Database returns usable path to get database lease path for specified role.
func Database(role string) string {
	return pathpkg.Join("database/creds", role)
}

// GetCredentials obtains(reads) secret from Vault.
//
// This could be used to get normal credentials and also lease them.
// If secret should be renewed - please use KeepRenewed function.
//
// 'path' is Vaults path that should return secret when it is read.
// For example any secret engine.
//
// Argument 'path' can be constructed with provided functions, for example Database.
//
// Example:
//	s, err := GetCredentials(Database("test_app_recon"))
//
//	// Secret 's' is already usable, but
//	// if needs to be renewed(for example when it is leased database credentials) - use this.
//	// Cancel this context when renewal should be stopped.
//	ctx, cancel := context.WithCancel(context.Background())
//
//	// After renewal is stopped secret will be valid for no more than it's TTL time.
//	go KeepRenewed(ctx, cl, s) // 'cl' is *vault/api.Client
//
//	// Now secret 's' will be valid for as long as context is not canceled.
func GetCredentials(path string) (*api.Secret, error) {
	cl, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}

	return GetCredentialsWithClient(cl, path)
}

// GetCredentialsWithClient see GetCredentials for path description.
//
// This function starts goroutine to renew secret. This is non-optional, but controlled by provided context.
// Renew will stop when context will be canceled.
// If context is canceled before renew would be started(for example canceled context was passed in this function) -
// renew will not kick in. Secret will be retrieved and returned without starting renewal process.
func GetCredentialsWithClient(cl *api.Client, path string) (*api.Secret, error) {
	return GetCredentialsWithReader(cl.Logical().Read, path)
}

// GetCredentialsWithReader uses provided reader to retrieve leased secret from specified path.
//
// In general case use GetCredentials.
func GetCredentialsWithReader(reader LogicReader, path string) (*api.Secret, error) {
	return reader(path)
}

// KeepRenewed will keep secret valid until context will be canceled or error happens.
//
// This is blocking function. But it does spawn goroutine that will keep secret updated.
// This means that developer needs only to handle error cases.
//
// If secret is not renewable - it will immediately exit.
//
// If function is stopped gracefully(eg context canceled) - no error is returned.
func KeepRenewed(ctx context.Context, cl *api.Client, secret *api.Secret) error {
	renewable, err := secret.TokenIsRenewable()
	if err != nil || renewable { // Do not start if secret is not renewable
		return err
	}

	select {
	case <-ctx.Done(): // Do not run if context already been canceled
		return nil
	default:
	}

	renewer, err := cl.NewLifetimeWatcher(&api.RenewerInput{
		Secret: secret,
	})
	if err != nil {
		return err
	}

	go renewer.Renew()

	for {
		select {
		case <-ctx.Done():
			renewer.Stop()
		case err := <-renewer.DoneCh():
			return err
		case <-renewer.RenewCh():
		}
	}
}
