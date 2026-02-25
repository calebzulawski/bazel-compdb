#!/usr/bin/env bash
set -euo pipefail

# Release workflows set VERSION (for example: v1.2.3).
# Strip a leading "v" so version output is consistent for both BCR and releases.
version="${VERSION:-}"
version="${version#v}"

echo "STABLE_VERSION ${version}"
