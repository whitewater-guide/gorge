EXTLDFLAGS=-Wl,--start-group -lm -pthread -ldl -lstdc++ -lsqlite3 -lproj -Wl,-end-group -static
# This is required becuase go tests run in tested package's directory, and therefore we need to use timezonedb's absolute path
TIMEZONE_DB_DIR=$(CURDIR)/
export TIMEZONE_DB_DIR

download:
	go mod download

tools: download
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %
# timezone lookup tool has binary named "cmd"
	test -f timezone.msgpack.snap.db || cmd

build: tools
build: GOOS=linux 
build:
	go build -o build/gorge-cli \
		-ldflags="-extldflags '${EXTLDFLAGS}' -s -w -X 'github.com/whitewater-guide/gorge/version.Version=$(VERSION)'" \
		-tags sqlite_omit_load_extension \
		github.com/whitewater-guide/gorge/cli
	go build -o build/gorge-server \
		-ldflags="-extldflags '${EXTLDFLAGS}' -s -w -X 'github.com/whitewater-guide/gorge/version.Version=$(VERSION)'" \
		-tags sqlite_omit_load_extension \
		github.com/whitewater-guide/gorge/server

test: tools
	go test -count=1 ./... 

test-nodocker: tools
	go test -count=1 -v -tags=nodocker ./...

lint: tools
	golangci-lint run

typescript: tools
	go run ./typescriptify

# Installs custom certificate from mitmproxy for development purposes
mitmcerts:
	openssl x509 -inform PEM -in /usr/share/ca-certificates/mitmproxy/mitmproxy-ca-cert.cer -out /usr/local/share/ca-certificates/mitmproxy-ca-cert.crt
	update-ca-certificates

run: tools mitmcerts
	modd

release: build typescript
