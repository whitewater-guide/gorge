##############################
# Base dev image             #
##############################

FROM golang:1.15.0-buster as development

ENV GO111MODULE=on

RUN apt-get update && \
    # Install Proj - C library for coordinate system conversion and its requirements 
    apt-get install -y libproj-dev \
    # Graphviz is needed for pprof
    graphviz 

# Symlink this, so it's available under same path both here and on Mac when installed via brew
RUN ln -s /usr/lib/x86_64-linux-gnu/libproj.a /usr/local/lib/libproj.a

WORKDIR /workspace

################################
# Test/lint production   image #
################################

FROM development as tester

COPY go.mod go.sum ./

RUN go mod download

COPY . .

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
FROM gcr.io/distroless/base-debian10 as production

COPY --from=builder /go/bin/gorge-server /go/bin/gorge-cli /usr/local/bin/

EXPOSE 7080

ENTRYPOINT ["gorge-server"]