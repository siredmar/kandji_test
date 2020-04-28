#!/usr/bin/env bash
set -euo pipefail

commitGetDiff() {
    if [ "${BUILDKITE_BRANCH}" = "master" ]; then
        PARENT_COMMIT=$(git log -2 --merges --pretty=format:"%H" | tail -1)
    else
        PARENT_COMMIT=$(git log -1 --merges --pretty=format:"%H" | tail -1)
    fi

    printf "Comparing HEAD to %s\\n" "${PARENT_COMMIT}" 1>&2

    git diff "${PARENT_COMMIT}"...HEAD --name-only
}
export -f commitGetDiff

COMMIT_DIFF=$(commitGetDiff)
if [ 0 -lt "$(echo "${COMMIT_DIFF}" | grep -E "VERSION" | wc -w )" ] 
  then
    exit 1
  else
    exit 0
fi
