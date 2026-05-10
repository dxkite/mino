#!/bin/bash

ANDROID_SDK_PATH="/home/harry/android-toolchain-linux/android"
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
  if [ -z "$ANDROID_SDK_PATH" ] || [ ! -d "$ANDROID_SDK_PATH" ]; then
    echo "ANDROID_SDK_PATH is empty or does not exist. Please set ANDROID_SDK_PATH to a valid Android SDK path."
    exit 1
  fi
  git clone https://github.com/TTHHR/mino-android
  cd mino-android
  echo "sdk.dir=$ANDROID_SDK_PATH" > local.properties
  chmod +x gradlew
  ./gradlew assembleDebug
  cp ./app/build/outputs/apk/debug/app-debug.apk ../mino-$VERSION-arm64-debug.apk
}

build "linux" "amd64"
build "linux" "386"
build "darwin" "amd64"

build "android" "arm64"
build_android