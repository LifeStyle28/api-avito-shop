package models

type AuthResponse struct {

	// JWT-токен для доступа к защищенным ресурсам.
	Token string `json:"token,omitempty"`
}

// AssertAuthResponseRequired checks if the required fields are not zero-ed
func AssertAuthResponseRequired(obj AuthResponse) error {
	return nil
}

// AssertAuthResponseConstraints checks if the values respects the defined constraints
func AssertAuthResponseConstraints(obj AuthResponse) error {
	return nil
}
