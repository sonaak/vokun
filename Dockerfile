FROM alpine:3.5

# the version of the docker file
ARG VERSION
ENV VERSION $VERSION

LABEL maintainer="evilwire"
LABEL version=$VERSION

# Copy the binary
COPY build/$VERSION/vokun /opt/sonaak/vokun
RUN chmod +x /opt/sonaak/vokun

# Expose the appropriate ports
# 9000 is the actual port to respond to requests
# 9090 is the healthcheck port
EXPOSE 9000 9090

# Make the path /opt/sonaak/vokun-api/ available
RUN mkdir -p /opt/sonaak/vokun-api

# Set the entry point
ENTRYPOINT ["/opt/sonaak/vokun"]