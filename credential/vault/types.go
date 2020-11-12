package vault

type DatabaseCred struct {
	Username string
	Password string
}

func ForDatabase(data map[string]interface{}) DatabaseCred {
	return DatabaseCred{
		Username: data["username"].(string),
		Password: data["password"].(string),
	}
}
