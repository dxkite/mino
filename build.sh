#!/bin/bash
VERSION=$(git describe --tags)
COMMIT=$(git rev-parse --short HEAD)

function build() {
  OS=$1
  ARCH=$2
  NAME="mino-$VERSION-$OS-$ARCH"
  LD_FLAG="-s -w -X 'dxkite.cn/mino.Version=$VERSION' -X 'dxkite.cn/mino.Commit=$COMMIT'"
  if [[ $OS == windows* ]]; then
      NAME="$NAME.exe"
      LD_FLAG="-H windowsgui $LD_FLAG"
  fi
  echo "build $NAME@$COMMIT for $OS"
  GOOS=$OS GOARCH=$ARCH go build -o "$NAME" -ldflags="$LD_FLAG" ./cmd/mino
  tar -cvzf $NAME.tar.gz $NAME
  echo "build $NAME success"
}

build "windows" "amd64"
build "windows" "386"
build "linux" "amd64"
build "linux" "386"
build "darwin" "amd64"
build "android" "arm64"