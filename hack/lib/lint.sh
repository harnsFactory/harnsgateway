#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SED_CMD=""

if [[ "$OSTYPE" == "darwin"* ]]
then
    SED_CMD=`which gsed`
    if [ -z $SED_CMD ]
    then
        echo "Please install gnu-sed (brew install gnu-sed)"
        exit 1
    fi
elif [[ "$OSTYPE" == "linux"* ]]
then
    SED_CMD=`which sed`
    if [ -z $SED_CMD ]
    then
        echo "Please install sed"
        exit 1
    fi
else
    echo "Unsupported OS $OSTYPE"
    exit 1
fi

harnsgateway::lint::check() {
    cd ${HARNSGATEWAY_ROOT}
    echo "start lint ..."
    set +o pipefail
    echo "check any whitenoise ..."
    # skip deleted files
    if [[ "$OSTYPE" == "darwin"* ]]
    then
        git diff --cached --name-only --diff-filter=ACRMTU master | grep -Ev "externalversions|fake|vendor|images|adopters" | xargs $SED_CMD -i 's/[ \t]*$//'
    elif [[ "$OSTYPE" == "linux"* ]]
    then
        git diff --cached --name-only --diff-filter=ACRMTU master | grep -Ev "externalversions|fake|vendor|images|adopters" | xargs --no-run-if-empty $SED_CMD -i 's/[ \t]*$//'
    else
        echo "Unsupported OS $OSTYPE"
        exit 1
    fi

    [[ $(git diff --name-only) ]] && {
        echo "Some files have white noise issue, please run \`make lint\` to slove this issue"
        return 1
    }
    set -o pipefail

    echo "check any issue by golangci-lint ..."
    GOOS="linux" golangci-lint run -v

    # only check format issue under staging dir
    echo "check any issue under staging dir by gofmt ..."
    gofmt -l -w staging
}
