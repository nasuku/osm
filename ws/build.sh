#!/bin/bash
#source .env
export CTR_REGISTRY=docker.dev.ws:5000
export CTR_TAG=test
make docker-push-osm-controller
make docker-push-init
#make build-osm
