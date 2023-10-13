#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# Version management helpers.  These functions help to set, save and load the
# following variables:
#
#    HARNSGATEWAY_GIT_COMMIT - The git commit id corresponding to this
#          source code.
#    HARNSGATEWAY_GIT_TREE_STATE - "clean" indicates no changes since the git commit id
#        "dirty" indicates source code changes after the git commit id
#        "archive" indicates the tree was produced by 'git archive'
#    HARNSGATEWAY_GIT_VERSION - "vX.Y" used to indicate the last release version.
#    HARNSGATEWAY_GIT_MAJOR - The major part of the version
#    HARNSGATEWAY_GIT_MINOR - The minor component of the version

# Grovels through git to set a set of env variables.
#
# If HARNSGATEWAY_GIT_VERSION_FILE, this function will load from that file instead of
# querying git.
harnsgateway::version::get_version_vars() {
  if [[ -n ${HARNSGATEWAY_GIT_VERSION_FILE-} ]]; then
    harnsgateway::version::load_version_vars "${HARNSGATEWAY_GIT_VERSION_FILE}"
    return
  fi

  local git=(git --work-tree "${HARNSGATEWAY_ROOT}")

  if [[ -n ${HARNSGATEWAY_GIT_COMMIT-} ]] || HARNSGATEWAY_GIT_COMMIT=$("${git[@]}" rev-parse "HEAD^{commit}" 2>/dev/null); then
    if [[ -z ${HARNSGATEWAY_GIT_TREE_STATE-} ]]; then
      # Check if the tree is dirty.  default to dirty
      if git_status=$("${git[@]}" status --porcelain 2>/dev/null) && [[ -z ${git_status} ]]; then
        HARNSGATEWAY_GIT_TREE_STATE="clean"
      else
        HARNSGATEWAY_GIT_TREE_STATE="dirty"
      fi
    fi

    # Use git describe to find the version based on tags.
    if [[ -n ${HARNSGATEWAY_GIT_VERSION-} ]] || HARNSGATEWAY_GIT_VERSION=$("${git[@]}" describe --tags --match='v*' --abbrev=14 "${HARNSGATEWAY_GIT_COMMIT}^{commit}" 2>/dev/null); then
      # This translates the "git describe" to an actual semver.org
      # compatible semantic version that looks something like this:
      #   v1.1.0-alpha.0.6+84c76d1142ea4d
      #
      # TODO: We continue calling this "git version" because so many
      # downstream consumers are expecting it there.
      #
      # These regexes are painful enough in sed...
      # We don't want to do them in pure shell, so disable SC2001
      # shellcheck disable=SC2001
      DASHES_IN_VERSION=$(echo "${HARNSGATEWAY_GIT_VERSION}" | sed "s/[^-]//g")
      if [[ "${DASHES_IN_VERSION}" == "---" ]] ; then
        # shellcheck disable=SC2001
        # We have distance to subversion (v1.1.0-subversion-1-gCommitHash)
        HARNSGATEWAY_GIT_VERSION=$(echo "${HARNSGATEWAY_GIT_VERSION}" | sed "s/-\([0-9]\{1,\}\)-g\([0-9a-f]\{14\}\)$/.\1\+\2/")
      elif [[ "${DASHES_IN_VERSION}" == "--" ]] ; then
        # shellcheck disable=SC2001
        # We have distance to base tag (v1.1.0-1-gCommitHash)
        HARNSGATEWAY_GIT_VERSION=$(echo "${HARNSGATEWAY_GIT_VERSION}" | sed "s/-g\([0-9a-f]\{14\}\)$/+\1/")
      fi
      if [[ "${HARNSGATEWAY_GIT_TREE_STATE}" == "dirty" ]]; then
        # git describe --dirty only considers changes to existing files, but
        # that is problematic since new untracked .go files affect the build,
        # so use our idea of "dirty" from git status instead.
        HARNSGATEWAY_GIT_VERSION+="-dirty"
      fi


      # Try to match the "git describe" output to a regex to try to extract
      # the "major" and "minor" versions and whether this is the exact tagged
      # version or whether the tree is between two tagged versions.
      if [[ "${HARNSGATEWAY_GIT_VERSION}" =~ ^v([0-9]+)\.([0-9]+)(\.[0-9]+)?([-].*)?([+].*)?$ ]]; then
        HARNSGATEWAY_GIT_MAJOR=${BASH_REMATCH[1]}
        HARNSGATEWAY_GIT_MINOR=${BASH_REMATCH[2]}
        if [[ -n "${BASH_REMATCH[4]}" ]]; then
          HARNSGATEWAY_GIT_MINOR+="+"
        fi
      fi

      # If HARNSGATEWAY_GIT_VERSION is not a valid Semantic Version, then refuse to build.
      if ! [[ "${HARNSGATEWAY_GIT_VERSION}" =~ ^v([0-9]+)\.([0-9]+)(\.[0-9]+)?(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$ ]]; then
          echo "HARNSGATEWAY_GIT_VERSION should be a valid Semantic Version. Current value: ${HARNSGATEWAY_GIT_VERSION}"
          echo "Please see more details here: https://semver.org"
          exit 1
      fi
    fi
  fi
}

# Saves the environment flags to $1
harnsgateway::version::save_version_vars() {
  local version_file=${1-}
  [[ -n ${version_file} ]] || {
    echo "!!! Internal error.  No file specified in harnsgateway::version::save_version_vars"
    return 1
  }

  cat <<EOF >"${version_file}"
HARNSGATEWAY_GIT_COMMIT='${HARNSGATEWAY_GIT_COMMIT-}'
HARNSGATEWAY_GIT_TREE_STATE='${HARNSGATEWAY_GIT_TREE_STATE-}'
HARNSGATEWAY_GIT_VERSION='${HARNSGATEWAY_GIT_VERSION-}'
HARNSGATEWAY_GIT_MAJOR='${HARNSGATEWAY_GIT_MAJOR-}'
HARNSGATEWAY_GIT_MINOR='${HARNSGATEWAY_GIT_MINOR-}'
EOF
}

# Loads up the version variables from file $1
harnsgateway::version::load_version_vars() {
  local version_file=${1-}
  [[ -n ${version_file} ]] || {
    echo "!!! Internal error.  No file specified in harnsgateway::version::load_version_vars"
    return 1
  }

  source "${version_file}"
}

# Prints the value that needs to be passed to the -ldflags parameter of go build
# in order to set the Kubernetes based on the git tree status.
# IMPORTANT: if you update any of these, also update the lists in
# pkg/version.
harnsgateway::version::ldflags() {
  harnsgateway::version::get_version_vars

  local -a ldflags
  function add_ldflag() {
    local key=${1}
    local val=${2}
    ldflags+=(
      "-X '${HARNSGATEWAY_GO_PACKAGE}/pkg/version.${key}=${val}'"
    )
  }

  add_ldflag "buildDate" "$(date ${SOURCE_DATE_EPOCH:+"--date=@${SOURCE_DATE_EPOCH}"} -u +'%Y-%m-%dT%H:%M:%SZ')"
  if [[ -n ${HARNSGATEWAY_GIT_COMMIT-} ]]; then
    add_ldflag "gitCommit" "${HARNSGATEWAY_GIT_COMMIT}"
    add_ldflag "gitTreeState" "${HARNSGATEWAY_GIT_TREE_STATE}"
  fi

  if [[ -n ${HARNSGATEWAY_GIT_VERSION-} ]]; then
    add_ldflag "gitVersion" "${HARNSGATEWAY_GIT_VERSION}"
  fi

  if [[ -n ${HARNSGATEWAY_GIT_MAJOR-} && -n ${HARNSGATEWAY_GIT_MINOR-} ]]; then
    add_ldflag "gitMajor" "${HARNSGATEWAY_GIT_MAJOR}"
    add_ldflag "gitMinor" "${HARNSGATEWAY_GIT_MINOR}"
  fi

  # The -ldflags parameter takes a single string, so join the output.
  echo "${ldflags[*]-}"
}
