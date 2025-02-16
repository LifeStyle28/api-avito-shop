package models

type ErrorResponse struct {

	// Сообщение об ошибке, описывающее проблему.
	Errors string `json:"errors,omitempty"`
}

// AssertErrorResponseRequired checks if the required fields are not zero-ed
func AssertErrorResponseRequired(obj ErrorResponse) error {
	return nil
}

// AssertErrorResponseConstraints checks if the values respects the defined constraints
func AssertErrorResponseConstraints(obj ErrorResponse) error {
	return nil
}
