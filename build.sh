#!/bin/bash

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