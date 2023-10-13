#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail


readonly HARNSGATEWAY_GO_PACKAGE=harnsgateway
readonly HARNSGATEWAY_GOPATH="${HARNSGATEWAY_OUTPUT}/go"

harnsgateway::check::env() {
    errors=()
    if [ -z $GOPATH ]; then
        errors+="GOPATH environment value not set"
    fi

    # check other env

    # check length of errors
    if [[ ${#errors[@]} -ne 0 ]] ; then
        local error
        for error in "${errors[@]}"; do
            echo "Error: "$error
        done
        exit 1
    fi
}

harnsgateway::golang::binaries_from_targets() {
  local target
  for target in "$@"; do
    # If the target starts with what looks like a domain name, assume it has a
    # fully-qualified package name rather than one that needs the Kubernetes
    # package prepended.
    if [[ "${target}" =~ ^([[:alnum:]]+".")+[[:alnum:]]+"/" ]]; then
      echo "${target}"
    else
      echo "${HARNSGATEWAY_GO_PACKAGE}/${target}"
    fi
  done
}

# Asks golang what it thinks the host platform is. The go tool chain does some
# slightly different things when the target platform matches the host platform.
harnsgateway::golang::host_platform() {
  echo "$(go env GOHOSTOS)/$(go env GOHOSTARCH)"
}

harnsgateway::golang::sever_targets() {
  local targets=(
    cmd/gateway
  )
  echo "${targets[@]}"
}

IFS=" " read -ra HARNSGATEWAY_SERVER_TARGETS <<< "$(harnsgateway::golang::sever_targets)"
readonly HARNSGATEWAY_SERVER_TARGETS
readonly HARNSGATEWAY_SERVER_BINARIES=("${HARNSGATEWAY_SERVER_TARGETS[@]##*/}")

readonly HARNSGATEWAY_ALL_TARGETS=(
  "${HARNSGATEWAY_SERVER_TARGETS[@]}"
)
readonly HARNSGATEWAY_ALL_BINARIES=("${HARNSGATEWAY_ALL_TARGETS[@]##*/}")

harnsgateway::golang::is_statically_linked_library() {
  local e
  [[ "$(go env GOHOSTOS)" == "darwin" && "$(go env GOOS)" == "darwin" &&
    "$1" == *"/kubectl" ]] && return 1
  for e in "${HARNSGATEWAY_SERVER_BINARIES[@]}"; do [[ "${1}" == *"/${e}" ]] && return 0; done;
  return 1;
}

# Takes the platform name ($1) and sets the appropriate golang env variables
# for that platform.
harnsgateway::golang::set_platform_envs() {
  [[ -n ${1-} ]] || {
    echo "!!! Internal error. No platform set in harnsgateway::golang::set_platform_envs"
  }

  export GOOS=${platform%/*}
  export GOARCH=${platform##*/}

  # Do not set CC when building natively on a platform, only if cross-compiling
  if [[ $(harnsgateway::golang::host_platform) != "$platform" ]]; then
    # Dynamic CGO linking for other server architectures than host architecture goes here
    # If you want to include support for more server platforms than these, add arch-specific gcc names here
    case "${platform}" in
      "linux/amd64")
        export CGO_ENABLED=1
        export CC=${HARNSGATEWAY_LINUX_AMD64_CC:-x86_64-linux-gnu-gcc}
        ;;
      "linux/arm")
        export CGO_ENABLED=1
        export CC=${HARNSGATEWAY_LINUX_ARM_CC:-arm-linux-gnueabihf-gcc}
        ;;
      "linux/arm64")
        export CGO_ENABLED=1
        export CC=${HARNSGATEWAY_LINUX_ARM64_CC:-aarch64-linux-gnu-gcc}
        ;;
      "linux/ppc64le")
        export CGO_ENABLED=1
        export CC=${HARNSGATEWAY_LINUX_PPC64LE_CC:-powerpc64le-linux-gnu-gcc}
        ;;
      "linux/s390x")
        export CGO_ENABLED=1
        export CC=$HARNSGATEWAY_LINUX_S390X_CC:-s390x-linux-gnu-gcc}
        ;;
    esac
  fi

  # if CC is defined for platform then always enable it
  ccenv=$(echo "$platform" | awk -F/ '{print "HARNSGATEWAY_" toupper($1) "_" toupper($2) "_CC"}')
  if [ -n "${!ccenv-}" ]; then
    export CGO_ENABLED=1
    export CC="${!ccenv}"
  fi
}

harnsgateway::golang::create_gopath_tree() {
  local go_pkg_dir="${HARNSGATEWAY_GOPATH}/src/${HARNSGATEWAY_GO_PACKAGE}"
  local go_pkg_basedir
  go_pkg_basedir=$(dirname "${go_pkg_dir}")

  mkdir -p "${go_pkg_basedir}"

  # TODO: This symlink should be relative.
  if [[ ! -e "${go_pkg_dir}" || "$(readlink "${go_pkg_dir}")" != "${HARNSGATEWAY_ROOT}" ]]; then
    ln -snf "${HARNSGATEWAY_ROOT}" "${go_pkg_dir}"
  fi
}


harnsgateway::golang::setup_env() {
  harnsgateway::golang::create_gopath_tree

  export GOPATH="${HARNSGATEWAY_GOPATH}"
  export GOCACHE="${HARNSGATEWAY_GOPATH}/cache"

  # Make sure our own Go binaries are in PATH.
  export PATH="${HARNSGATEWAY_GOPATH}/bin:${PATH}"

  # Change directories so that we are within the GOPATH.  Some tools get really
  # upset if this is not true.  We use a whole fake GOPATH here to collect the
  # resultant binaries.  Go will not let us use GOBIN with `go install` and
  # cross-compiling, and `go install -o <file>` only works for a single pkg.
  local subdir
  subdir=$(harnsgateway::realpath . | sed "s|${HARNSGATEWAY_ROOT}||")
  cd "${HARNSGATEWAY_GOPATH}/src/${HARNSGATEWAY_GO_PACKAGE}/${subdir}" || return 1

  # Set GOROOT so binaries that parse code can work properly.
  GOROOT=$(go env GOROOT)
  export GOROOT

  # Unset GOBIN in case it already exists in the current session.
  unset GOBIN

  # This seems to matter to some tools
  export GO15VENDOREXPERIMENT=1
}

# install' will place binaries that match the host platform directly in $GOBIN
# while placing cross compiled binaries into `platform_arch` subdirs.  This
# complicates pretty much everything else we do around packaging and such.
harnsgateway::golang::place_bins() {
  local host_platform
  host_platform=$(harnsgateway::golang::host_platform)

  echo "Placing binaries"

  local -a platforms
  IFS=" " read -ra platforms <<< "${HARNSGATEWAY_BUILD_PLATFORMS:-}"
  if [[ ${#platforms[@]} -eq 0 ]]; then
      platforms=("${host_platform}")
  fi

  local platform
  for platform in "${platforms[@]}"; do
    # The substitution on platform_src below will replace all slashes with
    # underscores.  It'll transform darwin/amd64 -> darwin_amd64.
    local platform_src="/${platform//\//_}"
    if [[ "${platform}" == "${host_platform}" ]]; then
      platform_src=""
      rm -f "${THIS_PLATFORM_BIN}"
      ln -s "${HARNSGATEWAY_OUTPUT_BINPATH}/${platform}" "${THIS_PLATFORM_BIN}"
    fi

    local full_binpath_src="${HARNSGATEWAY_GOPATH}/bin${platform_src}"
    if [[ -d "${full_binpath_src}" ]]; then
      mkdir -p "${HARNSGATEWAY_OUTPUT_BINPATH}/${platform}"
      find "${full_binpath_src}" -maxdepth 1 -type f -exec \
        rsync -pc {} "${HARNSGATEWAY_OUTPUT_BINPATH}/${platform}" \;
    fi
  done
}

# Try and replicate the native binary placement of go install without
# calling go install.
harnsgateway::golang::outfile_for_binary() {
  local binary=$1
  local platform=$2
  local output_path="${HARNSGATEWAY_GOPATH}/bin"
  local bin
  bin=$(basename "${binary}")
  if [[ "${platform}" != "${host_platform}" ]]; then
    output_path="${output_path}/${platform//\//_}"
  fi
  if [[ ${GOOS} == "windows" ]]; then
    bin="${bin}.exe"
  fi
  echo "${output_path}/${bin}"
}

harnsgateway::golang::build_binaries_for_platform() {
  # This is for sanity.  Without it, user umasks can leak through.
  umask 0022

  local platform=$1

  local -a statics=()
  local -a nonstatics=()
  local -a tests=()

  echo "Env for ${platform}: GOOS=${GOOS-} GOARCH=${GOARCH-} GOROOT=${GOROOT-} CGO_ENABLED=${CGO_ENABLED-} CC=${CC-}"

  for binary in "${binaries[@]}"; do
    if [[ "${binary}" =~ ".test"$ ]]; then
      tests+=("${binary}")
    elif harnsgateway::golang::is_statically_linked_library "${binary}"; then
      statics+=("${binary}")
    else
      nonstatics+=("${binary}")
    fi
  done

  local -a build_args
  if [[ "${#statics[@]}" != 0 ]]; then
    build_args=(
      -installsuffix static
      ${goflags:+"${goflags[@]}"}
      -gcflags "${gogcflags:-}"
      -ldflags "${goldflags:-}"
    )
    echo "> static build CGO_ENABLED=0: ${statics[*]}"
    CGO_ENABLED=0 go install "${build_args[@]}" "${statics[@]}"
  fi

  if [[ "${#nonstatics[@]}" != 0 ]]; then
    build_args=(
      ${goflags:+"${goflags[@]}"}
      -gcflags "${gogcflags:-}"
      -ldflags "${goldflags:-}"
    )
    echo "> non-static build: ${nonstatics[*]}"
    go install "${build_args[@]}" "${nonstatics[@]}"
  fi

#  for test in "${tests[@]:+${tests[@]}}"; do
#    local outfile testpkg
#    outfile=$(harnsgateway::golang::outfile_for_binary "${test}" "${platform}")
#    testpkg=$(dirname "${test}")
#
#    mkdir -p "$(dirname "${outfile}")"
#    go test -c \
#      ${goflags:+"${goflags[@]}"} \
#      -gcflags "${gogcflags:-}" \
#      -ldflags "${goldflags:-}" \
#      -o "${outfile}" \
#      "${testpkg}"
#  done
}

harnsgateway::golang::build_binaries() {
    # Create a sub-shell so that we don't pollute the outer environment
    (
        harnsgateway::check::env
        harnsgateway::golang::setup_env

        local host_platform
        host_platform=$(harnsgateway::golang::host_platform)

        local goflags goldflags goasmflags gogcflags gotags
        # If GOLDFLAGS is unset, then set it to the a default of "-s -w".
        # Disable SC2153 for this, as it will throw a warning that the local
        # variable goldflags will exist, and it suggest changing it to this.
        # shellcheck disable=SC2153
        goldflags="${GOLDFLAGS=-s -w -buildid=} $(harnsgateway::version::ldflags)"
        gogcflags="${GOGCFLAGS:-} -trimpath=${HARNSGATEWAY_ROOT}"

        local -a targets=()
        local arg

        for arg in "$@"; do
            targets+=("${arg}")
        done

        if [[ ${#targets[@]} -eq 0 ]]; then
            targets=("${HARNSGATEWAY_ALL_TARGETS[@]}")
        fi

        local -a platforms
        IFS=" " read -ra platforms <<< "${HARNSGATEWAY_BUILD_PLATFORMS:-}"
        if [[ ${#platforms[@]} -eq 0 ]]; then
            platforms=("${host_platform}")
        fi

        local -a binaries
        while IFS="" read -r binary; do binaries+=("$binary"); done < <(harnsgateway::golang::binaries_from_targets "${targets[@]}")

        for platform in "${platforms[@]}"; do
            echo "Building go targets for ${platform}:" "${targets[@]}"
            (
                harnsgateway::golang::set_platform_envs "${platform}"
                harnsgateway::golang::build_binaries_for_platform "${platform}"
            )
        done
    )
}
