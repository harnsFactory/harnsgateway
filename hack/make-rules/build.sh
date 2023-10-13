#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

HARNSGATEWAY_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
source "${HARNSGATEWAY_ROOT}/hack/lib/init.sh"

harnsgateway::golang::build_binaries "$@"
harnsgateway::golang::place_bins
