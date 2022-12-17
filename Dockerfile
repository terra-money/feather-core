# docker buildx build --platform linux/amd64 ./ -t feath/feather-based:latest
# docker run --rm -it feath/feather-based /bin/sh

###############################################################################

ARG BUILDPLATFORM=linux/amd64
ARG LWVM_VERSION=v1.1.1

FROM --platform=${BUILDPLATFORM} golang:1.18-alpine3.15 AS base

###############################################################################

FROM base as go-builder

ARG BUILDPLATFORM
ARG LWVM_VERSION

# NOTE: add libusb-dev to run with LEDGER_ENABLED=true
RUN set -eux &&\
    apk add --no-cache \
    ca-certificates \
    build-base \
    musl-dev \
    cmake \
    git

# install mimalloc for musl
RUN set -eux &&\
    git clone --depth 1 https://github.com/microsoft/mimalloc src/mimalloc &&\
    mkdir -p src/mimalloc/build &&\
    cd src/mimalloc/build &&\
    cmake .. &&\
    make -j$(nproc) &&\
    make install


# See https://github.com/CosmWasm/wasmvm/releases
RUN set -eux; \
    LWVM_DOWNLOADS="https://github.com/CosmWasm/wasmvm/releases/download/${LWVM_VERSION}"; \
    wget ${LWVM_DOWNLOADS}/checksums.txt -O /tmp/checksums.txt; \
    if [ ${BUILDPLATFORM} = "linux/amd64" ]; then \
        LWVM_URL="${LWVM_DOWNLOADS}/libwasmvm_muslc.x86_64.a"; \
    elif [ ${BUILDPLATFORM} = "darwin/arm64" ]; then \
        LWVM_URL="${LWVM_DOWNLOADS}/libwasmvm_muslc.aarch64.a"; \
    else \
        echo "Unsupported Build Platfrom ${BUILDPLATFORM}"; \
        exit 1; \
    fi; \
    wget ${LWVM_URL} -O /lib/libwasmvm_muslc.a; \
    CHECKSUM=`sha256sum /lib/libwasmvm_muslc.a | cut -d" " -f1`; \
    grep ${CHECKSUM} /tmp/checksums.txt; \
    rm /tmp/checksums.txt

# make env vars, use static lib (from above) not standard libgo_cosmwasm.so file
ENV CGO_ENABLED=0 \
    LEDGER_ENABLED=false \
    BUILD_FLAGS="-mod=readonly  -tags muslc,linux" \
    LD_FLAGS="-ldflags='-extldflags -L/go/src/mimalloc/build -lmimalloc -Wl,-z,muldefs -static'" \
    MIMALLOC_RESERVE_HUGE_OS_PAGES=4

# Build app
COPY . /app/src
WORKDIR /app/src 
RUN go install ./...
RUN make install

# ###############################################################################

# FROM base as feather-base

# COPY --from=go-builder ${GOPATH}/bin/feather-based /usr/local/bin/feather-based

# RUN addgroup feather-base \
#     && adduser -G feather-base -D -h /feather-base feather-base

# WORKDIR /feather-base

# CMD ["/usr/local/bin/feather-based", "--home", "/app", "start"]