#!/bin/bash

# Env variables (defaults in []):
#   VSB_DEV_ID:           Used in the "dev-id" label to separate
#                         multiple dev deployments int the same namespace ["dev"]
#   VSB_DEV_REGISTRY:     The registry hostname (including a trailing "/") to
#                         prefix each image with [""]
#   VSB_DEV_TOKEN:        The Vault root token ["123456790"]
#   VSB_DEV_IMAGE:        The Vault secret broker image ["bedag/vault-secret-broker:latest"]
#   VSB_DEV_STORAGECLASS: Storage class to be used [""]

if [ -f "$(dirname $0)/env" ]; then
  source "$(dirname $0)/env"
fi

echo "---"
spruce merge vault-statefulset.yml
echo "---"
spruce merge vault-service.yml
echo "---"
spruce merge vault-configmap.yml
echo "---"
spruce merge vault-secret-broker-init-configmap.yml
echo "---"
spruce merge vault-secret-broker-pvc.yml
echo "---"
spruce merge vault-secret-broker-statefulset.yml