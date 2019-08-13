package igconfig

// loadDefaults loads the config struct fields with their default value as defined in the tags
func (m *localData) loadDefaults() {
	t := m.userStruct.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if v, ok := field.Tag.Lookup("default"); ok {
			m.setValue(field.Name, v)
		}
	}
}
