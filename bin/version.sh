#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

VERSION=$(git tag -l 'v*' --sort=-version:refname | head -n1)

if [[ "$VERSION" == "" ]]; then
	echo "ERR: current version tag missing"
	exit 1
fi

if [[ ! "${VERSION}" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)(.*)$ ]]; then
	echo "ERR: current version ${VERSION} does not match scheme"
	exit 1
fi

export VERSION
export V_MAJOR="${BASH_REMATCH[1]}"
export V_MINOR="${BASH_REMATCH[2]}"
export V_PATCH="${BASH_REMATCH[3]}"
export V_EXTRA="${BASH_REMATCH[4]}"

echo "${VERSION}"
