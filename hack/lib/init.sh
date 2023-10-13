#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail


HARNSGATEWAY_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"

HARNSGATEWAY_OUTPUT_SUBPATH="${HARNSGATEWAY_OUTPUT_SUBPATH:-_output/local}"
HARNSGATEWAY_OUTPUT="${HARNSGATEWAY_ROOT}/${HARNSGATEWAY_OUTPUT_SUBPATH}"
HARNSGATEWAY_OUTPUT_BINPATH="${HARNSGATEWAY_OUTPUT}/bin"

export THIS_PLATFORM_BIN="${HARNSGATEWAY_ROOT}/_output/bin"

source "${HARNSGATEWAY_ROOT}/hack/lib/util.sh"
source "${HARNSGATEWAY_ROOT}/hack/lib/lint.sh"
source "${HARNSGATEWAY_ROOT}/hack/lib/version.sh"
source "${HARNSGATEWAY_ROOT}/hack/lib/golang.sh"

function harnsgateway::readlinkdashf {
  # run in a subshell for simpler 'cd'
  (
    if [[ -d "${1}" ]]; then # This also catch symlinks to dirs.
      cd "${1}"
      pwd -P
    else
      cd "$(dirname "${1}")"
      local f
      f=$(basename "${1}")
      if [[ -L "${f}" ]]; then
        readlink "${f}"
      else
        echo "$(pwd -P)/${f}"
      fi
    fi
  )
}


harnsgateway::realpath() {
  if [[ ! -e "${1}" ]]; then
    echo "${1}: No such file or directory" >&2
    return 1
  fi
  harnsgateway::readlinkdashf "${1}"
}
