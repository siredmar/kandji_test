#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

VERSION=$(git tag -l 'v*' --sort=-version:refname | head -n1)

if [[ "$VERSION" == "" ]]; then
	echo "ERR: current version tag missing"
	exit 1
fi

if [[ ! "${VERSION}" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
	echo "ERR: current version ${VERSION} does not match scheme"
	exit 1
fi

echo $VERSION
