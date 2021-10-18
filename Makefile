######################################################
# ⭣ Commands below run inside docker container      #
######################################################
download:
	go mod download

tools: download
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

build: GOOS=linux 
build:
	go build -o /go/bin/gorge-server -ldflags="-s -w -X 'github.com/whitewater-guide/gorge/version.Version=$(VERSION)'" github.com/whitewater-guide/gorge/server
	go build -o /go/bin/gorge-cli -ldflags="-s -w -X 'github.com/whitewater-guide/gorge/version.Version=$(VERSION)'" github.com/whitewater-guide/gorge/cli

# Downloads and builds timezones polygons database
timezonedb: tools
	curl -L "https://github.com/evansiroky/timezone-boundary-builder/releases/download/2020d/timezones.geojson.zip" -o ./timezonedb/timezones.geojson.zip
	unzip -o ./timezonedb/timezones.geojson.zip -d ./timezonedb
	rm ./timezonedb/timezones.geojson.zip
	go run ./timezonedb -json "./timezonedb/combined.json" -db=timezone -type=boltdb

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