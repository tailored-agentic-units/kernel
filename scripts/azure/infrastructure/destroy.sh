#!/bin/bash

unset resourceGroup
unset location
unset cognitiveService

POSITIONAL_ARGS=()

while [[ $# -gt 0 ]]; do
  case $1 in
  --resource-group)
    resourceGroup="$2"
    shift
    shift
    ;;
  --location)
    location="$2"
    shift
    shift
    ;;
  --cognitive-service)
    cognitiveService="$2"
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

# default value if not provided
: "${resourceGroup:="GoAgentsResourceGroup"}" \
  "${location:="eastus"}" \
  "${cognitiveService:="GoAgentsCognitiveService"}"

if [[ $(az group list --query "[?name=='$resourceGroup'] | length(@)") -gt 0 ]]; then
  az group delete --resource-group "$resourceGroup" -y
else
  echo "$resourceGroup does not exist"
fi

. "${SCRIPT_DIR}/../components/purge-cognitive-services-account.sh" \
  --location "$location" \
  --name "$cognitiveService" \
  --resource-group "$resourceGroup"
