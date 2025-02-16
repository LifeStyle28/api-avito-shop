package main

import (
	"log"
	"net/http"

	"api-avito-shop/database"
	"api-avito-shop/engine"
	openapi "api-avito-shop/openapi"
)

func main() {
	log.Printf("Server started")

	db := database.Postgres{}
	e := engine.NewEngine(&db)
	DefaultAPIService := openapi.NewDefaultAPIService(e)
	DefaultAPIController := openapi.NewDefaultAPIController(DefaultAPIService)

	router := openapi.NewRouter(DefaultAPIController)

	log.Fatal(http.ListenAndServe(":8080", router))
}
