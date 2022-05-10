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

function build_android() {
  git clone https://github.com/TTHHR/mino-android
  cp ./mino-$VERSION-android-arm64 ./mino-android/app/src/main/jniLibs/arm64-v8a/libmino.so
  cd mino-android
  chmod +x gradlew
  ./gradlew assembleDebug
  cp ./app/build/outputs/apk/debug/app-debug.apk ../mino-$VERSION-arm64-debug.apk
}

build "linux" "amd64"
build "linux" "386"
build "darwin" "amd64"

build "android" "arm64"
build_android