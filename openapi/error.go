package openapi

import (
	"api-avito-shop/models"
	"errors"
	"net/http"
)

// ErrorHandler defines the required method for handling error. You may implement it and inject this into a controller if
// you would like errors to be handled differently from the DefaultErrorHandler
type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error, result *models.ImplResponse)

// DefaultErrorHandler defines the default logic on how to handle errors from the controller. Any errors from parsing
// request params will return a StatusBadRequest. Otherwise, the error code originating from the servicer will be used.
func DefaultErrorHandler(w http.ResponseWriter, _ *http.Request, err error, result *models.ImplResponse) {
	var parsingErr *models.ParsingError
	if ok := errors.As(err, &parsingErr); ok {
		// Handle parsing errors
		_ = EncodeJSONResponse(err.Error(), func(i int) *int { return &i }(http.StatusBadRequest), w)
		return
	}

	var requiredErr *models.RequiredError
	if ok := errors.As(err, &requiredErr); ok {
		// Handle missing required errors
		_ = EncodeJSONResponse(err.Error(), func(i int) *int { return &i }(http.StatusUnprocessableEntity), w)
		return
	}

	// Handle all other errors
	_ = EncodeJSONResponse(err.Error(), &result.Code, w)
}
