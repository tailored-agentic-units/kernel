#!/bin/bash

unset resourceGroup
unset cognitiveService

POSITIONAL_ARGS=()

while [[ $# -gt 0 ]]; do
  case $1 in
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
  *)
    POSITIONAL_ARGS+=("$1")
    shift
    ;;
  esac
done

set -- "${POSITIONAL_ARGS[0]}"

: "${resourceGroup:="GoAgentsResourceGroup"}" \
  "${cognitiveService:="GoAgentsCognitiveService"}"

az cognitiveservices account keys list \
  --name "$cognitiveService" \
  --resource-group "$resourceGroup" |
  jq -r .key1
