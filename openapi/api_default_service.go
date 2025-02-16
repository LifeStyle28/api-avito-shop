package openapi

import (
	"api-avito-shop/engine"
	"api-avito-shop/models"
	"context"
)

// DefaultAPIService is a service that implements the logic for the DefaultAPIServicer
// This service should implement the business logic for every endpoint for the DefaultAPI API.
// Include any external packages or services that will be required by this service.
type DefaultAPIService struct {
	engine *engine.Engine
}

// NewDefaultAPIService creates a default api service
func NewDefaultAPIService(engine *engine.Engine) *DefaultAPIService {
	return &DefaultAPIService{
		engine: engine,
	}
}

// ApiInfoGet - Получить информацию о монетах, инвентаре и истории транзакций.
func (s *DefaultAPIService) ApiInfoGet(ctx context.Context) (models.ImplResponse, error) {
	return s.engine.HandleApiInfo(ctx)
}

// ApiSendCoinPost - Отправить монеты другому пользователю.
func (s *DefaultAPIService) ApiSendCoinPost(ctx context.Context, sendCoinRequest models.SendCoinRequest) (models.ImplResponse, error) {
	return s.engine.HandleApiSendCoin(ctx, sendCoinRequest)
}

// ApiBuyItemGet - Купить предмет за монеты.
func (s *DefaultAPIService) ApiBuyItemGet(ctx context.Context, item string) (models.ImplResponse, error) {
	return s.engine.HandleApiByuItem(ctx, item)
}

// ApiAuthPost - Аутентификация и получение JWT-токена. При первой аутентификации пользователь создается автоматически.
func (s *DefaultAPIService) ApiAuthPost(ctx context.Context, authRequest models.AuthRequest) (models.ImplResponse, error) {
	return s.engine.HandleApiAuth(ctx, authRequest)
}
