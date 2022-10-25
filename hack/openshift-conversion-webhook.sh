#!/usr/bin/env bash

set -o errexit
set -o pipefail

conversion_webhook_crds=$(yq 'select(.spec | has("conversion")) | filename' bundle/manifests/dbaas.redhat.com_*.yaml | grep -v -e "---")
for fname in $conversion_webhook_crds; do
    yq -i eval 'del(.spec.conversion)' "$fname"
done

rm bundle/manifests/*webhook-service*service.yaml