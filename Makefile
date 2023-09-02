EXTLDFLAGS=-Wl,--start-group -lm -pthread -ldl -lstdc++ -lsqlite3 -lproj -Wl,-end-group -static
# This is required becuase go tests run in tested package's directory, and therefore we need to use timezonedb's absolute path
TIMEZONE_DB_DIR=$(CURDIR)/
export TIMEZONE_DB_DIR

download:
	go mod download

tools: download
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %
# timezone lookup tool has binary named "cmd"
	test -f timezone.data || cmd -build -url https://github.com/evansiroky/timezone-boundary-builder/releases/download/2023b/timezones-with-oceans.geojson.zip

# netgo flag is required because we want to use db address `postgres.local`
# See https://pkg.go.dev/net#hdr-Name_Resolution
# Specifically: `
#   When cgo is available, the cgo-based resolver is used instead under a variety of conditions:
#   ... 
#   and when the name being looked up ends in .local or is an mDNS name.
# `
build: tools
build: GOOS=linux 
build:
ifndef CI
	go build -o build/gorge-cli \
		-ldflags="-extldflags '${EXTLDFLAGS}' -s -w -X 'github.com/whitewater-guide/gorge/version.Version=$(VERSION)'" \
		-tags sqlite_omit_load_extension,netgo \
		github.com/whitewater-guide/gorge/cli
endif
	go build -o build/gorge-server \
		-ldflags="-extldflags '${EXTLDFLAGS}' -s -w -X 'github.com/whitewater-guide/gorge/version.Version=$(VERSION)'" \
		-tags sqlite_omit_load_extension,netgo \
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
