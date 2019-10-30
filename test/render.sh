#!/bin/bash

# Env variables (defaults in []):
#   VSP_DEV_ID:       Used in the "dev-id" label to separate
#                     multiple dev deployments int the same namespace ["dev"]
#   VSP_DEV_REGISTRY: The registry hostname (including a trailing "/") to
#                     prefix each image with [""]
#   VSP_DEV_TOKEN:    The Vault root token ["123456790"]

if [ -f "$(dirname $0)/env" ]; then
  source "$(dirname $0)/env"
fi

echo "---"
spruce merge vault-statefulset.yml
echo "---"
spruce merge vault-service.yml
echo "---"
spruce merge vault-configmap.yml