##############################
# Base dev image             #
##############################

FROM golang:1.13.6-buster as development

ENV GO111MODULE=on

RUN apt-get update && \
    # Install Proj - C library for coordinate system conversion and its requirements 
    apt-get install -y libproj13 libproj-dev \
    # Graphviz is needed for pprof
    graphviz 

RUN go get github.com/cortesi/modd/cmd/modd && \
    go get github.com/go-bindata/go-bindata/...

# Unpack libproj shared library files to be copied to distroless debian image
RUN mkdir -p /temp/libproj && cp $(dpkg --listfiles libproj13 | grep .so) /temp/libproj

WORKDIR /workspace

EXPOSE 7080

CMD ["modd"]

################################
# Builder for production image #
################################

FROM development as builder

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go-bindata -o storage/migrations/sqlite/sqlite_migrations.go -pkg sqlite_migrations -prefix "storage/migrations/sqlite/" storage/migrations/sqlite/ \
    && go-bindata -o storage/migrations/postgres/postgres_migrations.go -pkg postgres_migrations -prefix "storage/migrations/postgres/" storage/migrations/postgres/

RUN GOOS=linux go build -o /go/bin/gorge-server -ldflags="-s -w" github.com/whitewater-guide/gorge/server \
    && GOOS=linux go build -o /go/bin/gorge-cli -ldflags="-s -w" github.com/whitewater-guide/gorge/cli

################################
# Production image             #
################################
FROM gcr.io/distroless/base-debian10 as production

COPY --from=builder /go/bin/gorge-server /go/bin/gorge-cli /usr/local/bin/
COPY --from=builder /temp/libproj /usr/lib

EXPOSE 7080

ENTRYPOINT ["gorge-server"]