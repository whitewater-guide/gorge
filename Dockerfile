##############################
# Image to build libproj     #
##############################
FROM golang:1.17.2-bullseye as proj

ARG DESTDIR="/build"
ARG PROJ_VERSION="7.2"

RUN git clone --depth 1 --branch ${PROJ_VERSION} https://github.com/OSGeo/PROJ.git

# Setup build env
RUN apt-get update -y && \
    apt-get install -y --fix-missing --no-install-recommends \
    software-properties-common build-essential ca-certificates \
    make cmake wget unzip libtool automake \
    # libtiff5-dev libcurl4-gnutls-dev\
    zlib1g-dev pkg-config libsqlite3-dev sqlite3

# https://github.com/OSGeo/PROJ/blob/7.2/docs/source/install.rst
# Build libproj without curl and tiff
# --disable-dependency-tracking speeds up one-time build
RUN cd PROJ \
    && ./autogen.sh \
    && ./configure --prefix=/usr --disable-dependency-tracking --disable-tiff --without-curl \
    && make -j$(nproc) \
    && make install

##############################
# Base dev image             #
##############################

FROM golang:1.17.2-bullseye as development

ENV GO111MODULE=on

RUN apt-get update && \
    apt-get install -y \
    # sqlite3 is required for proj
    libsqlite3-dev sqlite3 \
    # Graphviz is needed for pprof
    graphviz \
    # unzip is needed for timezones
    unzip

COPY --from=proj  /build/usr/share/proj/ /usr/local/share/proj/
COPY --from=proj  /build/usr/include/ /usr/local/include/
COPY --from=proj  /build/usr/bin/ /usr/local/bin/
COPY --from=proj  /build/usr/lib/ /usr/local/lib/

# Tell linker to look into /usr/local as well
# https://lonesysadmin.net/2013/02/22/error-while-loading-shared-libraries-cannot-open-shared-object-file/
RUN echo "/usr/local/lib\n" >> /etc/ld.so.conf && \
    cat /etc/ld.so.conf && \
    ldconfig

WORKDIR /workspace

################################
# Test/lint production image   #
################################

FROM development as tester

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Although it's set in Makefile, still set it here
# This way we can use vscode debug, which doesn't use make
ENV TIMEZONE_DB_DIR="/workspace/"

RUN make test && \
    make lint && \
    make typescript

################################
# Builder for production image #
################################

FROM tester as builder

ARG VERSION=0.0.0

RUN make build

################################
# Production image             #
################################
FROM gcr.io/distroless/base-debian11 as production

COPY --from=builder /go/bin/gorge-server /go/bin/gorge-cli /usr/local/bin/

# Copy generated timezonedb
ENV TIMEZONE_DB_DIR="/usr/local/share/timezonedb/"
COPY --from=builder /workspace/timezone.msgpack.snap.db /usr/local/share/timezonedb/

# Copy data for libproj
ENV PROJ_LIB=/usr/local/share/proj/
COPY --from=builder /usr/local/share/proj /usr/local/share/proj/

EXPOSE 7080

ENTRYPOINT ["gorge-server"]