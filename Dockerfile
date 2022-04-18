# NOTE docker images do not work with systemd because it cannot be statically
# linked.
FROM golang:1.18.1-bullseye as builder

RUN set -ex \
 && apt-get update && apt-get install -y \
   git \
   libsystemd-dev \
 && rm -rf /var/lib/apt/lists/*

# Add dependencies into mod cache
COPY go.mod go.sum /src/
WORKDIR /src

RUN set -ex \
 && go mod download

# Add the application itself and build it
COPY ./ /src/

ARG VERSION

RUN set -ex \
 && go build \
      -ldflags "-X main.GitDescribe=$(git describe --always --tags --dirty) -extldflags '-static'" \
      -mod=readonly \
      -o taily \
      ./cmd

RUN ls -l

FROM scratch

COPY --from=builder /src/taily /usr/local/bin/taily

STOPSIGNAL SIGINT

ENTRYPOINT ["/usr/local/bin/taily"]
