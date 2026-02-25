#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail

TAG="$1"
VERSION="${TAG#v}"
PREFIX="bazel-compdb-${VERSION}"
ARCHIVE="bazel-compdb-$TAG.tar.gz"
WORKSPACE_DIR="${GITHUB_WORKSPACE:-$PWD}"
ARTIFACTS_DIR="${WORKSPACE_DIR}/artifacts"

git archive --format=tar --prefix=${PREFIX}/ ${TAG} | gzip >"$ARCHIVE"
SHA=$(shasum -a 256 "$ARCHIVE" | awk '{print $1}')

find "${ARTIFACTS_DIR}" -type f -name "*.tar.gz" -exec cp -f -t "${WORKSPACE_DIR}" {} +

cat <<EOF
## Using a prebuilt version

Download the release for your platform.

## Using Bzlmod

Paste this snippet into your \`MODULE.bazel\` file:

\`\`\`starlark
bazel_dep(name = "bazel-compdb", version = "${VERSION}")
\`\`\`

EOF
