#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

usage() {
	echo -e "gxctl release script"
	echo -e "usage: $0 [-pmMy]\n"
	echo -e "-p\tpatch release"
	echo -e "-m\tminor release"
	echo -e "-M\tmajor release"
	echo -e "-y\tskip confirmation"
	echo -e "-d\tdry run - generate change log and build release, but do not commit anything"
}

RELEASE_SCOPE=""
SKIP_CONFIRM=""
DRY_RUN=""
GH_TOKEN="${GH_TOKEN:-}"

PLATFORMS=("linux/amd64" "darwin/amd64" "windows/amd64")
CHANGELOG="CHANGELOG.md"
README="README.md"

OPTERR=0
while getopts dmMpy OPT; do
	case "${OPT}" in
	m)
		RELEASE_SCOPE="minor"
		;;
	M)
		RELEASE_SCOPE="major"
		;;
	p)
		RELEASE_SCOPE="patch"
		;;
	y)
		SKIP_CONFIRM="yes"
		;;
	d)
		DRY_RUN="yes"
		;;
	h)
		usage
		exit 0
		;;
	*)
		echo -e "unknown option\n"
		usage
		exit 1
		;;
	esac
done

main() {

	BRANCH="$(git rev-parse --abbrev-ref HEAD)"
	if [[ "${BRANCH}" != "master" ]]; then
		echo "WARN: must be on master to release"
		DRY_RUN="yes"
  fi

	if [[ "${GH_TOKEN}" == "" ]]; then
		echo "WARN: must set GH_TOKEN to GitHub access token with repository rights"
		DRY_RUN="yes"
	fi

	if ! git diff-files --quiet; then
		if [[ "${DRY_RUN}" == "" ]]; then
			echo "ERR: unstaged changes"
			return 1
		else
			echo "WARN: unstaged changes"
			DRY_RUN="yes"
		fi
	fi

	if ! git diff-index --quiet --cached HEAD --; then
		if [[ "${DRY_RUN}" == "" ]]; then
			echo "ERR: uncommited changes"
			return 1
		else
			echo "WARN: uncommited changes"
			DRY_RUN="yes"
		fi
	fi

	if [[ "${DRY_RUN}" != "" ]]; then
		echo "Continuing in dry run mode."
	fi

	if [[ "${DRY_RUN}" == "" ]]; then
		if [[ "${SKIP_CONFIRM}" == "" ]]; then
			read -p "Pull master? " -n 1 -r
			echo
			if [[ ! $REPLY =~ ^[Yy]$ ]]; then
				return 0
			fi
		fi
		git pull
	fi

	. bin/version.sh

	case "${RELEASE_SCOPE}" in
	major)
		V_NEXT="v$((V_MAJOR + 1)).0.0"
		;;
	minor)
		V_NEXT="v${V_MAJOR}.$((V_MINOR + 1)).0"
		;;
	patch)
		V_NEXT="v${V_MAJOR}.${V_MINOR}.$((V_PATCH + 1))"
		;;
	*)
		echo "RELEASE_SCOPE must be one of [major,minor,patch]"
		return 1
		;;
	esac

	echo -e "Next Version:    $V_NEXT\n"

	if [[ "${SKIP_CONFIRM}" == "" ]]; then
		read -p "Continue with ${RELEASE_SCOPE} release? " -n 1 -r
		echo
		if [[ ! $REPLY =~ ^[Yy]$ ]]; then
			return 0
		fi
	fi

	echo -e "Generate Changelog…\n"
	COMMITS=$(git log "${VERSION}..HEAD" --pretty=format:"%H" --no-merges)
	CHANGES="### ${V_NEXT}\n"
	for REF in $COMMITS; do
		MSG=$(git log -1 ${REF} --pretty=format:"%s")
		CHANGES+="${MSG}\n"
	done
	echo -e "# Changelog\n${CHANGES}\n$(cat ${CHANGELOG})\n" >"${CHANGELOG}"
	echo -e "${CHANGES}"

	if [[ "${DRY_RUN}" == "" ]]; then
		echo "NOTE: You may manually edit ${CHANGELOG} before continuing"
		if [[ "${SKIP_CONFIRM}" == "" ]]; then
			read -p "Stage Changelog? " -n 1 -r
			echo
			if [[ ! $REPLY =~ ^[Yy]$ ]]; then
				git checkout -- "${CHANGELOG}"
				return 0
			fi
		fi
		git stage ${CHANGELOG}
	fi

	echo "Build…"
	for PLATFORM in "${PLATFORMS[@]}"; do
		IFS='/' read -ra TARGET <<<"$PLATFORM"
		export GOOS=${TARGET[0]}
		export GOARCH=${TARGET[1]}
		BUILD="bin/gxctl-${GOOS}-${GOARCH}"
		make build
		if [ $GOOS = "windows" ]; then
			mv "${BUILD}" "${BUILD}.exe"
			BUILD="${BUILD}.exe"
		fi
		mkdir -p dist
		zip -j "dist/gxctl-${V_NEXT}-${GOOS}-${GOARCH}.zip" "${BUILD}" "${CHANGELOG}" "${README}"
	done

	if [[ "${DRY_RUN}" == "yes" ]]; then
		echo "Skip GitHub release in dry run."
		return 0
	fi

	if [[ "${SKIP_CONFIRM}" == "" ]]; then
		read -p "Checkout release branch, commit and tag? " -n 1 -r
		echo
		if [[ ! $REPLY =~ ^[Yy]$ ]]; then
			git checkout -- "${CHANGELOG}"
			return 0
		fi
	fi
	BRANCH="feat/release_${V_NEXT}"
	git checkout -b "${BRANCH}"
	git commit -m "feat: release ${V_NEXT}"
	git tag -d "${V_NEXT}"

	if [[ "${SKIP_CONFIRM}" == "" ]]; then
		read -p "Push release branch? " -n 1 -r
		echo
		if [[ ! $REPLY =~ ^[Yy]$ ]]; then
			return 0
		fi
	fi

	git push --set-upstream origin "${BRANCH}"

	if [[ "${SKIP_CONFIRM}" == "" ]]; then
		read -p "Create GitHub release? " -n 1 -r
		echo
		if [[ ! $REPLY =~ ^[Yy]$ ]]; then
			return 0
		fi
	fi

	API_BASE="https://api.github.com/repos/grid-x/gxctl"
	AUTH_HEADER="Authorization:token ${GH_TOKEN}"
	BIN_HEADER="Content-Type: application/octet-stream"

	IFS=$' '

	echo "get releases…"
	declare -a RELEASES
	RELEASES=($(http --check-status "${API_BASE}/releases" "${AUTH_HEADER}" | jq -r 'map(.tag_name) | @sh'))
	echo "  ${RELEASES[@]}"

	for TAG in "${RELEASES[@]}"; do
		TAG=$(echo $TAG | tr -d \')
		if [[ "${TAG}" == "${V_NEXT}" ]]; then
			echo "ERR: ${TAG} already exists!"
			exit 1
		fi
	done

	echo "create release…"
	RELEASE=$(
		http --check-status POST "${API_BASE}/releases" "${AUTH_HEADER}" \
			tag_name="${V_NEXT}" \
			draft:=true \
			target_commitish="${V_NEXT}" \
			body=@CHANGELOG.md
	)
	RELEASE_ID=$(echo $RELEASE | jq .id)

	API_ASSETS="https://uploads.github.com/repos/grid-x/gxctl/releases/${RELEASE_ID}/assets"
	for FILE in dist/*.zip; do
		case $FILE in
		dist/gxctl-${V_NEXT}*)
			echo "upload ${FILE}…"
			NAME=$(http --check-status POST "${API_ASSETS}?name=$(basename ${FILE})" "${AUTH_HEADER}" "${BIN_HEADER}" <${FILE})
			;;
		esac
	done

	echo "Success!"
	echo "Review draft release @ $(echo ${RELEASE} | jq -r .html_url)"
}

main
