package vault

// DatabaseCred is a credential for database
type DatabaseCred struct {
	Username string
	Password string
}

// ForDatabase map data to DatabaseCred.
func ForDatabase(data map[string]interface{}) DatabaseCred {
	return DatabaseCred{
		Username: data["username"].(string),
		Password: data["password"].(string),
	}
}
