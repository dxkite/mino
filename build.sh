#!/bin/bash
VERSION=$(git describe --tags)
COMMIT=$(git rev-parse --short HEAD)
OS="linux"

if [[ $OSTYPE == darwin* ]]; then
    OS="macos"
fi

echo "build $VERSION-$COMMIT-$OS"
go build -o "mino-$VERSION-$OS" -ldflags="-s -w -X 'dxkite.cn/mino.Version=$VERSION' -X 'dxkite.cn/mino.Commit=$COMMIT'" ./cmd/mino
tar -cvzf mino-$VERSION-$OS.tar.gz mino-$VERSION-$OS