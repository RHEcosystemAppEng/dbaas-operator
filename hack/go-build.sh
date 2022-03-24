#!/bin/bash

set -e

OUTDIR_BIN="build/_output/bin"

mkdir -p ${OUTDIR_BIN}

go build -o ${OUTDIR_BIN}/metrics-exporter ./metrics/main.go