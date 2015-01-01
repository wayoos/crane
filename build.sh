#!/bin/bash

VERSION=$( cat src/crane/crane.go | grep Version )
VERSION=${VERSION#*\"}
VERSION=${VERSION%\"*}

echo "Build crane version $VERSION"

# I commit this source into crane repo. I'm not happy with this but what's append if github repo are closed
#go get github.com/codegangsta/cli
#go get github.com/go-martini/martini
#go get github.com/martini-contrib/render
#go get gopkg.in/yaml.v2
#go get github.com/jmcvetta/napping

rm -rf pkg/darwin_amd64/wayoos.com
rm -rf pkg/linux_amd64/wayoos.com

go install crane

export GOARCH="amd64"
export GOOS="linux"

go install crane

unset GOARCH
unset GOOS

rm -rf dist
mkdir -p dist

cp bin/crane dist/crane-Darwin-x86_64
cp bin/linux_amd64/crane dist/crane-Linux-x86_64
