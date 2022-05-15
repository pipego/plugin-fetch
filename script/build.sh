#!/bin/bash

ldflags="-s -w"

go env -w GOPROXY=https://goproxy.cn,direct

# Plugin
filepath="plugin"
for item in "$filepath"/*.go; do
  buf=${item%.go}
  name=${buf##*/}
  # go tool dist list
  CGO_ENABLED=0 GOARCH=$(go env GOARCH) GOOS=$(go env GOOS) go build -ldflags "$ldflags" -o plugin/fetch-"$name" plugin/"$name".go
  upx plugin/fetch-"$name"
done

# Test
target="plugin-fetch-test"
# go tool dist list
CGO_ENABLED=0 GOARCH=$(go env GOARCH) GOOS=$(go env GOOS) go build -ldflags "$ldflags" -o $target main.go
upx $target
