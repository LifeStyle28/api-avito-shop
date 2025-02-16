#!/bin/bash

if [ "$MODE" = "test" ]; then
    . ${GOPATH}/venv/bin/activate
    pytest -v
fi
