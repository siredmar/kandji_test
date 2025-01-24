#!/usr/bin/env bash

# Exit on errors
set -e

# Arguments
VERSION=$1
CHANGELOG_PATH=$2
BUILD_FOR_EIS=${3:-false}

echo "Version: $VERSION"
echo "Changelog Path: $CHANGELOG_PATH"
echo "Build for EIS: $BUILD_FOR_EIS"

# Variables
PROJECT_NAME="gxctl"
FULL_COMMIT="$(git rev-parse HEAD || echo unknown)"
BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
OUTPUT_DIR=$"dist"

# Supported platforms and architectures
GOOS_LIST=("linux" "windows" "darwin")
GOARCH_LIST=("amd64" "arm64")

# LDFLAGS for builds
LDFLAGS="-w -s -X 'github.com/grid-x/gxctl/internal/version.GitCommit=${FULL_COMMIT}' \
-X 'github.com/grid-x/gxctl/internal/version.BuildTime=${BUILD_TIME}' \
-X 'github.com/grid-x/gxctl/internal/version.Version=${VERSION}'"

# Clean up and prepare output directory
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"
OUTPUT_DIR=$(realpath "$OUTPUT_DIR")

# Build for each platform and architecture
for GOOS in "${GOOS_LIST[@]}"; do
  for GOARCH in "${GOARCH_LIST[@]}"; do
    OUTPUT_NAME="${PROJECT_NAME}_${VERSION}_"
    if [ "$GOOS" == "darwin" ]; then
      OUTPUT_NAME+="Mac"
    else
      OUTPUT_NAME+="$(echo "$GOOS" | sed 's/.*/\u&/')"
    fi
    OUTPUT_NAME+="_${GOARCH}"

    if [ "$GOOS" == "windows" ]; then
      OUTPUT_NAME+=".exe"
    fi

    echo "Building $OUTPUT_NAME for $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -o "$OUTPUT_DIR/$OUTPUT_NAME" -ldflags "$LDFLAGS" ./cmd/gxctl
  done
done


# Build for each platform and architecture
for GOOS in "${GOOS_LIST[@]}"; do
  for GOARCH in "${GOARCH_LIST[@]}"; do
    # Recompute the binary output name for this GOOS and GOARCH
    BINARY_NAME="${PROJECT_NAME}_${VERSION}_"
    if [ "$GOOS" == "darwin" ]; then
      BINARY_NAME+="Mac"
    else
      BINARY_NAME+="$(echo "$GOOS" | sed 's/.*/\u&/')"
    fi
    BINARY_NAME+="_${GOARCH}"

    if [ "$GOOS" == "windows" ]; then
      BINARY_NAME+=".exe"
    fi

    ARCHIVE_NAME="${PROJECT_NAME}_${VERSION}_"
    if [ "$GOOS" == "darwin" ]; then
      ARCHIVE_NAME+="Mac"
    else
      ARCHIVE_NAME+="$(echo "$GOOS" | sed 's/.*/\u&/')"
    fi
    ARCHIVE_NAME+="_${GOARCH}"

    TEMP_DIR=$(mktemp -d)

    # Include changelog if provided
    CHANGELOG_FILE=""
    if [ ! -z "$CHANGELOG_PATH" ]; then
      cp "${CHANGELOG_PATH}" "$TEMP_DIR/CHANGELOG.md"
      CHANGELOG_FILE=CHANGELOG.md
    fi
    cp docs/README.md "$TEMP_DIR/README.md"
    cp config/gridx/config.yaml "$TEMP_DIR/config.yaml"

    ROOT=$(pwd)
    if [ "$GOOS" == "linux" ]; then
      ARCHIVE_NAME+=".tar.gz"
      echo "Packaging $ARCHIVE_NAME..."
      cp "$OUTPUT_DIR/$BINARY_NAME" "$TEMP_DIR/gxctl"
      cd "$TEMP_DIR"
      tar -czf "$OUTPUT_DIR/$ARCHIVE_NAME" gxctl ${CHANGELOG_FILE} README.md config.yaml
      cd "$ROOT"
    else
      ARCHIVE_NAME+=".zip"
      echo "Packaging $ARCHIVE_NAME..."
      if [ "$GOOS" == "windows" ]; then
        cp "$OUTPUT_DIR/$BINARY_NAME" "$TEMP_DIR/gxctl.exe"
      else
        cp "$OUTPUT_DIR/$BINARY_NAME" "$TEMP_DIR/gxctl"
      fi
      cd "$TEMP_DIR"
      zip -j "$OUTPUT_DIR/$ARCHIVE_NAME" gxctl* ${CHANGELOG_FILE} README.md config.yaml
      cd "$ROOT"
    fi

    rm -r "$TEMP_DIR"
  done
done

if [ "$BUILD_FOR_EIS" != "true" ]; then
  exit 0
fi
# Build for each platform and architecture
for GOOS in "${GOOS_LIST[@]}"; do
  for GOARCH in "${GOARCH_LIST[@]}"; do
    # Recompute the binary output name for this GOOS and GOARCH
    BINARY_NAME="${PROJECT_NAME}_eis_${VERSION}_"
    if [ "$GOOS" == "darwin" ]; then
      BINARY_NAME+="Mac"
    else
      BINARY_NAME+="$(echo "$GOOS" | sed 's/.*/\u&/')"
    fi
    BINARY_NAME+="_${GOARCH}"

    if [ "$GOOS" == "windows" ]; then
      BINARY_NAME+=".exe"
    fi

    ARCHIVE_NAME="${PROJECT_NAME}_${VERSION}_"
    if [ "$GOOS" == "darwin" ]; then
      ARCHIVE_NAME+="Mac"
    else
      ARCHIVE_NAME+="$(echo "$GOOS" | sed 's/.*/\u&/')"
    fi
    ARCHIVE_NAME+="_${GOARCH}"

    TEMP_DIR=$(mktemp -d)

    # Include changelog if provided
    CHANGELOG_FILE=""
    if [ ! -z "$CHANGELOG_PATH" ]; then
      cp "${CHANGELOG_PATH}" "$TEMP_DIR/CHANGELOG.md"
      CHANGELOG_FILE=CHANGELOG.md
    fi
    cp docs/README.md "$TEMP_DIR/README.md"
    cp config/eis/config.yaml "$TEMP_DIR/config.yaml"

    ROOT=$(pwd)
    if [ "$GOOS" == "linux" ]; then
      ARCHIVE_NAME+=".tar.gz"
      echo "Packaging $ARCHIVE_NAME..."
      cp "$OUTPUT_DIR/$BINARY_NAME" "$TEMP_DIR/gxctl"
      cd "$TEMP_DIR"
      tar -czf "$OUTPUT_DIR/$ARCHIVE_NAME" gxctl ${CHANGELOG_FILE} README.md config.yaml
      cd "$ROOT"
    else
      ARCHIVE_NAME+=".zip"
      echo "Packaging $ARCHIVE_NAME..."
      if [ "$GOOS" == "windows" ]; then
        cp "$OUTPUT_DIR/$BINARY_NAME" "$TEMP_DIR/gxctl.exe"
      else
        cp "$OUTPUT_DIR/$BINARY_NAME" "$TEMP_DIR/gxctl"
      fi
      cd "$TEMP_DIR"
      zip -j "$OUTPUT_DIR/$ARCHIVE_NAME" gxctl* ${CHANGELOG_FILE} README.md config.yaml
      cd "$ROOT"
    fi

    rm -r "$TEMP_DIR"
  done
done
