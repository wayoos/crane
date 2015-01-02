#!/bin/bash

VERSION=$( cat src/crane/crane.go | grep Version )
VERSION=${VERSION#*\"}
VERSION=${VERSION%\"*}

function d_build () {

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
}

function d_install() {
    if [[ ! -f bin/github-release ]]; then
        echo "Install github-release"
        mkdir -p tmp
        curl -# -L https://github.com/aktau/github-release/releases/download/v0.5.3/darwin-amd64-github-release.tar.bz2 > tmp/github-release.tar.bz2
        tar xvjf tmp/github-release.tar.bz2
        mv bin/darwin/amd64/github-release bin
        rm -rf bin/darwin
        rm -rf tmp
    fi
}

function d_release() {
    echo "Create release"

    d_build

    git tag $VERSION
    git push --tags

    bin/github-release release \
        --user wayoos \
        --repo crane \
        --tag $VERSION \
        --name "Crane v${VERSION}" \
        --description "Crane release v${VERSION}" \

    bin/github-release upload \
        --user wayoos \
        --repo crane \
        --tag $VERSION \
        --name "crane-Darwin-x86_64" \
        --file dist/crane-Darwin-x86_64

    bin/github-release upload \
        --user wayoos \
        --repo crane \
        --tag $VERSION \
        --name "crane-Linux-x86_64" \
        --file dist/crane-Linux-x86_64

}

case "$1" in
    build)
        d_build
        ;;
    install)
        d_install
        ;;
    release)
        d_release
        ;;
    help)
        echo "Usage: build.sh {build|help|release}"
        ;;
    *)
        d_build
        ;;
esac