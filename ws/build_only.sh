#!/bin/bash

#ws docker registry push
export CTR_REGISTRY=docker.dev.ws:5000
export CTR_TAG=testosm1
make docker-push-osm-controller
make docker-push-init
