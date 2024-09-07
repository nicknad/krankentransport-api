# syntax=docker/dockerfile:1

################################################################################

# Create a stage for building the application.
ARG GO_VERSION=1.23.0
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build

RUN apt-get update && apt-get install -y gcc libc6-dev

WORKDIR /src

# Download dependencies as a separate step to take advantage of Docker's caching.
# Leverage a cache mount to /go/pkg/mod/ to speed up subsequent builds.
# Leverage bind mounts to go.sum and go.mod to avoid having to copy them into
# the container.
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

# This is the architecture you're building for, which is passed in by the builder.
# Placing it here allows the previous steps to be cached across architectures.
ARG TARGETARCH

# Build the application.
# Leverage a cache mount to /go/pkg/mod/ to speed up subsequent builds.
# Leverage a bind mount to the current directory to avoid having to copy the
# source code into the container.
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    CGO_ENABLED=1 GOARCH=$TARGETARCH go build -o /bin/server .

################################################################################
# Create a new stage for running the application that contains the minimal
# runtime dependencies for the application. This often uses a different base
# image from the build stage where the necessary files are copied from the build
# stage.


FROM debian:stable AS final

# Install any runtime dependencies that are needed to run your application.
# Leverage a cache mount to /var/cache/apk/ to speed up subsequent builds.

RUN apt-get update && apt-get install -y libsqlite3-0

RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata && \
    update-ca-certificates

RUN mkdir app

# Copy the executable from the "build" stage.
COPY --from=build /bin/server /app
COPY *.db /app
COPY .env /

# Expose the port that the application listens on.
EXPOSE 2209

# What the container should run when it is started.
ENTRYPOINT [ "/app/server" ]