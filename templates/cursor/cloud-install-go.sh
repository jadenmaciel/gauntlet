#!/usr/bin/env bash
set -euo pipefail

version="${GAUNTLET_VERSION:-v0.1.0}"
dest="${GAUNTLET_BIN_DIR:-$HOME/.local/bin}"

if ! command -v go >/dev/null; then
	echo "cloud-install-go: go is not on PATH" >&2
	exit 2
fi

mkdir -p "$dest"
GOBIN="$dest" go install "github.com/jadenmaciel/gauntlet/cmd/crap4go@${version}"
GOBIN="$dest" go install "github.com/jadenmaciel/gauntlet/cmd/gauntlet@${version}"

if [ -n "${GITHUB_PATH:-}" ]; then
	echo "$dest" >> "$GITHUB_PATH"
fi

case ":${PATH}:" in
*":${dest}:"*) ;;
*)
	export PATH="${dest}:${PATH}"
	echo "Add ${dest} to PATH: export PATH=\"${dest}:\$PATH\""
	;;
esac
