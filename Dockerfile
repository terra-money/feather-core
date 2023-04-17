# syntax=docker/dockerfile:1

FROM golang:1.20-bullseye AS base

###############################################################################

FROM base AS dependency-builder

ARG CHAIN_REPO

# NOTE: add libusb-dev to run with LEDGER_ENABLED=true
# set -e: exit immediately if a command exits with a non-zero status
# set -u: treat unset variables as an error when expanding variables
# set -x: print commands and their arguments as they are executed
RUN set -eux && \
    apt-get update -y && \
    # don't cache downloaded packages to keep container size small
    apt-get install -y \ 
    # for 'make install'
    cmake \
    # for 'git clone'
    git && \
    # clean up cache to keep container size small
    apt-get clean

RUN set -eux && \
    # clone the latest revision of the repo
    git clone --depth 1 ${CHAIN_REPO} /app

###############################################################################

FROM dependency-builder AS chain-builder

WORKDIR /app
RUN set -eux && \
    # build the chain using its makefile
    make install

CMD ["tail", "-f", "/dev/null"]
