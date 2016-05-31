#!/bin/bash

if [[ -z "$CIRCLECI" ]]; then
	echo "Not on CircleCI" 1>&2
	exit 1
fi

set -e

export GOPATH="$HOME/workspace"

PKG_PATH_DIR="$GOPATH/src/github.com/$CIRCLE_PROJECT_USERNAME"
PKG_PATH="$PKG_PATH_DIR/$CIRCLE_PROJECT_REPONAME"

create_workspace() {
	mkdir -p "$PKG_PATH_DIR"
	ln -s "$(pwd)" "$PKG_PATH"
}

fetch_dependencies() {
	cd "$PKG_PATH"
	go get -v -d ./...
}

cross_compile() {
	cd "$PKG_PATH"
	for os in linux darwin windows; do
		echo "---> Building ${os}/amd64"
		GOOS="$os" GOARCH='amd64' go build \
			-o "build/steemreduce_${os}_amd64" 'github.com/tchap/steemreduce'
	done
}

archive_artifacts() {
	cd "$PKG_PATH/build"
	cp * "$CIRCLE_ARTIFACTS/"
}

case "$1" in
	setup)
		create_workspace
		;;
	deps)
		fetch_dependencies
		;;
	compile)
		cross_compile
		;;
	archive)
		archive_artifacts
		;;
	*)
		exit 1
		;;
esac
