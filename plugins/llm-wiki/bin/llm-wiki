#!/bin/sh
# Selects the packaged runtime for the host OS and CPU. Diagnostics go to
# stderr: stdout of `serve` is reserved for the MCP protocol.
set -eu
here=$(cd "$(dirname "$0")" && pwd)
case "$(uname -s)" in
Darwin)
	runtime=llm-wiki-darwin
	;;
Linux)
	case "$(uname -m)" in
	x86_64 | amd64) runtime=llm-wiki-linux-amd64 ;;
	aarch64 | arm64) runtime=llm-wiki-linux-arm64 ;;
	*)
		echo "llm-wiki: unsupported CPU $(uname -m)" >&2
		exit 1
		;;
	esac
	;;
*)
	echo "llm-wiki: unsupported OS $(uname -s)" >&2
	exit 1
	;;
esac
exec "$here/$runtime" "$@"
