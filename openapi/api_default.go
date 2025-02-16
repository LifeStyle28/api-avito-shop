package openapi

import (
	"api-avito-shop/models"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// DefaultAPIController binds http requests to an api service and writes the service results to the http response
type DefaultAPIController struct {
	service      DefaultAPIServicer
	errorHandler ErrorHandler
}

// DefaultAPIOption for how the controller is set up.
type DefaultAPIOption func(*DefaultAPIController)

// WithDefaultAPIErrorHandler inject ErrorHandler into controller
func WithDefaultAPIErrorHandler(h ErrorHandler) DefaultAPIOption {
	return func(c *DefaultAPIController) {
		c.errorHandler = h
	}
}

// NewDefaultAPIController creates a default api controller
func NewDefaultAPIController(s DefaultAPIServicer, opts ...DefaultAPIOption) *DefaultAPIController {
	controller := &DefaultAPIController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the DefaultAPIController
func (c *DefaultAPIController) Routes() Routes {
	return Routes{
		"ApiInfoGet": Route{
			strings.ToUpper("Get"),
			"/api/info",
			c.ApiInfoGet,
			true,
		},
		"ApiSendCoinPost": Route{
			strings.ToUpper("Post"),
			"/api/sendCoin",
			c.ApiSendCoinPost,
			true,
		},
		"ApiBuyItemGet": Route{
			strings.ToUpper("Get"),
			"/api/buy/{item}",
			c.ApiBuyItemGet,
			true,
		},
		"ApiAuthPost": Route{
			strings.ToUpper("Post"),
			"/api/auth",
			c.ApiAuthPost,
			false,
		},
	}
}

// ApiInfoGet - Получить информацию о монетах, инвентаре и истории транзакций.
func (c *DefaultAPIController) ApiInfoGet(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.ApiInfoGet(r.Context())
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

// ApiSendCoinPost - Отправить монеты другому пользователю.
func (c *DefaultAPIController) ApiSendCoinPost(w http.ResponseWriter, r *http.Request) {
	var sendCoinRequestParam models.SendCoinRequest
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&sendCoinRequestParam); err != nil {
		c.errorHandler(w, r, &models.ParsingError{Err: err}, nil)
		return
	}
	if err := models.AssertSendCoinRequestRequired(sendCoinRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := models.AssertSendCoinRequestConstraints(sendCoinRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ApiSendCoinPost(r.Context(), sendCoinRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

// ApiBuyItemGet - Купить предмет за монеты.
func (c *DefaultAPIController) ApiBuyItemGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	itemParam := params["item"]
	if itemParam == "" {
		c.errorHandler(w, r, &models.RequiredError{Field: "item"}, nil)
		return
	}
	result, err := c.service.ApiBuyItemGet(r.Context(), itemParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

// ApiAuthPost - Аутентификация и получение JWT-токена. При первой аутентификации пользователь создается автоматически.
func (c *DefaultAPIController) ApiAuthPost(w http.ResponseWriter, r *http.Request) {
	var authRequestParam models.AuthRequest
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&authRequestParam); err != nil {
		c.errorHandler(w, r, &models.ParsingError{Err: err}, nil)
		return
	}
	if err := models.AssertAuthRequestRequired(authRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := models.AssertAuthRequestConstraints(authRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ApiAuthPost(r.Context(), authRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}
