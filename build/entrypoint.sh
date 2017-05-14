#! /bin/sh

# build the binary with a certain version
build() {
    echo "--> Building vokun (version ${VERSION})"
    if [[ ! -d ${VERSION} ]]; then
        mkdir -p /go/out/${VERSION}
    fi

    go install && cp /go/bin/vokun /go/out/${VERSION}/vokun
}

build $1