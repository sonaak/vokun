FROM golang:1.7-alpine

ARG VERSION
LABEL maintainer="evilwire"
LABEL version=$VERSION
LABEL description=""

# setup the default workdir env var
ENV WORKDIR $GOPATH/src/github.com/sonaak/vokun
ENV VERSION $VERSION

# get bash, git, openssh, and gvt
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh
RUN go get -u github.com/FiloSottile/gvt


# setup the work directory
WORKDIR $WORKDIR

# setup the dependents
COPY ./vendor/manifest $WORKDIR/vendor/manifest
RUN cd $WORKDIR && gvt restore

# Copy the code to the workdir
COPY . $WORKDIR

# copy the entrypoint script
COPY build/entrypoint.sh $WORKDIR/entrypoint.sh

# Expose the volume /go/out
VOLUME /go/out


ENTRYPOINT ["/bin/sh", "entrypoint.sh"]