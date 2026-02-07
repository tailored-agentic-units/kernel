#!/bin/bash

unset output

POSTITIONAL_ARGS=()

while [[ $# -gt 0 ]]; do
  case $1 in
  --output)
    output="$2"
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

: "${output:="table"}"

az account list-locations --output $output
