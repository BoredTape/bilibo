#bin/bash

appName="bilibo"

FetchWeb() {
  cd web
  rm -rf dist
  curl -L https://github.com/BoredTape/bilibo-web/releases/latest/download/dist.tar.gz -o dist.tar.gz
  tar -zxvf dist.tar.gz
  rm -rf dist.tar.gz
  cd ..
}

BuildRelease() {
  rm -rf .git/
  rm -rf "build"
  mkdir -p "build"
  xgo -out "build/$appName" -tags=jsoniter .
  # why? Because some target platforms seem to have issues with upx compression
  upx -9 ./build/$appName-linux-amd64
  cp ./build/$appName-windows-amd64.exe ./build/$appName-windows-amd64-upx.exe
  upx -9 ./build/$appName-windows-amd64-upx.exe
}

MakeRelease() {
  cd build
  mkdir compress
  for i in $(find . -type f -name "bilibo-linux-*"); do
    cp "$i" bilibo
    tar -czvf compress/"$i".tar.gz bilibo
    rm -f bilibo
  done
  for i in $(find . -type f -name "bilibo-darwin-*"); do
    cp "$i" bilibo
    tar -czvf compress/"$i".tar.gz bilibo
    rm -f bilibo
  done
  for i in $(find . -type f -name "bilibo-windows-*"); do
    cp "$i" bilibo.exe
    zip compress/$(echo $i | sed 's/\.[^.]*$//').zip bilibo.exe
    rm -f bilibo.exe
  done
  cd compress
  find . -type f -print0 | xargs -0 md5sum >"$1"
  cat "$1"
  cd ../..
}

BuildDocker () {
  export GOOS=linux
  export GOARCH=amd64
  export CGO_ENABLED=1
  export CC=musl-gcc
  go build -ldflags '-s -w --extldflags "-static -fpic"' -o ./bin/bilibo_linux_amd64 -tags=jsoniter .
  export GOOS=linux
  export GOARCH=arm64
  export CGO_ENABLED=1
  export CC=~/aarch64-linux-musl-cross/bin/aarch64-linux-musl-gcc
  go build -ldflags '-s -w --extldflags "-static -fpic"' -o ./bin/bilibo_linux_arm64 -tags=jsoniter .
}

if [ "$1" = "release" ]; then
  FetchWeb
  BuildRelease
  MakeRelease "md5.txt"
elif [ "$1" = "docker" ]; then
  FetchWeb
  BuildDocker
fi