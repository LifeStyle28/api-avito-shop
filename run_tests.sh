#!/bin/bash

# прогоним юнит-тесты
cd engine && go test --cover && cd -

# прогоним интеграционные тесты
export MODE=test
docker-compose up --build -d
docker wait $(docker-compose ps -q tests)
docker-compose logs tests
docker-compose down
unset MODE
