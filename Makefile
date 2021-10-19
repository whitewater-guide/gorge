######################################################
# ⭣ Commands below run inside docker container      #
######################################################
EXTLDFLAGS=-Wl,--start-group -lm -pthread -ldl -lstdc++ -lsqlite3 -lproj -Wl,-end-group -static
# This is required becuase go tests run in tested package's directory, and therefore we need to use timezonedb's absolute path
TIMEZONE_DB_DIR=$(CURDIR)/
export TIMEZONE_DB_DIR

download:
	go mod download

# timezone lookup tool has binary named "cmd" so we crate symlink
tools: download
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %
	ln -sf ${GOPATH}/bin/cmd ${GOPATH}/bin/timezone

build: GOOS=linux 
build:
	go build -o /go/bin/gorge-cli \
		-ldflags="-extldflags '${EXTLDFLAGS}' -s -w -X 'github.com/whitewater-guide/gorge/version.Version=$(VERSION)'" \
		-tags sqlite_omit_load_extension \
		github.com/whitewater-guide/gorge/cli
	go build -o /go/bin/gorge-server \
		-ldflags="-extldflags '${EXTLDFLAGS}' -s -w -X 'github.com/whitewater-guide/gorge/version.Version=$(VERSION)'" \
		-tags sqlite_omit_load_extension \
		github.com/whitewater-guide/gorge/server

# Downloads and builds timezones polygons database
timezonedb: tools
	test -f timezone.msgpack.snap.db || timezone

test: timezonedb
	go test -count=1 ./... 

test-nodocker: timezonedb 
	go test -count=1 -v -tags=nodocker ./...

lint: tools
	golangci-lint run

typescript: tools
	go run ./typescriptify

# Installs custom certificate from mitmproxy for development purposes
mitmcerts:
	openssl x509 -inform PEM -in /usr/share/ca-certificates/mitmproxy/mitmproxy-ca-cert.cer -out /usr/local/share/ca-certificates/mitmproxy-ca-cert.crt
	update-ca-certificates

run: tools mitmcerts timezonedb
	modd

######################################################
# ⭣ Commands below run on host machine   		    #
######################################################

compose:
	touch .env.development
	docker-compose up

######################################################
# ⭣ Commands below run in CI					    #
######################################################
verify:
	docker build --target tester -t gorge_tester .
	# Extract index.d.ts from image
	docker run --rm --entrypoint "/bin/sh" \
           -v $(shell pwd):/extract gorge_tester -c "cp index.d.ts /extract"
prepare:
	docker build --build-arg VERSION=$(VERSION) -t ghcr.io/whitewater-guide/gorge:latest .
	docker tag ghcr.io/whitewater-guide/gorge:latest ghcr.io/whitewater-guide/gorge:$(VERSION)
publish:
	docker login ghcr.io -u $(GITHUB_USER) -p $(GITHUB_TOKEN)
	docker push ghcr.io/whitewater-guide/gorge:latest
	docker push ghcr.io/whitewater-guide/gorge:$(VERSION)