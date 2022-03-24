#!/bin/bash

set -e

docker build -f build/Dockerfile -t "quay.io/rchikatw/dbaas-metrics-exporter:v$1" build/
docker push quay.io/rchikatw/dbaas-metrics-exporter:v$1
