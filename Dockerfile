FROM golang:1.19.4 AS builder

COPY takeoff /takeoff

WORKDIR /takeoff

RUN make

FROM golang:1.19.4 AS deps

# Why this step? That's because the "takeoff" image technically could be extremely lightweight, should not depend
# on "curl" or anything else

WORKDIR /runc
RUN curl -o runc -L https://github.com/opencontainers/runc/releases/download/v1.1.4/runc.amd64 && chmod +x ./runc

WORKDIR /falco
RUN curl -o falco.tar.gz -L https://download.falco.org/packages/bin/x86_64/falco-0.36.0-x86_64.tar.gz

FROM ubuntu:22.04 AS takeoff

ENV HOST_ROOT=/host

RUN apt-get update && apt-get install -y ca-certificates

COPY --from=deps /falco/falco.tar.gz /falco/falco.tar.gz

WORKDIR /
RUN tar xvf /falco/falco.tar.gz  --strip-components=1 && rm -fr /falco

# smoke test
RUN falco --version

COPY --from=builder /takeoff/takeoff /usr/bin/takeoff
COPY --from=deps /runc/runc /usr/bin/runc
