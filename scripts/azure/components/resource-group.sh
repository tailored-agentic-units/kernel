#!/bin/bash

unset location
unset resourceGroup

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
  *)
    POSITIONAL_ARGS+=("$1")
    shift
    ;;
  esac
done

set -- "${POSITIONAL_ARGS[0]}"

# default values if not provided
: "${resourceGroup:="GoAgentsResourceGroup"}" \
  "${location:="eastus"}"

if [[ $(az group list --query "[?name=='$resourceGroup'] | length(@)") -gt 0 ]]; then
  echo "$resourceGroup already exists"
else
  az group create \
    -g "$resourceGroup" \
    -l "$location"
fi
