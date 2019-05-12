#!/usr/bin/env

export GOOS=linux


ring-pwa:
	go build -o ring-pwa 

.PHONY: ring-pwa