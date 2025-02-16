package models

type AuthRequest struct {

	// Имя пользователя для аутентификации.
	Username string `json:"username"`

	// Пароль для аутентификации.
	Password string `json:"password"`
}

// AssertAuthRequestRequired checks if the required fields are not zero-ed
func AssertAuthRequestRequired(obj AuthRequest) error {
	elements := map[string]interface{}{
		"username": obj.Username,
		"password": obj.Password,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertAuthRequestConstraints checks if the values respects the defined constraints
func AssertAuthRequestConstraints(obj AuthRequest) error {
	return nil
}
