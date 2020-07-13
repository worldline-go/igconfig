package igconfig

// loadDefaults loads the config struct fields with their default value as defined in the tags
func (m *localData) loadDefaults() {
	t := m.userStruct.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if m.userStruct.Field(i).IsValid() && !m.userStruct.Field(i).IsZero() {
			// Value is already set, default should not update it
			continue
		}

		if v, ok := field.Tag.Lookup("default"); ok {
			m.setValue(field.Name, v)
		}
	}
}
