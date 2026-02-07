#!/bin/bash

unset location
unset resourceGroup
unset cognitiveService
unset cognitiveServiceSku
unset cognitiveServiceKind
unset cognitiveServiceDomain
unset cognitiveServiceRole
unset modelDeployment
unset modelName
unset modelVersion
unset modelFormat
unset modelSku
unset modelSkuCapacity

POSITIONAL_ARGS=()

while [[ $# -gt 0 ]]; do
  case $1 in
  --location)
    location="$2"
    shift
    shift
    ;;
  --resource-group)
    resourceGroup="$2"
    shift
    shift
    ;;
  --cognitive-service)
    cognitiveService="$2"
    shift
    shift
    ;;
  --cognitive-services-sku)
    cognitiveServiceSku="$2"
    shift
    shift
    ;;
  --cognitive-service-kind)
    cognitiveServiceKind="$2"
    shift
    shift
    ;;
  --cognitive-service-domain)
    cognitiveServiceDomain="$2"
    shift
    shift
    ;;
  --cognitive-service-role)
    cognitiveServiceRole="$2"
    shift
    shift
    ;;
  --model-deployment)
    modelDeployment="$2"
    shift
    shift
    ;;
  --model-name)
    modelName="$2"
    shift
    shift
    ;;
  --model-version)
    modelVersion="$2"
    shift
    shift
    ;;
  --model-format)
    modelFormat="$2"
    shift
    shift
    ;;
  --model-sku)
    modelSku="$2"
    shift
    shift
    ;;
  --model-sku-capacity)
    modelSkuCapacity="$2"
    shift
    shift
    ;;
  *)
    POSITIONAL_ARGS+=("$1")
    shift
    ;;
  esac
done

set -- "${POSITIONAL_ARGS[0]}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

: "${location:="eastus"}" \
  "${resourceGroup:="GoAgentsResourceGroup"}" \
  "${cognitiveService:="GoAgentsCognitiveService"}" \
  "${cognitiveServiceSku:="S0"}" \
  "${cognitiveServiceKind:="OpenAI"}" \
  "${cognitiveServiceDomain:="go-agents-platform"}" \
  "${cognitiveServiceRole:="Cognitive Services OpenAI User"}" \
  "${modelDeployment:="o3-mini"}" \
  "${modelName:="o3-mini"}" \
  "${modelVersion:="2025-01-31"}" \
  "${modelFormat:="OpenAI"}" \
  "${modelSku:="GlobalStandard"}" \
  "${modelSkuCapacity:="10"}"

. "${SCRIPT_DIR}/../components/resource-group.sh" \
  --resource-group "$resourceGroup" \
  --location "$location"

. "${SCRIPT_DIR}/../components/cognitive-services-account.sh" \
  --kind "$cognitiveServiceKind" \
  --location "$location" \
  --name "$cognitiveService" \
  --resource-group "$resourceGroup" \
  --sku "$cognitiveServiceSku" \
  --domain "$cognitiveServiceDomain"

. "${SCRIPT_DIR}/../components/cognitive-services-deployment.sh" \
  --model-format "$modelFormat" \
  --model-name "$modelName" \
  --model-version "$modelVersion" \
  --name "$cognitiveService" \
  --resource-group "$resourceGroup" \
  --deployment-name "$modelDeployment" \
  --sku "$modelSku" \
  --sku-capacity "$modelSkuCapacity"

. "${SCRIPT_DIR}/../components/cognitive-services-grant-permissions.sh" \
  --name "$cognitiveService" \
  --role "$cognitiveServiceRole" \
  --resource-group "$resourceGroup"
